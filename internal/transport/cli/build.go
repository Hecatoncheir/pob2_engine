package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

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
	connection := connectFlags(fs)
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
	ctx, cancel := context.WithTimeout(ctx, connection.timeout)
	defer cancel()
	if *file != "" {
		if connection.profile != "" || connection.address != "" || connection.startServer || connection.dataDir != "" || connection.baseDir != "" || *request != "" || *revision != 0 {
			return &UsageError{Err: fmt.Errorf("расчёт --file не принимает параметры сервера и сессии")}
		}
		data, err := r.readInput(ctx, *file)
		if err != nil {
			return err
		}
		doc, err := build.Parse(data)
		if err != nil {
			return &UsageError{Err: err}
		}
		dataset, err := gamedata.Embedded()
		if err != nil {
			return err
		}
		engine, err := calculation.New(dataset)
		if err != nil {
			return err
		}
		result, err := engine.Calculate(ctx, doc)
		if err != nil {
			return err
		}
		if err := r.write(result); err != nil {
			return err
		}
		return resultStatus(result.Status)
	}
	client, err := connection.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	snapshot, err := client.Calculate(ctx, application.CalculateBuild{SessionID: *id, ExpectedRevision: *revision, RequestID: *request})
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
	operation := args[0]
	if operation != "import" && operation != "edit" && operation != "export" && operation != "compare" {
		return &UsageError{Err: fmt.Errorf("неизвестная операция build %q", operation)}
	}
	fs := r.flags("build " + operation)
	connection := connectFlags(fs)
	id := fs.String("session", "", "явно выбранный session_id")
	file := fs.String("file", "", "XML или JSON-файл правки; - для stdin")
	revision := fs.Uint64("revision", 0, "ожидаемая ревизия")
	request := fs.String("request-id", "", "ID правки для безопасного повтора")
	asEdit := fs.Bool("edit", false, "при compare входной файл содержит JSON правки")
	raw := fs.Bool("raw", false, "при export вывести только XML")
	output := fs.String("output", "", "при export сохранить XML в новый файл UTF-8")
	if err := parse(fs, args[1:]); err != nil {
		return err
	}
	if *id == "" {
		return &UsageError{Err: fmt.Errorf("нужна --session")}
	}
	if *output != "" && (operation != "export" || *raw) {
		return &UsageError{Err: fmt.Errorf("--output предназначен для export и не сочетается с --raw")}
	}
	ctx, cancel := context.WithTimeout(ctx, connection.timeout)
	defer cancel()
	var xml string
	var edit *build.Edit
	if operation != "export" {
		data, err := r.readInput(ctx, *file)
		if err != nil {
			return err
		}
		if operation == "edit" || operation == "compare" && *asEdit {
			edit, err = parseEdit(data)
			if err != nil {
				return err
			}
		} else {
			xml = string(data)
		}
	}
	if *raw && operation != "export" || *asEdit && operation != "compare" {
		return &UsageError{Err: fmt.Errorf("--raw предназначен для export, --edit для compare")}
	}
	client, err := connection.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	switch operation {
	case "import", "edit":
		value, err := client.Apply(ctx, application.ApplyBuild{SessionID: *id, ExpectedRevision: *revision, RequestID: *request, XML: xml, Edit: edit})
		if err != nil {
			return err
		}
		return r.write(value)
	case "compare":
		value, err := client.Compare(ctx, application.CompareBuild{SessionID: *id, ExpectedRevision: *revision, XML: xml, Edit: edit})
		if err != nil {
			return err
		}
		if err := r.write(value); err != nil {
			return err
		}
		if value.Comparison.Before.Status == "unsupported" || value.Comparison.After.Status == "unsupported" {
			return resultStatus("unsupported")
		}
		if err := resultStatus(value.Comparison.Before.Status); err != nil {
			return err
		}
		return resultStatus(value.Comparison.After.Status)
	case "export":
		if *file != "" || *request != "" || *revision != 0 {
			return &UsageError{Err: fmt.Errorf("export не принимает --file, --revision и --request-id")}
		}
		value, err := client.Export(ctx, *id)
		if err != nil {
			return err
		}
		if *raw {
			_, err = io.WriteString(r.out, value.XML)
			return err
		}
		if *output != "" {
			if err := saveXML(*output, value.XML); err != nil {
				return err
			}
			return r.write(map[string]any{"path": *output, "input_sha256": value.InputSHA256, "revision": value.Revision})
		}
		return r.write(value)
	}
	return nil
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
