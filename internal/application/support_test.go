package application

import (
	"context"
	"errors"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func preview(t *testing.T, s *Service, support string) SupportPreview {
	t.Helper()
	value, err := s.PreviewSupport(t.Context(), PreviewSupport{SessionID: "main", ExpectedRevision: 2, SupportID: support})
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func importPreviewXML(t *testing.T, s *Service, xml string) {
	t.Helper()
	_, err := s.Apply(t.Context(), ApplyBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "variant", XML: xml})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSupportPreviewDoesNotWriteOrPublish(t *testing.T) {
	journal := &testJournal{}
	s, initial := importedService(t, journal)
	sub, err := s.Watch(t.Context(), "main", 2)
	if err != nil {
		t.Fatal(err)
	}
	value := preview(t, s, "Rapid Casting I")
	if !value.Compatible || value.Status != "complete" || value.GroupID != 1 || value.Revision != 2 {
		t.Fatalf("неверный предпросмотр: %+v", value)
	}
	if value.Comparison.Before.InputSHA256 != initial.InputSHA256 || value.Comparison.After.InputSHA256 == initial.InputSHA256 {
		t.Fatal("нет привязки вариантов к входам")
	}
	if len(value.Comparison.Deltas) == 0 || value.Comparison.After.Stats["TotalDPS"] <= value.Comparison.Before.Stats["TotalDPS"] {
		t.Fatal("нет изменения показателей")
	}
	if current, _ := s.Snapshot(t.Context(), "main"); current != initial {
		t.Fatal("предпросмотр изменил сессию")
	}
	if len(journal.changes) != 2 {
		t.Fatal("предпросмотр записан в журнал")
	}
	assertNoPreviewEvent(t, sub)
}

func assertNoPreviewEvent(t *testing.T, sub *Subscription) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	if _, err := sub.Next(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("предпросмотр опубликовал событие: %v", err)
	}
}

func TestSupportPreviewUsesFullGroupAndRejectsFamilyRepeat(t *testing.T) {
	s, _ := importedService(t, nil)
	file, err := os.ReadFile("../../testdata/reference/cases/support-rapid.xml")
	if err != nil {
		t.Fatal(err)
	}
	importPreviewXML(t, s, string(file))
	combined, err := s.PreviewSupport(t.Context(), PreviewSupport{SessionID: "main", ExpectedRevision: 3, SupportID: "Controlled Destruction"})
	if err != nil {
		t.Fatal(err)
	}
	if !combined.Compatible || combined.Comparison.Before.Stats["Speed"] != combined.Comparison.After.Stats["Speed"] {
		t.Fatal("пропала установленная поддержка скорости")
	}
	duplicate, err := s.PreviewSupport(t.Context(), PreviewSupport{SessionID: "main", ExpectedRevision: 3, SupportID: "Rapid Casting I"})
	if err != nil {
		t.Fatal(err)
	}
	if duplicate.Compatible || duplicate.Status != "unsupported" || duplicate.Comparison.After.Stats != nil {
		t.Fatal("повтор семейства рассчитан")
	}
	if len(duplicate.Comparison.Deltas) != 0 || len(duplicate.Comparison.After.Issues) == 0 {
		t.Fatal("нет причины отказа или выданы нулевые дельты")
	}
}

func TestSupportPreviewPreservesUnsupportedXML(t *testing.T) {
	s, initial := importedService(t, nil)
	xml := strings.Replace(initial.BuildXML, "</Config>", `<Input name="conditionEnemyShocked" boolean="true"/></Config>`, 1)
	importPreviewXML(t, s, xml)
	value, err := s.PreviewSupport(t.Context(), PreviewSupport{SessionID: "main", ExpectedRevision: 3, SupportID: "Rapid Casting I"})
	if err != nil {
		t.Fatal(err)
	}
	if value.Compatible || value.Comparison.Before.Stats != nil || value.Comparison.After.Stats != nil {
		t.Fatal("предпросмотр потерял неподдержанное условие")
	}
	_, err = s.PreviewSupport(t.Context(), PreviewSupport{SessionID: "main", ExpectedRevision: 2, SupportID: "Rapid Casting I"})
	expectCode(t, err, RevisionConflict)
}

func TestSupportPreviewRejectsWrongGemAndCancelledRequest(t *testing.T) {
	s, _ := importedService(t, nil)
	for _, id := range []string{"", "Unknown Support", "Fireball", string([]byte{0xff}), strings.Repeat("a", 257)} {
		_, err := s.PreviewSupport(t.Context(), PreviewSupport{SessionID: "main", ExpectedRevision: 2, SupportID: id})
		expectCode(t, err, InvalidArgument)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := s.PreviewSupport(ctx, PreviewSupport{SessionID: "main", ExpectedRevision: 2, SupportID: "Rapid Casting I"})
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestSupportPreviewSharesCalculationQueue(t *testing.T) {
	s, _ := importedService(t, nil)
	release := holdWorkers(s)
	defer release()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	answers := make(chan error, cap(s.admission))
	for range cap(s.admission) {
		go func() {
			_, err := s.PreviewSupport(ctx, PreviewSupport{SessionID: "main", ExpectedRevision: 2, SupportID: "Rapid Casting I"})
			answers <- err
		}()
	}
	waitAdmission(t, s, cap(s.admission))
	_, err := s.Compare(t.Context(), CompareBuild{SessionID: "main", ExpectedRevision: 2, Edit: &build.Edit{Level: &build.LevelEdit{Level: 2}}})
	expectCode(t, err, ResourceExhausted)
	_, err = s.PreviewSupport(t.Context(), PreviewSupport{SessionID: "main", ExpectedRevision: 2, SupportID: "Rapid Casting I"})
	expectCode(t, err, ResourceExhausted)
	cancel()
	for range cap(s.admission) {
		if err := <-answers; !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	}
	waitAdmission(t, s, 0)
}

func TestQueuedSupportPreviewKeepsOriginalSnapshot(t *testing.T) {
	s, initial := importedService(t, nil)
	release := holdWorkers(s)
	answers := make(chan SupportPreview, 1)
	failures := make(chan error, 1)
	go func() {
		value, err := s.PreviewSupport(t.Context(), PreviewSupport{SessionID: "main", ExpectedRevision: 2, SupportID: "Rapid Casting I"})
		answers <- value
		failures <- err
	}()
	waitAdmission(t, s, 1)
	changed, err := s.Apply(t.Context(), ApplyBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "while-preview", Edit: &build.Edit{Level: &build.LevelEdit{Level: 2}}})
	release()
	if err != nil {
		t.Fatal(err)
	}
	value := <-answers
	if err := <-failures; err != nil {
		t.Fatal(err)
	}
	if value.Revision != initial.Revision || value.Comparison.Before.InputSHA256 != initial.InputSHA256 {
		t.Fatal("подменён вход ожидающего предпросмотра")
	}
	if current, _ := s.Snapshot(t.Context(), "main"); current != changed {
		t.Fatal("предпросмотр затёр новую правку")
	}
}

func TestSupportPreviewKeepsRequirementWarnings(t *testing.T) {
	s, _ := importedService(t, nil)
	xml, err := os.ReadFile("../../testdata/reference/cases/gem-requirement-warning.xml")
	if err != nil {
		t.Fatal(err)
	}
	importPreviewXML(t, s, string(xml))
	value, err := s.PreviewSupport(t.Context(), PreviewSupport{SessionID: "main", ExpectedRevision: 3, SupportID: "Controlled Destruction"})
	if err != nil {
		t.Fatal(err)
	}
	if !value.Compatible || value.Status != "partial" || len(value.Comparison.After.Issues) == 0 {
		t.Fatal("потеряно предупреждение о требованиях")
	}
}

func TestSupportPreviewForThreeSpells(t *testing.T) {
	for _, spell := range []struct {
		name, path string
		dps        float64
	}{
		{"Fireball", "../../testdata/reference/fireball.xml", 8.9166666666667},
		{"Frostbolt", "../../testdata/reference/0.0.2/cases/frostbolt-minimal.xml", 9.1},
		{"Spark", "../../testdata/reference/0.0.2/cases/spark-minimal.xml", 8.5642857142857},
	} {
		t.Run(spell.name, func(t *testing.T) { checkSpellPreview(t, spell.path, spell.name, spell.dps) })
	}
}

func checkSpellPreview(t *testing.T, path, name string, dps float64) {
	t.Helper()
	s, _ := importedService(t, nil)
	xml, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	importPreviewXML(t, s, string(xml))
	value, err := s.PreviewSupport(t.Context(), PreviewSupport{SessionID: "main", ExpectedRevision: 3, SupportID: "Controlled Destruction"})
	if err != nil {
		t.Fatal(err)
	}
	if value.Comparison.Before.Skill != name || math.Abs(value.Comparison.Before.Stats["TotalDPS"]-dps) > 1e-9 {
		t.Fatal("рассчитано другое умение")
	}
	if !value.Compatible || value.Comparison.After.Stats["CritChance"] != 0 {
		t.Fatal("поддержка не отключила критические попадания")
	}
}
