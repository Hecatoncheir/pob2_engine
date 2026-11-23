package pob

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

type Layout struct {
	Source  string
	Runtime string
	Native  string
}

type Calculator struct {
	installation Installation
	layout       Layout
	gate         chan struct{}
	worker       *worker
	data         *gamedata.Dataset
}

// Парсер XML в PoB распознаёт именованные сущности, а Go кодирует эти символы числами.
var pobXML = strings.NewReplacer("&#34;", "&quot;", "&#39;", "&apos;", "&#x9;", "\t", "&#xA;", "\n", "&#xD;", "\r")

func NewCalculator(installation Installation) (*Calculator, error) {
	layout, err := installation.Layout()
	if err != nil {
		return nil, err
	}
	return &Calculator{installation: installation, layout: layout, gate: make(chan struct{}, 1)}, nil
}

func (c *Calculator) Close() error {
	c.gate <- struct{}{}
	defer func() { <-c.gate }()
	c.stop()
	return nil
}

func (c *Calculator) Calculate(ctx context.Context, doc *build.Document) (calculation.Result, error) {
	xml, err := doc.Export()
	if err != nil {
		return calculation.Result{}, err
	}
	digest, err := doc.InputSHA256()
	if err != nil {
		return calculation.Result{}, err
	}
	answer, err := c.run(ctx, &request{XML: pobXML.Replace(string(xml))})
	if err != nil {
		return calculation.Result{}, err
	}
	return c.result(answer, digest)
}

func (c *Calculator) Compare(ctx context.Context, before, after *build.Document) (calculation.Comparison, error) {
	return calculation.Compare(ctx, c, before, after)
}

// ValidateEquipment проверяет слот и предмет в изменённом документе через Lua PoB.
func (c *Calculator) ValidateEquipment(ctx context.Context, doc *build.Document, slot string) error {
	model, err := doc.Model()
	if err != nil {
		return err
	}
	id, ok := model.Slots[slot]
	if !ok {
		return fmt.Errorf("слот %s отсутствует в выбранном наборе", slot)
	}
	xml, err := doc.Export()
	if err != nil {
		return err
	}
	answer, err := c.run(ctx, &request{
		Operation: "validate_equipment", XML: pobXML.Replace(string(xml)), Slot: slot, ItemID: id,
	})
	if err != nil {
		return err
	}
	if answer.Error != "" {
		return fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	if !answer.EquipmentValid {
		return fmt.Errorf("PoB Lua не подтвердил правку экипировки")
	}
	return nil
}

func (c *Calculator) result(answer response, digest string) (calculation.Result, error) {
	if answer.Error != "" {
		return calculation.Result{}, fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	if len(answer.Stats) == 0 {
		return calculation.Result{}, fmt.Errorf("PoB Lua не вернул показатели")
	}
	r := calculation.Result{Status: "complete", DataID: c.installation.dataID(),
		DataSHA256: c.installation.SHA256, PoBCommit: c.installation.Commit, InputSHA256: digest,
		Skill: answer.Skill, Mode: answer.Mode, Stats: answer.Stats, Issues: []build.Issue{}, Modifiers: []calculation.Modifier{}}
	for _, warning := range answer.Warnings {
		r.Status = "partial"
		r.Issues = append(r.Issues, build.Issue{Code: "pob_warning", Message: warning})
	}
	for _, line := range answer.Unsupported {
		r.Status = "partial"
		r.Issues = append(r.Issues, build.Issue{Code: "unsupported_modifier", Path: "Config/customMods", Message: line})
	}
	if len(answer.Issues) > 0 {
		r.Status = "partial"
		r.Issues = append(r.Issues, answer.Issues...)
	}
	return r, nil
}

func (i Installation) dataID() string {
	if i.SHA256 == "" && i.Commit == "" {
		return calculation.LocalPoBDataID
	}
	return "pob-lua-" + i.Version
}

func (c *Calculator) acquire(ctx context.Context) (func(), error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case c.gate <- struct{}{}:
	}
	return func() { <-c.gate }, nil
}

func (c *Calculator) run(ctx context.Context, value *request) (response, error) {
	release, err := c.acquire(ctx)
	if err != nil {
		return response{}, err
	}
	defer release()
	return c.execute(ctx, value)
}

func (c *Calculator) execute(ctx context.Context, value *request) (response, error) {
	if err := ctx.Err(); err != nil {
		return response{}, err
	}
	if c.worker == nil {
		if err := c.start(ctx); err != nil {
			return response{}, err
		}
	}
	return c.exchange(ctx, value)
}

// CatalogData читает каталог той же установки, которая выполняет расчёты.
func (c *Calculator) CatalogData(ctx context.Context) (*gamedata.Dataset, error) {
	release, err := c.acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer release()
	if c.data != nil {
		return c.data, ctx.Err()
	}
	answer, err := c.execute(ctx, &request{Operation: "catalog"})
	if err != nil {
		return nil, err
	}
	if answer.Error != "" || answer.Catalog == nil {
		return nil, fmt.Errorf("PoB Lua не вернул каталог: %s", answer.Error)
	}
	c.data, err = gamedata.NewRuntimeCatalog(*answer.Catalog, c.installation.dataID(), c.installation.SHA256, c.installation.Commit)
	return c.data, err
}

func (c *Calculator) start(ctx context.Context) error {
	w, err := startWorker(c.layout)
	if err != nil {
		return err
	}
	c.worker = w
	ready, err := c.exchange(ctx, nil)
	if err != nil {
		return err
	}
	if !ready.Ready {
		c.stop()
		return fmt.Errorf("воркер Lua не подтвердил готовность")
	}
	return nil
}

func (c *Calculator) exchange(ctx context.Context, value *request) (response, error) {
	w := c.worker
	done := make(chan workerReply, 1)
	go func() { done <- w.request(value) }()
	select {
	case <-ctx.Done():
		c.stop()
		<-done
		return response{}, ctx.Err()
	case reply := <-done:
		if reply.err != nil {
			c.stop()
		}
		return reply.answer, reply.err
	}
}

func (c *Calculator) stop() {
	if c.worker != nil {
		c.worker.close()
		c.worker = nil
	}
}

type response struct {
	Ready          bool                     `json:"ready"`
	EquipmentValid bool                     `json:"equipment_valid"`
	Skill          string                   `json:"skill"`
	Mode           string                   `json:"mode"`
	Stats          map[string]float64       `json:"stats"`
	Warnings       []string                 `json:"warnings"`
	Unsupported    []string                 `json:"unsupported"`
	Error          string                   `json:"error"`
	Catalog        *gamedata.RuntimeCatalog `json:"catalog,omitempty"`
	Issues         []build.Issue            `json:"issues,omitempty"`
}

type request struct {
	XML       string `json:"xml,omitempty"`
	Operation string `json:"operation,omitempty"`
	Slot      string `json:"slot,omitempty"`
	ItemID    int    `json:"item_id,omitempty"`
}

type workerReply struct {
	answer response
	err    error
}

type worker struct {
	cmd    *exec.Cmd
	input  io.WriteCloser
	output *bufio.Scanner
	work   string
}

func startWorker(layout Layout) (*worker, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	work, err := os.MkdirTemp("", "pob2-lua-")
	if err != nil {
		return nil, err
	}
	bridgeDir, err := unpackBridge(work)
	if err != nil {
		os.RemoveAll(work)
		return nil, err
	}
	cmd := exec.Command(executable, "_pob-lua", layout.Runtime, layout.Native, bridgeDir)
	cmd.Dir, cmd.Stderr = layout.Source, os.Stderr
	cmd.Env = append(cmd.Environ(), "LUA_PATH=", "LUA_CPATH=", "HOME="+work, "USERPROFILE="+work, "APPDATA="+work,
		"LOCALAPPDATA="+work, "XDG_CONFIG_HOME="+work, "XDG_DATA_HOME="+work)
	w, err := openWorker(cmd)
	if err != nil {
		os.RemoveAll(work)
		return nil, err
	}
	w.work = work
	return w, nil
}

func openWorker(cmd *exec.Cmd) (*worker, error) {
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	in, err := cmd.StdinPipe()
	if err != nil {
		out.Close()
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		in.Close()
		out.Close()
		return nil, err
	}
	scanner := bufio.NewScanner(out)
	scanner.Buffer(nil, 8<<20)
	return &worker{cmd: cmd, input: in, output: scanner}, nil
}

func (w *worker) request(value *request) workerReply {
	if value != nil {
		if err := json.NewEncoder(w.input).Encode(value); err != nil {
			return workerReply{err: fmt.Errorf("запись в PoB Lua: %w", err)}
		}
	}
	if !w.output.Scan() {
		err := w.output.Err()
		if err == nil {
			err = io.EOF
		}
		return workerReply{err: fmt.Errorf("чтение ответа PoB Lua: %w", err)}
	}
	var result response
	err := json.Unmarshal(w.output.Bytes(), &result)
	if err != nil {
		return workerReply{err: fmt.Errorf("неверный ответ PoB Lua: %w", err)}
	}
	return workerReply{answer: result}
}

func (w *worker) close() {
	w.input.Close()
	_ = w.cmd.Process.Kill()
	_ = w.cmd.Wait()
	_ = os.RemoveAll(w.work)
}

func (i Installation) Layout() (Layout, error) {
	source, runtimeDir := i.Directory, i.Directory
	if regularFile(filepath.Join(i.Directory, "src", "Launch.lua")) {
		source = filepath.Join(i.Directory, "src")
	}
	if regularFile(filepath.Join(i.Directory, "runtime", "lua", "dkjson.lua")) {
		runtimeDir = filepath.Join(i.Directory, "runtime")
	}
	nativeDir := runtimeDir
	if i.NativeDirectory != "" {
		nativeDir = i.NativeDirectory
	}
	library, extension := luaLibrary(runtimeDir, nativeDir)
	for _, required := range []string{filepath.Join(source, "Launch.lua"), filepath.Join(runtimeDir, "lua", "dkjson.lua"), library, filepath.Join(nativeDir, "lua-utf8."+extension)} {
		if !regularFile(required) {
			return Layout{}, fmt.Errorf("не найден файл PoB: %s", required)
		}
	}
	return Layout{source, runtimeDir, nativeDir}, nil
}

func regularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
