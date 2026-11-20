package grpcapi_test

import (
	"context"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
)

func TestFullBuildCommandsAndEvents(t *testing.T) {
	_, client := start(t, application.NewService())
	ctx := t.Context()
	if _, err := client.Create(ctx, application.CreateSession{SessionID: "main", BuildName: "Сборка", RequestID: "create"}); err != nil {
		t.Fatal(err)
	}
	stream, err := client.Watch(ctx, "main", 1)
	if err != nil {
		t.Fatal(err)
	}
	xml, err := os.ReadFile("../../../testdata/reference/fireball.xml")
	if err != nil {
		t.Fatal(err)
	}
	imported, err := client.Apply(ctx, application.ApplyBuild{SessionID: "main", ExpectedRevision: 1, RequestID: "import", XML: string(xml)})
	if err != nil {
		t.Fatal(err)
	}
	event, err := stream.Recv()
	if err != nil || event.GetSnapshot().GetBuild().GetModel().GetClassName() != "Sorceress" || !slices.Contains(event.GetChangedFields(), "build.model.groups") {
		t.Fatalf("событие: %v %v", event, err)
	}
	calculated, err := client.Calculate(ctx, application.CalculateBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "calculate"})
	if err != nil {
		t.Fatal(err)
	}
	view, err := application.ViewSnapshot(calculated)
	if err != nil || view.Build.Calculation.Status != "complete" || view.Build.Calculation.Stats["TotalDPS"] <= 0 {
		t.Fatalf("расчёт: %+v %v", view, err)
	}
	edit := &build.Edit{Config: &build.ConfigEdit{Name: "customMods", Value: &build.ConfigValue{Type: "string", Value: "100% increased Spell Damage"}}}
	comparison, err := client.Compare(ctx, application.CompareBuild{SessionID: "main", ExpectedRevision: 3, Edit: edit})
	if err != nil || comparison.Comparison.After.Stats["TotalDPS"] <= comparison.Comparison.Before.Stats["TotalDPS"] {
		t.Fatalf("сравнение: %+v %v", comparison, err)
	}
	before, err := client.Snapshot(ctx, "main")
	if err != nil || before != calculated {
		t.Fatal("сравнение изменило сессию")
	}
	changed, err := client.Apply(ctx, application.ApplyBuild{SessionID: "main", ExpectedRevision: 3, RequestID: "edit", Edit: edit})
	if err != nil {
		t.Fatal(err)
	}
	if changed.CalculationJSON != "" || changed.BuildRevision != 4 {
		t.Fatal("старый расчёт остался текущим")
	}
	replay, err := client.Calculate(ctx, application.CalculateBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "calculate"})
	if err != nil || replay != calculated {
		t.Fatal("повтор потерял исходный результат")
	}
	value, err := client.Export(ctx, "main")
	if err != nil || value.XML != changed.BuildXML || imported.InputSHA256 == value.InputSHA256 {
		t.Fatal("неверный экспорт")
	}
	_, err = client.Apply(ctx, application.ApplyBuild{SessionID: "main", ExpectedRevision: 4, RequestID: "overflow", Edit: &build.Edit{Level: &build.LevelEdit{Level: int(uint64(1)<<32) + 10}}})
	if err == nil {
		t.Fatal("целое число усечено до допустимого значения")
	}
}

type canceledCalculation struct {
	application.BuildCommands
	started, canceled, finish chan struct{}
}

func (c canceledCalculation) Calculate(ctx context.Context, _ application.CalculateBuild) (domain.Snapshot, error) {
	close(c.started)
	<-ctx.Done()
	close(c.canceled)
	<-c.finish
	return domain.Snapshot{}, ctx.Err()
}

func TestCanceledStreamRemainsTaskUntilWorkerFinishes(t *testing.T) {
	commands := canceledCalculation{BuildCommands: application.NewService(), started: make(chan struct{}), canceled: make(chan struct{}), finish: make(chan struct{})}
	server, err := grpcapi.StartWithConfig("127.0.0.1:0", "token", commands, lifecycle.Config{Mode: lifecycle.Automatic, IdleTimeout: 100 * time.Millisecond, LeaseTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Close(ctx)
	})
	client, err := grpcapi.Dial(t.Context(), server.Ready().Address, server.Ready().Token)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		_, err := client.Calculate(ctx, application.CalculateBuild{SessionID: "main", ExpectedRevision: 1, RequestID: "calculation"})
		done <- err
	}()
	select {
	case <-commands.started:
	case <-time.After(2 * time.Second):
		t.Fatal("расчёт не начался")
	}
	cancel()
	select {
	case <-commands.canceled:
	case <-time.After(time.Second):
		t.Fatal("отмена не передана")
	}
	if err := <-done; err == nil {
		t.Fatal("отменённый RPC завершился успешно")
	}
	client.Close()
	info, err := grpcapi.Inspect(t.Context(), server.Ready().Address, server.Ready().Token)
	if err != nil || info.GetActiveTasks() != 1 {
		t.Fatalf("счётчик потерял работающую задачу: %v %v", info, err)
	}
	select {
	case <-server.Draining():
		t.Fatal("сервер завершает работающую задачу преждевременно")
	default:
	}
	close(commands.finish)
	select {
	case <-server.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("сервер не завершился после задачи")
	}
}
