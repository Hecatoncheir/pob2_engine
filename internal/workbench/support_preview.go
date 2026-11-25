package workbench

import (
	"context"
	"encoding/json"
	"errors"
	"maps"

	"github.com/Hecatoncheir/pob2_engine/internal/pob"
)

type preparedPreview struct {
	worker   *pob.Workbench
	revision int
	draft    string
	group    int
}

func (s *Service) compatibleSupports(ctx context.Context, arguments map[string]any) (_ pob.WorkbenchResult, err error) {
	defer s.finishPreview(&err)
	preview, err := s.preparePreview(ctx, arguments)
	if err != nil {
		return pob.WorkbenchResult{}, err
	}
	supports, err := preview.worker.Call(ctx, "getCompatibleSupports", map[string]any{"groupId": preview.group})
	supports.Revision = preview.revision
	return supports, err
}

func (s *Service) previewSupport(ctx context.Context, args map[string]any) (_ pob.WorkbenchResult, err error) {
	supportID := args["supportId"].(string)
	updated, err := addPreviewSupport(args["gems"].([]any), supportID)
	if err != nil {
		return pob.WorkbenchResult{}, err
	}
	defer s.finishPreview(&err)
	preview, err := s.preparePreview(ctx, args)
	if err != nil {
		return pob.WorkbenchResult{}, err
	}
	result, err := preview.worker.Call(ctx, "previewSupport", map[string]any{
		"groupId": preview.group, "targetIndex": args["targetIndex"], "supportId": supportID, "gems": updated,
	})
	if err != nil {
		return pob.WorkbenchResult{}, err
	}
	return previewSupportResult(preview.revision, result.Result)
}

func (s *Service) preparePreview(ctx context.Context, args map[string]any) (*preparedPreview, error) {
	export, err := s.session.Call(ctx, "exportBuild", nil)
	if err != nil {
		return nil, err
	}
	arguments := map[string]any{"gems": args["gems"], "groupId": args["groupId"]}
	draft, err := json.Marshal(arguments)
	if err != nil {
		return nil, err
	}
	if s.preview != nil && s.preview.revision == export.Revision && s.preview.draft == string(draft) {
		return s.preview, nil
	}
	if s.preview == nil {
		worker, err := pob.NewWorkbench(ctx, s.installation)
		if err != nil {
			return nil, err
		}
		s.preview = &preparedPreview{worker: worker}
	}
	group, err := loadPreviewGroup(ctx, s.preview.worker, export.Result, arguments)
	if err != nil {
		return nil, err
	}
	s.preview.revision, s.preview.draft, s.preview.group = export.Revision, string(draft), group
	return s.preview, nil
}

func loadPreviewGroup(ctx context.Context, worker *pob.Workbench, exported json.RawMessage, arguments map[string]any) (int, error) {
	var document struct {
		XML string `json:"xml"`
	}
	if err := json.Unmarshal(exported, &document); err != nil {
		return 0, err
	}
	if _, err := worker.Call(ctx, "loadBuild", map[string]any{"xml": document.XML}); err != nil {
		return 0, err
	}
	operation := "addSkillGroup"
	if arguments["groupId"] != nil {
		operation = "updateSkillGroup"
	}
	edited, err := worker.Call(ctx, operation, arguments)
	if err != nil {
		return 0, err
	}
	var group struct {
		ID int `json:"groupId"`
	}
	err = json.Unmarshal(edited.Result, &group)
	return group.ID, err
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

func previewSupportResult(revision int, result json.RawMessage) (pob.WorkbenchResult, error) {
	summary := struct {
		Before   *float64        `json:"before"`
		After    *float64        `json:"after"`
		Warnings json.RawMessage `json:"warnings"`
	}{Warnings: json.RawMessage("[]")}
	if err := json.Unmarshal(result, &summary); err != nil {
		return pob.WorkbenchResult{}, err
	}
	before, after := summary.Before, summary.After
	var delta *float64
	if before != nil && after != nil {
		change := *after - *before
		delta = &change
	}
	return workbenchResult(revision, map[string]any{
		"source": "original-lua", "before": before, "after": after, "delta": delta, "warnings": summary.Warnings,
	})
}
