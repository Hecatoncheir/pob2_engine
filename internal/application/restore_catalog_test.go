package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

type changingCatalogCalculator struct {
	calculation.Calculator
	data *gamedata.Dataset
	err  error
}

func (c *changingCatalogCalculator) CatalogData(context.Context) (*gamedata.Dataset, error) {
	return c.data, c.err
}

func TestRecoveryReplaysEditWithoutCurrentCatalog(t *testing.T) {
	journal := &testJournal{}
	service, snapshot := importedService(t, journal)
	data, err := gamedata.Embedded()
	if err != nil {
		t.Fatal(err)
	}
	provider := &changingCatalogCalculator{data: data}
	service.calculator = provider
	command := ApplyBuild{SessionID: snapshot.SessionID, ExpectedRevision: snapshot.Revision, RequestID: "gems",
		Edit: &build.Edit{Gems: &build.GemsEdit{GroupID: 1, Enabled: true,
			Gems: []build.GemInput{{ID: "Fireball", Level: 1, Enabled: true}}}}}
	edited, err := service.Apply(t.Context(), command)
	if err != nil {
		t.Fatal(err)
	}
	provider.err = errors.New("прежний каталог больше недоступен")
	restored, err := OpenServiceWithCalculator(t.Context(), journal, provider)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := restored.Apply(t.Context(), command)
	if err != nil || actual != edited {
		t.Fatal("не восстановлен исходный ответ правки", err)
	}
	current, err := restored.Snapshot(t.Context(), snapshot.SessionID)
	if err != nil || current != edited {
		t.Fatal("повтор изменил текущую сессию", err)
	}
}

func TestRecoveryRejectsDamagedSavedEdit(t *testing.T) {
	journal := &testJournal{}
	service, snapshot := importedService(t, journal)
	_, err := service.Apply(t.Context(), ApplyBuild{SessionID: snapshot.SessionID,
		ExpectedRevision: snapshot.Revision, RequestID: "level", Edit: &build.Edit{Level: &build.LevelEdit{Level: 2}}})
	if err != nil {
		t.Fatal(err)
	}
	last := len(journal.changes) - 1
	valid := journal.changes[last]
	mutations := []func(*Change){
		func(change *Change) { change.Result.InputSHA256 = "damaged" },
		func(change *Change) { change.Result.Revision++ },
		func(change *Change) { change.Result.BuildName = "Другая сборка" },
		func(change *Change) { change.Result.BuildXML = "<broken>" },
	}
	for _, mutate := range mutations {
		journal.changes[last] = valid
		mutate(&journal.changes[last])
		if restored, err := OpenService(t.Context(), journal); err == nil || restored != nil {
			t.Fatal("повреждённая правка принята")
		}
	}
}
