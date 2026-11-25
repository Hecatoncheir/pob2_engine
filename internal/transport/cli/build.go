package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
	"unicode/utf8"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

// Первый вызов на сервере может загружать Lua PoB и данные дерева.
const buildTimeout = 30 * time.Second

type ResultStatusError struct{ Status string }

func (e *ResultStatusError) Error() string {
	if e.Status == "partial" {
		return "расчёт частичный; причины указаны в issues"
	}
	return "расчёт не поддерживается; причины указаны в issues"
}

func (r runner) readInput(ctx context.Context, path string) ([]byte, error) {
	if path == "" {
		return nil, &UsageError{Err: fmt.Errorf("нужен --file; используйте - для stdin")}
	}
	reader := r.in
	if path != "-" {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() || info.Size() > build.MaxXMLSize {
			return nil, &UsageError{Err: fmt.Errorf("нужен обычный файл размером до 1 МиБ")}
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader = file
	}
	stop := context.AfterFunc(ctx, func() { _ = reader.Close() })
	defer stop()
	data, err := io.ReadAll(io.LimitReader(reader, build.MaxXMLSize+1))
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, err
	}
	if len(data) > build.MaxXMLSize || !utf8.Valid(data) {
		return nil, &UsageError{Err: fmt.Errorf("ввод: до 1 МиБ корректного UTF-8")}
	}
	return data, nil
}

func parseEdit(data []byte) (*build.Edit, error) {
	var edit build.Edit
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&edit); err != nil {
		return nil, &UsageError{Err: err}
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, &UsageError{Err: fmt.Errorf("лишние данные после правки JSON")}
	}
	return &edit, nil
}

func (r runner) calculate(ctx context.Context, args []string) error {
	fs := r.flags("calculate")
	connection := connectFlagsWithTimeout(fs, buildTimeout)
	file := fs.String("file", "", "XML-файл либо - для stdin; одноразовый расчёт без сервера")
	id := fs.String("session", "", "сессия запущенного сервера")
	revision := fs.Uint64("revision", 0, "ожидаемая ревизия сессии")
	request := fs.String("request-id", "", "ID сохраняемого расчёта")
	if err := parse(fs, args); err != nil {
		return err
	}
	if (*file == "") == (*id == "") {
		return &UsageError{Err: fmt.Errorf("нужен --file либо --session")}
	}
	if *file != "" {
		if connection.usesServer() || *request != "" || *revision != 0 {
			return &UsageError{Err: fmt.Errorf("расчёт --file не принимает параметры сервера и сессии")}
		}
		return r.calculateFile(ctx, *file, connection)
	}
	return r.calculateSession(ctx, connection, application.CalculateBuild{SessionID: *id, ExpectedRevision: *revision, RequestID: *request})
}

func (r runner) calculateFile(ctx context.Context, file string, c *connection) error {
	data, err := r.readInput(ctx, file)
	if err != nil {
		return err
	}
	doc, err := build.Parse(data)
	if err != nil {
		return &UsageError{Err: err}
	}
	engine, closeCalculator, err := openCalculator(ctx)
	if err != nil {
		return err
	}
	defer closeCalculator()
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	result, err := engine.Calculate(ctx, doc)
	if err != nil {
		return err
	}
	if err := r.write(result); err != nil {
		return err
	}
	return resultStatus(result.Status)
}

func (r runner) calculateSession(ctx context.Context, connection *connection, cmd application.CalculateBuild) error {
	ctx, cancel := context.WithTimeout(ctx, connection.timeout)
	defer cancel()
	client, err := connection.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	snapshot, err := client.Calculate(ctx, cmd)
	if err != nil {
		return err
	}
	if err := r.write(snapshot); err != nil {
		return err
	}
	result, err := application.DecodeCalculation(snapshot)
	if err != nil {
		return err
	}
	return resultStatus(result.Status)
}

func resultStatus(status string) error {
	if status == "partial" || status == "unsupported" {
		return &ResultStatusError{Status: status}
	}
	return nil
}

func (r runner) build(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return &UsageError{Err: fmt.Errorf("нужна операция build import|edit|export|compare")}
	}
	switch args[0] {
	case "import", "edit":
		return r.applyBuild(ctx, args[0], args[1:])
	case "compare":
		return r.compareBuild(ctx, args[1:])
	case "export":
		return r.exportBuild(ctx, args[1:])
	default:
		return &UsageError{Err: fmt.Errorf("неизвестная операция build %q", args[0])}
	}
}

func (r runner) applyBuild(ctx context.Context, operation string, args []string) error {
	fs := r.flags("build " + operation)
	connection := connectFlagsWithTimeout(fs, buildTimeout)
	asEdit := operation == "edit"
	fileDescription := "XML-файл сборки; - для stdin"
	if asEdit {
		fileDescription = "JSON-файл правки; - для stdin"
	}
	file := fs.String("file", "", fileDescription)
	var command application.ApplyBuild
	fs.StringVar(&command.SessionID, "session", "", "явно выбранный session_id")
	fs.Uint64Var(&command.ExpectedRevision, "revision", 0, "ожидаемая ревизия")
	fs.StringVar(&command.RequestID, "request-id", "", "ID правки для безопасного повтора")
	if err := parse(fs, args); err != nil {
		return err
	}
	if command.RequestID == "" {
		return &UsageError{Err: fmt.Errorf("нужен --request-id")}
	}
	if err := validateBuildInput(connection, command.SessionID, command.ExpectedRevision, *file); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, connection.timeout)
	defer cancel()
	xml, edit, err := r.readBuildInput(ctx, *file, asEdit)
	if err != nil {
		return err
	}
	command.XML, command.Edit = xml, edit
	client, err := connection.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	value, err := client.Apply(ctx, command)
	if err != nil {
		return err
	}
	return r.write(value)
}

func validateBuildInput(connection *connection, id string, revision uint64, file string) error {
	switch {
	case id == "":
		return &UsageError{Err: fmt.Errorf("нужна --session")}
	case revision == 0:
		return &UsageError{Err: fmt.Errorf("нужна положительная --revision")}
	case file == "":
		return &UsageError{Err: fmt.Errorf("нужен --file; используйте - для stdin")}
	}
	return connection.validateFlags()
}

func (r runner) readBuildInput(ctx context.Context, file string, asEdit bool) (string, *build.Edit, error) {
	data, err := r.readInput(ctx, file)
	if err != nil {
		return "", nil, err
	}
	if !asEdit {
		return string(data), nil, nil
	}
	edit, err := parseEdit(data)
	return "", edit, err
}

func (r runner) compareBuild(ctx context.Context, args []string) error {
	fs := r.flags("build compare")
	connection := connectFlagsWithTimeout(fs, buildTimeout)
	file := fs.String("file", "", "XML-файл варианта или JSON-файл правки; - для stdin")
	asEdit := fs.Bool("edit", false, "входной файл содержит JSON правки")
	var command application.CompareBuild
	fs.StringVar(&command.SessionID, "session", "", "явно выбранный session_id")
	fs.Uint64Var(&command.ExpectedRevision, "revision", 0, "ожидаемая ревизия")
	if err := parse(fs, args); err != nil {
		return err
	}
	if err := validateBuildInput(connection, command.SessionID, command.ExpectedRevision, *file); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, connection.timeout)
	defer cancel()
	xml, edit, err := r.readBuildInput(ctx, *file, *asEdit)
	if err != nil {
		return err
	}
	command.XML, command.Edit = xml, edit
	client, err := connection.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	value, err := client.Compare(ctx, command)
	if err != nil {
		return err
	}
	if err := r.write(value); err != nil {
		return err
	}
	before, after := value.Comparison.Before.Status, value.Comparison.After.Status
	if before == "unsupported" || after == "unsupported" {
		return resultStatus("unsupported")
	}
	if err := resultStatus(before); err != nil {
		return err
	}
	return resultStatus(after)
}

func (r runner) exportBuild(ctx context.Context, args []string) error {
	fs := r.flags("build export")
	connection := connectFlagsWithTimeout(fs, buildTimeout)
	id := fs.String("session", "", "явно выбранный session_id")
	raw := fs.Bool("raw", false, "вывести только XML")
	output := fs.String("output", "", "сохранить XML в новый файл UTF-8")
	if err := parse(fs, args); err != nil {
		return err
	}
	if *id == "" {
		return &UsageError{Err: fmt.Errorf("нужна --session")}
	}
	if *raw && *output != "" {
		return &UsageError{Err: fmt.Errorf("--output не сочетается с --raw")}
	}
	ctx, cancel := context.WithTimeout(ctx, connection.timeout)
	defer cancel()
	client, err := connection.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	value, err := client.Export(ctx, *id)
	if err != nil {
		return err
	}
	if *raw {
		_, err := io.WriteString(r.out, value.XML)
		return err
	}
	if *output == "" {
		return r.write(value)
	}
	if err := saveXML(*output, value.XML); err != nil {
		return err
	}
	return r.write(map[string]any{"path": *output, "input_sha256": value.InputSHA256, "revision": value.Revision})
}

func saveXML(path, xml string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(file, xml); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
