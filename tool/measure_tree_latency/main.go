// Команда сравнивает задержки выбора и назначения пассива через публичный gRPC.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"slices"
	"syscall"
)

type target struct {
	Binary string `json:"binary"`
	PoB    string `json:"pob_directory"`
}

type options struct {
	Targets   [2]target
	Fixture   string
	Node      int
	Cold      int
	Warm      int
	Workbench bool
	Geometry  bool
}

func parseOptions() options {
	var opt options
	flag.StringVar(&opt.Targets[0].Binary, "baseline", "", "бинарник исходного движка")
	flag.StringVar(&opt.Targets[1].Binary, "candidate", "", "бинарник нового движка")
	flag.StringVar(&opt.Targets[0].PoB, "baseline-pob", "", "установка PoB для исходного движка")
	flag.StringVar(&opt.Targets[1].PoB, "candidate-pob", "", "установка PoB для нового движка")
	flag.StringVar(&opt.Fixture, "fixture", "testdata/reference/0.0.10/warrior.xml", "XML эталона сборки")
	flag.IntVar(&opt.Node, "node", 13397, "идентификатор неназначенного пассива")
	flag.IntVar(&opt.Cold, "cold", 3, "число новых процессов каждого движка")
	flag.IntVar(&opt.Warm, "warm", 20, "число повторов каждого движка после первого вызова")
	flag.BoolVar(&opt.Workbench, "workbench", true, "включить обновление XML рабочей копии")
	flag.BoolVar(&opt.Geometry, "geometry", true, "включить получение полной сцены дерева")
	flag.Parse()
	return opt
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, parseOptions()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, opt options) error {
	if flag.NArg() != 0 || opt.Node < 1 || opt.Cold < 1 || opt.Warm < 0 {
		return fmt.Errorf("укажите положительные --node и --cold, неотрицательное --warm и только именованные параметры")
	}
	pob, err := installationHash(&opt)
	if err != nil {
		return err
	}
	xml, err := os.ReadFile(opt.Fixture)
	if err != nil {
		return err
	}
	value, err := measureAB(ctx, opt, string(xml))
	if err != nil {
		return err
	}
	result := report{
		Fixture: opt.Fixture, FixtureSHA: fmt.Sprintf("%x", sha256.Sum256(xml)), PoBManifestSHA: pob,
		Node: opt.Node, Workbench: opt.Workbench, Geometry: opt.Geometry,
		Evidence: value.Evidence, Baseline: value.Results[0], Candidate: value.Results[1],
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func installationHash(opt *options) (string, error) {
	var manifestSHA [2][32]byte
	for i := range opt.Targets {
		item := &opt.Targets[i]
		if item.Binary == "" || item.PoB == "" {
			return "", fmt.Errorf("нужны --baseline, --candidate, --baseline-pob и --candidate-pob")
		}
		binary, err := filepath.Abs(item.Binary)
		if err != nil {
			return "", err
		}
		item.Binary = binary
		manifest, err := os.ReadFile(filepath.Join(item.PoB, "manifest.xml"))
		if err != nil {
			return "", fmt.Errorf("чтение манифеста PoB: %w", err)
		}
		manifestSHA[i] = sha256.Sum256(manifest)
	}
	if manifestSHA[0] != manifestSHA[1] {
		return "", fmt.Errorf("манифесты двух установок PoB различаются")
	}
	return fmt.Sprintf("%x", manifestSHA[0]), nil
}

type comparison struct {
	Results  [2]targetReport
	Evidence evidence
}

func measureAB(ctx context.Context, opt options, xml string) (comparison, error) {
	var value comparison
	for round := range opt.Cold {
		for _, i := range []int{round % 2, 1 - round%2} {
			name := [2]string{"исходный", "новый"}[i]
			fmt.Fprintf(os.Stderr, "Прогон %d/%d, %s движок\n", round+1, opt.Cold, name)
			result, err := measureRound(ctx, opt, i, round, xml)
			if err != nil {
				return value, fmt.Errorf("прогон %d, %s движок: %w", round+1, name, err)
			}
			if err := value.record(i, opt.Targets[i], result); err != nil {
				return value, fmt.Errorf("прогон %d, %s движок: %w", round+1, name, err)
			}
		}
	}
	for i := range value.Results {
		value.Results[i].ColdSummary = summaries(value.Results[i].Cold)
		value.Results[i].WarmSummary = summaries(value.Results[i].Warm)
	}
	return value, nil
}

func (c *comparison) record(index int, target target, result []sample) error {
	report := &c.Results[index]
	report.Target = target
	for i, value := range result {
		if c.Evidence.BeforeSHA == "" {
			c.Evidence = value.Evidence
		}
		if !reflect.DeepEqual(c.Evidence, value.Evidence) {
			return fmt.Errorf("замер %d: ответ расходится с первым прогоном: %+v / %+v", i+1, c.Evidence, value.Evidence)
		}
	}
	report.Cold = append(report.Cold, result[0].Timing)
	for _, value := range result[1:] {
		report.Warm = append(report.Warm, value.Timing)
	}
	return nil
}

type report struct {
	Fixture        string       `json:"fixture"`
	FixtureSHA     string       `json:"fixture_sha256"`
	PoBManifestSHA string       `json:"pob_manifest_sha256"`
	Node           int          `json:"node_id"`
	Workbench      bool         `json:"workbench"`
	Geometry       bool         `json:"geometry"`
	Evidence       evidence     `json:"verified_result"`
	Baseline       targetReport `json:"baseline"`
	Candidate      targetReport `json:"candidate"`
}

type targetReport struct {
	Target      target                 `json:"target"`
	Cold        []timing               `json:"cold_samples_ns"`
	Warm        []timing               `json:"warm_samples_ns"`
	ColdSummary map[string]percentiles `json:"cold_summary_ns"`
	WarmSummary map[string]percentiles `json:"warm_summary_ns"`
}

type percentiles struct {
	P50 int64 `json:"p50"`
	P95 int64 `json:"p95"`
}

func summaries(samples []timing) map[string]percentiles {
	if len(samples) == 0 {
		return nil
	}
	metrics := map[string]func(timing) int64{
		"open_session":        func(t timing) int64 { return t.OpenSession },
		"initial_workbench":   func(t timing) int64 { return t.InitialWorkbench },
		"initial_report":      func(t timing) int64 { return t.InitialReport },
		"initial_geometry":    func(t timing) int64 { return t.InitialGeometry },
		"open_to_report":      func(t timing) int64 { return t.OpenToReport },
		"open_to_tree_state":  func(t timing) int64 { return t.OpenToTree },
		"preview":             func(t timing) int64 { return t.Preview },
		"apply":               func(t timing) int64 { return t.Apply },
		"workbench":           func(t timing) int64 { return t.Workbench },
		"report":              func(t timing) int64 { return t.Report },
		"geometry":            func(t timing) int64 { return t.Geometry },
		"apply_to_report":     func(t timing) int64 { return t.ApplyToReport },
		"apply_to_tree_state": func(t timing) int64 { return t.ApplyToTree },
	}
	result := make(map[string]percentiles, len(metrics))
	for key, selectTime := range metrics {
		values := make([]int64, len(samples))
		for i, value := range samples {
			values[i] = selectTime(value)
		}
		slices.Sort(values)
		result[key] = percentiles{P50: values[(50*len(values)+99)/100-1], P95: values[(95*len(values)+99)/100-1]}
	}
	return result
}
