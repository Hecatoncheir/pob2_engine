package workbench

import (
	"context"
	"encoding/json"
	"errors"
	"maps"

	"github.com/Hecatoncheir/pob2_engine/internal/pob"
)

type skillMeasurement struct {
	DPS      *float64 `json:"dps"`
	EffectID string   `json:"effectId"`
	Part     int      `json:"part"`
	StatSet  int      `json:"statSet"`
}

func (s *Service) previewSupport(ctx context.Context, args map[string]any) (pob.WorkbenchResult, error) {
	gems := args["gems"].([]any)
	supportID := args["supportId"].(string)
	updated, err := addPreviewSupport(gems, supportID)
	if err != nil {
		return pob.WorkbenchResult{}, err
	}
	preview, revision, err := s.openPreview(ctx)
	if err != nil {
		return pob.WorkbenchResult{}, err
	}
	defer preview.Close()
	group, err := applyPreviewGroup(ctx, preview, map[string]any{"gems": gems, "groupId": args["groupId"]})
	if err != nil {
		return pob.WorkbenchResult{}, err
	}
	return comparePreviewSupport(ctx, preview, revision, group, args["targetIndex"], supportID, updated)
}

func addPreviewSupport(gems []any, supportID string) ([]map[string]any, error) {
	updated := make([]map[string]any, 0, len(gems)+1)
	for i, raw := range gems {
		gem := raw.(map[string]any)
		if gem["id"] == supportID {
			return nil, failure("INVALID_REQUEST", errors.New("эта поддержка уже есть в группе"))
		}
		copy := maps.Clone(gem)
		copy["sourceIndex"] = i + 1
		updated = append(updated, copy)
	}
	return append(updated, map[string]any{"id": supportID, "level": 1, "quality": 0, "enabled": true}), nil
}

func comparePreviewSupport(ctx context.Context, preview *pob.Workbench, revision, group int, targetIndex any, supportID string, updated []map[string]any) (pob.WorkbenchResult, error) {
	target := map[string]any{"groupId": group, "targetIndex": targetIndex, "supportId": supportID}
	before, err := measurePreviewSkill(ctx, preview, target)
	if err != nil {
		return pob.WorkbenchResult{}, err
	}
	if _, err := applyPreviewGroup(ctx, preview, map[string]any{"groupId": group, "gems": updated}); err != nil {
		return pob.WorkbenchResult{}, err
	}
	delete(target, "supportId")
	target["effectId"] = before.EffectID
	after, err := measurePreviewSkill(ctx, preview, target)
	if err != nil {
		return pob.WorkbenchResult{}, err
	}
	if before.Part != after.Part || before.StatSet != after.StatSet {
		return pob.WorkbenchResult{}, errors.New("изменился вариант выбранного умения")
	}
	return previewSupportResult(ctx, preview, revision, before.DPS, after.DPS)
}

func measurePreviewSkill(ctx context.Context, preview *pob.Workbench, target map[string]any) (skillMeasurement, error) {
	result, err := preview.Call(ctx, "measureSupportPreview", target)
	var measurement skillMeasurement
	if err != nil {
		return measurement, err
	}
	err = json.Unmarshal(result.Result, &measurement)
	return measurement, err
}

func previewSupportResult(ctx context.Context, preview *pob.Workbench, revision int, before, after *float64) (pob.WorkbenchResult, error) {
	report, err := preview.Call(ctx, "getBuildReport", nil)
	if err != nil {
		return pob.WorkbenchResult{}, err
	}
	summary := struct {
		Warnings json.RawMessage `json:"warnings"`
	}{Warnings: json.RawMessage("[]")}
	if err := json.Unmarshal(report.Result, &summary); err != nil {
		return pob.WorkbenchResult{}, err
	}
	var delta *float64
	if before != nil && after != nil {
		change := *after - *before
		delta = &change
	}
	return workbenchResult(revision, map[string]any{
		"source": "original-lua", "before": before, "after": after, "delta": delta, "warnings": summary.Warnings,
	})
}
