package application

import (
	"context"
	"unicode/utf8"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

type PreviewSupport struct {
	SessionID        string `json:"session_id"`
	ExpectedRevision uint64 `json:"expected_revision"`
	SupportID        string `json:"support_id"`
}

type SupportPreview struct {
	SessionID  string                 `json:"session_id"`
	Revision   uint64                 `json:"revision"`
	GroupID    int                    `json:"group_id"`
	Support    CatalogEntry           `json:"support"`
	Compatible bool                   `json:"compatible"`
	Status     string                 `json:"status"`
	Comparison calculation.Comparison `json:"comparison"`
}

func (s *Service) PreviewSupport(ctx context.Context, cmd PreviewSupport) (SupportPreview, error) {
	snapshot, gem, err := s.supportSnapshot(ctx, cmd)
	if err != nil {
		return SupportPreview{}, err
	}
	release, err := s.acquire(ctx)
	if err != nil {
		return SupportPreview{}, err
	}
	defer release()
	before, err := build.Parse([]byte(snapshot.BuildXML))
	if err != nil {
		return SupportPreview{}, failure(InvalidArgument, err.Error())
	}
	after, group, err := before.WithSupport(gem)
	if err != nil {
		return SupportPreview{}, failure(InvalidArgument, err.Error())
	}
	comparison, err := s.calculator.Compare(ctx, before, after)
	if err != nil {
		return SupportPreview{}, err
	}
	return SupportPreview{SessionID: cmd.SessionID, Revision: snapshot.Revision, GroupID: group, Support: gemEntry(gem),
		Compatible: comparison.Before.Stats != nil && comparison.After.Stats != nil, Status: comparisonStatus(comparison), Comparison: comparison}, nil
}

func (s *Service) supportSnapshot(ctx context.Context, cmd PreviewSupport) (domain.Snapshot, gamedata.Gem, error) {
	if !utf8.ValidString(cmd.SupportID) || len(cmd.SupportID) > 256 {
		return domain.Snapshot{}, gamedata.Gem{}, failure(InvalidArgument, "support_id: нужен UTF-8 до 256 байт")
	}
	snapshot, err := s.Snapshot(ctx, cmd.SessionID)
	if err != nil {
		return domain.Snapshot{}, gamedata.Gem{}, err
	}
	if snapshot.Revision != cmd.ExpectedRevision {
		return domain.Snapshot{}, gamedata.Gem{}, failure(RevisionConflict, "снимок для предпросмотра устарел")
	}
	if err := s.loadData(); err != nil {
		return domain.Snapshot{}, gamedata.Gem{}, err
	}
	gem, ok := s.data.Gem(cmd.SupportID)
	if !ok || !gem.Support {
		return domain.Snapshot{}, gamedata.Gem{}, failure(InvalidArgument, "неизвестная или неподдерживаемая поддержка")
	}
	return snapshot, gem, nil
}

func comparisonStatus(value calculation.Comparison) string {
	if value.Before.Status == "unsupported" || value.After.Status == "unsupported" {
		return "unsupported"
	}
	if value.Before.Status == "partial" || value.After.Status == "partial" {
		return "partial"
	}
	return "complete"
}
