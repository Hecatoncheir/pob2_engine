package maintenance

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestReplaceRunningExecutable(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	target, staged := filepath.Join(dir, "running.exe"), filepath.Join(dir, "staged.exe")
	for _, path := range []string{target, staged} {
		if err := os.WriteFile(path, data, 0755); err != nil {
			t.Fatal(err)
		}
	}
	old := executableInfo(t, target)
	startReplacementHelper(t, target)
	if err := replaceExecutable(staged, target); err != nil {
		t.Fatal(err)
	}
	if os.SameFile(old, executableInfo(t, target)) {
		t.Fatal("файл по исходному пути не заменён")
	}
	command := exec.CommandContext(t.Context(), target, "-test.run=^$")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("новый файл не запускается: %v\n%s", err, output)
	}
}

func executableInfo(t *testing.T, path string) os.FileInfo {
	t.Helper()
	info, err := snapshotFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return info
}

func startReplacementHelper(t *testing.T, executable string) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), executable, "-test.run=^TestReplacementHelper$")
	cmd.Env = append(os.Environ(), "POB2_TEST_REPLACEMENT_HELPER=1")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = stdin.Close()
		_ = cmd.Wait()
	})
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() || scanner.Text() != "ready" {
		t.Fatalf("дочерний процесс не сообщил о готовности: %q, %v", scanner.Text(), scanner.Err())
	}
}

func TestReplacementHelper(t *testing.T) {
	if os.Getenv("POB2_TEST_REPLACEMENT_HELPER") != "1" {
		return
	}
	fmt.Println("ready")
	if _, err := io.Copy(io.Discard, os.Stdin); err != nil {
		t.Fatal(err)
	}
}
