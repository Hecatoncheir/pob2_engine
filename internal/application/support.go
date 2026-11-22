package application

import (
	"context"
	"strings"
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
	Level            int    `json:"level,omitempty"`
	Quality          int    `json:"quality,omitempty"`
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
	level := cmd.Level
	if level == 0 {
		level = 1
	}
	after, group, err := before.WithSupportParameters(gem, level, cmd.Quality)
	if err != nil {
		return SupportPreview{}, failure(InvalidArgument, err.Error())
	}
	comparison, err := s.calculator.Compare(ctx, before, after)
	if err != nil {
		return SupportPreview{}, err
	}
	return SupportPreview{SessionID: cmd.SessionID, Revision: snapshot.Revision, GroupID: group, Support: gemEntry(gem),
		Compatible: supportCompatible(comparison), Status: comparisonStatus(comparison), Comparison: comparison}, nil
}

func supportCompatible(value calculation.Comparison) bool {
	if value.Before.Stats == nil || value.After.Stats == nil {
		return false
	}
	for _, issue := range value.After.Issues {
		if strings.HasPrefix(issue.Code, "pob_support_") || issue.Code == "pob_gem_error" {
			return false
		}
	}
	return true
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
	data, err := s.editData(ctx)
	if err != nil {
		return domain.Snapshot{}, gamedata.Gem{}, err
	}
	gem, ok := data.Gem(cmd.SupportID)
	if !ok || !gem.Support {
		return domain.Snapshot{}, gamedata.Gem{}, failure(InvalidArgument, "неизвестная или неподдерживаемая поддержка")
	}
	if err := validateSupportParameters(cmd, gem, data.IsRuntime()); err != nil {
		return domain.Snapshot{}, gamedata.Gem{}, err
	}
	return snapshot, gem, nil
}

func validateSupportParameters(cmd PreviewSupport, gem gamedata.Gem, runtime bool) error {
	level := cmd.Level
	if level == 0 {
		level = 1
	}
	if level < 1 || level > 40 || cmd.Quality < 0 || cmd.Quality > 100 {
		return failure(InvalidArgument, "поддержка: уровень 1–40, качество 0–100")
	}
	if !runtime && (level != 1 || cmd.Quality != 0) {
		return failure(InvalidArgument, "режим Go поддерживает только уровень 1 и качество 0")
	}
	for _, candidate := range gem.Levels {
		if candidate.Level == level {
			return nil
		}
	}
	return failure(InvalidArgument, "этого уровня поддержки нет в активном PoB")
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
