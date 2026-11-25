package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
	"github.com/Hecatoncheir/pob2_engine/internal/testreference"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
)

func TestBuildRejectsInvalidFlagsBeforeIO(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"файл для экспорта", []string{"export", "--session", "main", "--file", "unused.xml"}, "-file"},
		{"запрос для сравнения", []string{"compare", "--session", "main", "--revision", "1", "--file", "-", "--request-id", "compare"}, "-request-id"},
		{"сырой XML при импорте", []string{"import", "--session", "main", "--revision", "1", "--request-id", "import", "--file", "-", "--raw"}, "-raw"},
		{"два формата экспорта", []string{"export", "--session", "main", "--raw", "--output", "unused.xml"}, "--output не сочетается с --raw"},
		{"пропущенный идентификатор правки", []string{"edit", "--session", "main", "--revision", "1", "--file", "-"}, "нужен --request-id"},
		{"пропущенная ревизия", []string{"compare", "--session", "main", "--file", "-"}, "--revision"},
		{"пропущенный файл", []string{"compare", "--session", "main", "--revision", "1"}, "--file"},
		{"два способа подключения", []string{"compare", "--session", "main", "--revision", "1", "--file", "-", "--profile", "main", "--address", "127.0.0.1:1"}, "выберите --profile или --address"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"build"}, tc.args...)
			err := Run(t.Context(), args, unreadInput{t}, io.Discard, io.Discard)
			var usage *UsageError
			if !errors.As(err, &usage) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("нужна ошибка аргументов с %q, получено %v", tc.want, err)
			}
		})
	}
}

type unreadInput struct{ t *testing.T }

func (in unreadInput) Read([]byte) (int, error) {
	in.t.Fatal("команда читает stdin при неверных аргументах")
	return 0, io.EOF
}

func (unreadInput) Close() error { return nil }

func TestBuildHelpListsFlagsForSubcommand(t *testing.T) {
	cases := []struct {
		operation string
		present   []string
		absent    []string
	}{
		{"import", []string{"file", "revision", "request-id"}, []string{"edit", "raw", "output"}},
		{"edit", []string{"file", "revision", "request-id"}, []string{"edit", "raw", "output"}},
		{"compare", []string{"file", "revision", "edit"}, []string{"request-id", "raw", "output"}},
		{"export", []string{"raw", "output"}, []string{"file", "revision", "request-id", "edit"}},
	}
	for _, tc := range cases {
		t.Run(tc.operation, func(t *testing.T) {
			var help bytes.Buffer
			err := Run(t.Context(), []string{"build", tc.operation, "--help"}, unreadInput{t}, io.Discard, &help)
			if err != nil {
				t.Fatal(err)
			}
			checkBuildHelpFlags(t, help.String(), append([]string{"session", "profile"}, tc.present...), true)
			checkBuildHelpFlags(t, help.String(), tc.absent, false)
		})
	}
}

func checkBuildHelpFlags(t *testing.T, help string, flags []string, want bool) {
	t.Helper()
	for _, name := range flags {
		prefix := "\n  -" + name
		got := strings.Contains(help, prefix+" ") || strings.Contains(help, prefix+"\n")
		if got != want {
			t.Errorf("флаг --%s в справке: %t, должно быть %t", name, got, want)
		}
	}
}

func TestBuildSubcommandsThroughGRPC(t *testing.T) {
	server := startBuildServer(t)
	run := func(input string, operation string, flags ...string) string {
		t.Helper()
		args := []string{"build", operation, "--session", "main", "--address", server.Address, "--token", server.Token}
		var stdout, stderr bytes.Buffer
		if err := Run(t.Context(), append(args, flags...), io.NopCloser(strings.NewReader(input)), &stdout, &stderr); err != nil {
			t.Fatalf("%s: %v, stderr: %s", operation, err, &stderr)
		}
		return stdout.String()
	}
	xml, err := os.ReadFile("../../../testdata/reference/fireball.xml")
	if err != nil {
		t.Fatal(err)
	}
	run(string(xml), "import", "--file", "-", "--revision", "1", "--request-id", "import")
	run(string(xml), "compare", "--file", "-", "--revision", "2")
	const edit = `{"config":{"name":"customMods","value":{"type":"string","value":"100% increased Spell Damage"}}}`
	rawComparison := run(edit, "compare", "--file", "-", "--revision", "2", "--edit")
	var comparison application.BuildComparison
	if err := json.Unmarshal([]byte(rawComparison), &comparison); err != nil {
		t.Fatal(err)
	}
	if comparison.Comparison.After.Stats["TotalDPS"] <= comparison.Comparison.Before.Stats["TotalDPS"] {
		t.Fatal("JSON-правка не изменила результат сравнения")
	}
	run(edit, "edit", "--file", "-", "--revision", "2", "--request-id", "edit")
	if output := run("", "export", "--raw"); !strings.Contains(output, "100% increased Spell Damage") {
		t.Fatal("экспортированный XML не содержит сохранённую правку")
	}
}

func startBuildServer(t *testing.T) grpcapi.Ready {
	t.Helper()
	data, err := gamedata.Embedded()
	if err != nil {
		t.Fatal(err)
	}
	calculator, err := testreference.New(data)
	if err != nil {
		t.Fatal(err)
	}
	service := application.NewServiceWithCalculator(calculator)
	_, err = service.Create(t.Context(), application.CreateSession{SessionID: "main", BuildName: "Сборка", RequestID: "create"})
	if err != nil {
		t.Fatal(err)
	}
	server, err := grpcapi.Start("127.0.0.1:0", "", service)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := server.Close(ctx); err != nil {
			t.Error(err)
		}
	})
	return server.Ready()
}
