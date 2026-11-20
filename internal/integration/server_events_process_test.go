package integration_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
)

func TestShutdownEventAndDurableRestartWithoutFlutter(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	base := t.TempDir()
	profile := testProfile(t, base, "server-events")
	args := profileArgs(base, "server-events")
	s := startServer(t, ctx, "--profile", "server-events", "--runtime-dir", base)
	want := commandSnapshot(t, ctx, args, "session", "create", "--session", "main", "--name", "Сохранённая", "--request-id", "create")
	client, err := grpcapi.Dial(ctx, s.ready.Address, s.ready.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	watch, err := client.WatchServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := watch.Recv()
	if err != nil || initial.GetState() != "ready" {
		t.Fatalf("начальный снимок сервера: %v %v", initial, err)
	}
	changes, err := client.Watch(ctx, "main", 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := changes.Recv(); err != nil {
		t.Fatal(err)
	}
	cli(t, ctx, 0, append([]string{"shutdown"}, args...)...)
	stopping, err := watch.Recv()
	if err != nil || stopping.GetState() != "draining" || stopping.GetInstanceId() != initial.GetInstanceId() {
		t.Fatalf("событие остановки: %v %v", stopping, err)
	}
	if _, err := watch.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("конец серверной подписки: %v", err)
	}
	if _, err := changes.Recv(); err == nil {
		t.Fatal("бесконечная подписка не прервана")
	}
	client.Close()
	select {
	case <-s.done:
		if s.err != nil {
			t.Fatal(s.err)
		}
	case <-ctx.Done():
		t.Fatal("сервер завис при shutdown")
	}
	awaitGone(t, ctx, profile)
	next := startServer(t, ctx, "--profile", "server-events", "--runtime-dir", base)
	if next.ready.InstanceID == initial.GetInstanceId() {
		t.Fatal("перезапуск не сменил instance_id")
	}
	if got := commandSnapshot(t, ctx, args, "session", "get", "--session", "main"); got != want {
		t.Fatalf("данные потеряны: %v", got)
	}
	cli(t, ctx, 0, append([]string{"shutdown"}, args...)...)
	select {
	case <-next.done:
		if next.err != nil {
			t.Fatal(next.err)
		}
	case <-ctx.Done():
		t.Fatal("повторная остановка зависла")
	}
}
