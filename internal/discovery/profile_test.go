package discovery

import (
	"context"
	"errors"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestExclusiveLockAndContext(t *testing.T) {
	p, err := Open(t.TempDir(), "main")
	if err != nil {
		t.Fatal(err)
	}
	first, err := p.TryLock("server")
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	if lock, err := p.TryLock("server"); !errors.Is(err, ErrLocked) {
		if lock != nil {
			lock.Close()
		}
		t.Fatalf("повторная блокировка: %v", err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	if _, err := p.Lock(ctx, "server"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ожидание блокировки: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := p.TryLock("server")
	if err != nil {
		t.Fatal(err)
	}
	second.Close()
}

func TestRecordReplacementAndOwnerCleanup(t *testing.T) {
	p, err := Open(t.TempDir(), "main")
	if err != nil {
		t.Fatal(err)
	}
	first := Record{InstanceID: "first", Address: "127.0.0.1:1", Token: "secret", ProtocolVersion: "pob2.v1", PID: os.Getpid()}
	if err := p.Publish(first); err != nil {
		t.Fatal(err)
	}
	second := first
	second.InstanceID = "second"
	if err := p.Publish(second); err != nil {
		t.Fatal(err)
	}
	if err := p.RemoveOwn("first"); err != nil {
		t.Fatal(err)
	}
	got, err := p.Read()
	if err != nil || got != second {
		t.Fatalf("удалён чужой файл: %v, %v", got, err)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(p.RecordPath())
		if err != nil || info.Mode().Perm() != 0600 {
			t.Fatalf("права файла: %v, %v", info, err)
		}
	}
	if err := p.RemoveOwn("second"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Read(); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("файл не удалён: %v", err)
	}
}

func TestProfileCannotEscapeBaseDirectory(t *testing.T) {
	base := t.TempDir()
	for _, name := range []string{"", "../other", "a/b", `a\b`, "C:", ".", ".."} {
		if _, err := Open(base, name); err == nil {
			t.Fatalf("принят профиль %q", name)
		}
	}
}
