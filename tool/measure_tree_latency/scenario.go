package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
)

type timing struct {
	initialTiming
	Preview       int64 `json:"preview"`
	Apply         int64 `json:"apply"`
	Workbench     int64 `json:"workbench"`
	Report        int64 `json:"report"`
	Geometry      int64 `json:"geometry"`
	ApplyToReport int64 `json:"apply_to_report"`
	ApplyToTree   int64 `json:"apply_to_tree_state"`
}

type initialTiming struct {
	OpenSession      int64 `json:"open_session"`
	InitialWorkbench int64 `json:"initial_workbench"`
	InitialReport    int64 `json:"initial_report"`
	InitialGeometry  int64 `json:"initial_geometry"`
	OpenToReport     int64 `json:"open_to_report"`
	OpenToTree       int64 `json:"open_to_tree_state"`
}

type evidence struct {
	BeforeSHA       string             `json:"before_sha256"`
	AfterSHA        string             `json:"after_sha256"`
	PreviewRevision uint64             `json:"preview_revision"`
	AppliedRevision uint64             `json:"applied_revision"`
	Added           []int              `json:"added_nodes"`
	Before          map[string]float64 `json:"before_stats"`
	After           map[string]float64 `json:"after_stats"`
}

type sample struct {
	Timing   timing
	Evidence evidence
}

type scenario struct {
	session   *grpcapi.Client
	workbench *workbenchClient
	geometry  bool
	node      int
}

func measureRound(ctx context.Context, opt options, index, round int, xml string) ([]sample, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Minute)
	defer cancel()
	scratch, err := os.MkdirTemp("", "pob-tree-latency-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(scratch)
	target := opt.Targets[index]
	var ready grpcapi.Ready
	server, err := startChild(ctx, target, scratch, &ready, "serve", "--profile", "measure-tree", "--runtime-dir", scratch, "--data-dir", scratch)
	if err != nil {
		return nil, err
	}
	defer server.close()
	client, err := grpcapi.Dial(ctx, ready.Address, ready.Token)
	if err != nil {
		return nil, err
	}
	defer stopSession(client)
	scenario := scenario{session: client, geometry: opt.Geometry, node: opt.Node}
	if opt.Workbench {
		var ready grpcapi.WorkbenchReady
		worker, err := startChild(ctx, target, scratch, &ready, "workbench", "--grpc")
		if err != nil {
			return nil, err
		}
		defer worker.close()
		scenario.workbench, err = openWorkbench(ready)
		if err != nil {
			return nil, err
		}
		defer scenario.workbench.connection.Close()
	}
	count := 1 + opt.Warm/opt.Cold
	if round < opt.Warm%opt.Cold {
		count++
	}
	return scenario.measure(ctx, count, xml)
}

func (s scenario) measure(ctx context.Context, count int, xml string) ([]sample, error) {
	results := make([]sample, count)
	for i := range results {
		entry, err := s.measureSample(ctx, fmt.Sprintf("build-%d", i), xml)
		if err != nil {
			return nil, fmt.Errorf("замер %d: %w", i+1, err)
		}
		results[i] = entry
	}
	return results, nil
}

func (s scenario) measureSample(ctx context.Context, id, xml string) (sample, error) {
	before, initial, err := s.openBuild(ctx, id, xml)
	if err != nil {
		return sample{}, err
	}
	edit := &build.Edit{TreeNode: &build.TreeNodeEdit{NodeID: s.node, Action: "allocate", AllocationMode: "common"}}
	started := time.Now()
	preview, err := s.session.Compare(ctx, application.CompareBuild{SessionID: id, ExpectedRevision: before.Revision, Edit: edit})
	previewNS := time.Since(started).Nanoseconds()
	if err != nil {
		return sample{}, fmt.Errorf("предпросмотр: %w", err)
	}
	if err := s.checkUnchanged(ctx, before); err != nil {
		return sample{}, err
	}
	started = time.Now()
	after, err := s.session.Apply(ctx, application.ApplyBuild{SessionID: id, ExpectedRevision: before.Revision, RequestID: "allocate", Edit: edit})
	applyNS := time.Since(started).Nanoseconds()
	if err != nil {
		return sample{}, fmt.Errorf("применение: %w", err)
	}
	stages, stats, err := s.measureRefresh(ctx, after)
	if err != nil {
		return sample{}, err
	}
	verified, err := verifyResult(s.node, before, preview, after, stats)
	if err != nil {
		return sample{}, err
	}
	stages.Preview = previewNS
	stages.Apply = applyNS
	stages.ApplyToReport = applyNS + stages.Workbench + stages.Report
	stages.ApplyToTree = stages.ApplyToReport + stages.Geometry
	stages.initialTiming = initial
	return sample{Timing: stages, Evidence: verified}, nil
}

func (s scenario) checkUnchanged(ctx context.Context, before domain.Snapshot) error {
	current, err := s.session.Snapshot(ctx, before.SessionID)
	if err != nil {
		return fmt.Errorf("проверка снимка после предпросмотра: %w", err)
	}
	if current != before {
		return fmt.Errorf("предпросмотр изменил исходный снимок")
	}
	return nil
}

func (s scenario) openBuild(ctx context.Context, id, xml string) (domain.Snapshot, initialTiming, error) {
	started := time.Now()
	snapshot, err := s.session.Create(ctx, application.CreateSession{SessionID: id, BuildName: "Warrior", RequestID: "open-" + id, XML: xml})
	openNS := time.Since(started).Nanoseconds()
	if err != nil {
		return domain.Snapshot{}, initialTiming{}, fmt.Errorf("открытие эталона: %w", err)
	}
	refresh, _, err := s.measureRefresh(ctx, snapshot)
	if err != nil {
		return domain.Snapshot{}, initialTiming{}, fmt.Errorf("начальный отчёт: %w", err)
	}
	toReport := openNS + refresh.Workbench + refresh.Report
	initial := initialTiming{
		OpenSession: openNS, InitialWorkbench: refresh.Workbench, InitialReport: refresh.Report, InitialGeometry: refresh.Geometry,
		OpenToReport: toReport, OpenToTree: toReport + refresh.Geometry,
	}
	return snapshot, initial, nil
}

func (s scenario) measureRefresh(ctx context.Context, snapshot domain.Snapshot) (timing, map[string]float64, error) {
	var stages timing
	if s.workbench != nil {
		started := time.Now()
		if err := s.workbench.importXML(ctx, snapshot.BuildXML); err != nil {
			return stages, nil, fmt.Errorf("обновление рабочей копии: %w", err)
		}
		stages.Workbench = time.Since(started).Nanoseconds()
	}
	started := time.Now()
	view, err := s.session.BuildReport(ctx, snapshot.SessionID, snapshot.Revision)
	stages.Report = time.Since(started).Nanoseconds()
	if err != nil {
		return stages, nil, fmt.Errorf("отчёт Lua: %w", err)
	}
	stats, err := readReportStats(view, snapshot.Revision)
	if err != nil {
		return stages, nil, err
	}
	if s.geometry {
		stages.Geometry, err = s.measureGeometry(ctx, snapshot)
	}
	return stages, stats, err
}

func (s scenario) measureGeometry(ctx context.Context, snapshot domain.Snapshot) (int64, error) {
	started := time.Now()
	tree, err := s.session.TreeGeometry(ctx, snapshot.SessionID, snapshot.Revision, true)
	elapsed := time.Since(started).Nanoseconds()
	if err != nil {
		return 0, fmt.Errorf("полная сцена дерева: %w", err)
	}
	if tree.Revision != snapshot.Revision || tree.Scene == nil || len(tree.Nodes) == 0 {
		return 0, fmt.Errorf("полная сцена дерева не соответствует ревизии %d", snapshot.Revision)
	}
	return elapsed, nil
}

var statNames = []string{"Str", "Life", "TotalDPS"}

func readReportStats(view application.BuildReportView, revision uint64) (map[string]float64, error) {
	if view.Revision != revision {
		return nil, fmt.Errorf("отчёт Lua относится к ревизии %d вместо %d", view.Revision, revision)
	}
	var payload struct {
		Stats map[string]json.RawMessage `json:"stats"`
	}
	if err := json.Unmarshal(view.Report, &payload); err != nil {
		return nil, fmt.Errorf("чтение отчёта Lua: %w", err)
	}
	values := make(map[string]float64, len(statNames))
	for _, name := range statNames {
		var value float64
		if err := json.Unmarshal(payload.Stats[name], &value); err != nil || value <= 0 {
			return nil, fmt.Errorf("у отчёта Lua отсутствует положительный показатель %s: %v", name, err)
		}
		values[name] = value
	}
	return values, nil
}

func verifyResult(node int, before domain.Snapshot, preview application.BuildComparison, after domain.Snapshot, stats map[string]float64) (evidence, error) {
	if err := verifyEdit(node, before, preview, after); err != nil {
		return evidence{}, err
	}
	initial := make(map[string]float64, len(stats))
	for _, name := range statNames {
		afterPreview := preview.Comparison.After.Stats[name]
		if math.Abs(stats[name]-afterPreview) > 1e-7*math.Max(1, math.Abs(afterPreview)) {
			return evidence{}, fmt.Errorf("%s в отчёте %g, в предпросмотре %g", name, stats[name], afterPreview)
		}
		initial[name] = preview.Comparison.Before.Stats[name]
	}
	return evidence{
		BeforeSHA: before.InputSHA256, AfterSHA: after.InputSHA256, PreviewRevision: preview.Revision,
		AppliedRevision: after.Revision, Added: preview.Passives.Added, Before: initial, After: stats,
	}, nil
}

func verifyEdit(node int, before domain.Snapshot, preview application.BuildComparison, after domain.Snapshot) error {
	if preview.Revision != before.Revision || after.Revision != before.Revision+1 || after.BuildRevision != after.Revision {
		return fmt.Errorf("неверные ревизии: открытие %d, предпросмотр %d, применение %d", before.Revision, preview.Revision, after.Revision)
	}
	comparison := preview.Comparison
	if comparison.Before.Status != "complete" || comparison.After.Status != "complete" {
		return fmt.Errorf("неполный расчёт: %s, %s, предупреждения %v / %v", comparison.Before.Status, comparison.After.Status, comparison.Before.Issues, comparison.After.Issues)
	}
	if before.InputSHA256 != comparison.Before.InputSHA256 || after.InputSHA256 != comparison.After.InputSHA256 || before.InputSHA256 == after.InputSHA256 {
		return fmt.Errorf("хеши предпросмотра не соответствуют снимкам до и после правки")
	}
	if preview.Passives == nil || !slices.Contains(preview.Passives.Added, node) || len(preview.Passives.Removed) != 0 {
		return fmt.Errorf("неожиданный путь до пассива: %+v", preview.Passives)
	}
	view, err := application.ViewSnapshot(after)
	if err != nil {
		return err
	}
	if view.Build == nil || !slices.Contains(view.Build.Model.Nodes, node) {
		return fmt.Errorf("пассив %d отсутствует в снимке после применения", node)
	}
	return nil
}
