package pob

import (
	"slices"
	"testing"
)

func TestPackedTreePathPreservesPoBRules(t *testing.T) {
	input := []uint32{5, 4, 1,
		0, 1, 0, pathAllocated | pathUnlocked | pathClassStart, 1,
		0, 0, 1000, pathUnlocked, 2,
		0, 0, 1000, pathUnlocked | pathMastery, 1,
		0, 2, 0, pathAllocated | pathUnlocked, 0,
		0, 0, 1000, pathUnlocked, 0,
		1, 2, 3, 4, 0,
	}
	output := make([]uint32, 24)
	written := solvePathWords(&input[0], uint32(len(input)), &output[0], uint32(len(output)))
	want := []uint32{2,
		0, 1, 0, 1, 1, 1, 2, 1, 2, 0, 0, 0, 1000, 0, 0,
		2, 0, 3, 1,
	}
	if !slices.Equal(output[:written], want) {
		t.Fatalf("неверные расстояния, корни или связи: %v", output[:written])
	}
	if written := solvePathWords(&input[0], uint32(len(input)), &output[0], 16); written != 0 {
		t.Fatalf("малый выходной буфер должен вызвать переход на Lua: %d", written)
	}
}

func TestPackedTreePathRejectsInvalidGraph(t *testing.T) {
	for _, input := range [][]uint32{
		{1, 0}, {1, 0, 0}, {1, 0, 1, 0, 0, 0, 0, 0, 99},
		{1, 0, 1, 0, 3, 0, 0, 0, 0}, {1, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 1, 0, 0, 1001, 0, 0, 0}, {1, 0, 1, 0, 0, 0, 16, 0, 0},
	} {
		if _, _, ok := unpackPathGraph(input); ok {
			t.Errorf("принят некорректный граф: %v", input)
		}
	}
}

func TestTreePathCallbackUsesVerifiedPassiveSpec(t *testing.T) {
	t.Chdir(t.TempDir())
	const source = "test"
	got := treePathProgram(source, func() uintptr {
		t.Error("регистрация Go для неизвестного исходника PassiveSpec")
		return 0
	})
	if got != source {
		t.Fatalf("должен использоваться штатный Lua: %q", got)
	}
	if got := withTreePathCallback(source, ^uintptr(0)); !slices.Contains([]string{
		"POB2_GO_PATH_CALLBACK = 0xffffffffULL\ntest",
		"POB2_GO_PATH_CALLBACK = 0xffffffffffffffffULL\ntest",
	}, got) {
		t.Fatalf("адрес callback потерял точность: %q", got)
	}
}
