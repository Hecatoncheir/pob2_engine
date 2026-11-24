package application

import (
	"context"
	"strings"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
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
	edit := build.AddSupportEdit{SupportID: cmd.SupportID, Level: cmd.Level, Quality: cmd.Quality}
	if err := edit.ValidateID(); err != nil {
		return SupportPreview{}, failure(InvalidArgument, err.Error())
	}
	snapshot, err := s.supportSnapshot(ctx, cmd)
	if err != nil {
		return SupportPreview{}, err
	}
	data, err := s.editData(ctx)
	if err != nil {
		return SupportPreview{}, err
	}
	gem, level, err := edit.Resolve(data)
	if err != nil {
		return SupportPreview{}, failure(InvalidArgument, err.Error())
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
	after, group, err := before.WithSupportParameters(gem, level, cmd.Quality)
	if err != nil {
		return SupportPreview{}, failure(InvalidArgument, err.Error())
	}
	if s.calculator == nil {
		return SupportPreview{}, failure(CalculationFailure, "модуль расчёта не настроен")
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

func (s *Service) supportSnapshot(ctx context.Context, cmd PreviewSupport) (domain.Snapshot, error) {
	snapshot, err := s.Snapshot(ctx, cmd.SessionID)
	if err != nil {
		return domain.Snapshot{}, err
	}
	if snapshot.Revision != cmd.ExpectedRevision {
		return domain.Snapshot{}, failure(RevisionConflict, "снимок для предпросмотра устарел")
	}
	return snapshot, nil
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
