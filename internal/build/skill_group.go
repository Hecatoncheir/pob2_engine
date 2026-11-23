package build

import (
	"fmt"
	"unicode/utf8"
)

type SkillGroupEdit struct {
	GroupID int             `json:"group_id,omitempty"`
	Label   *string         `json:"label,omitempty"`
	Delete  bool            `json:"delete,omitempty"`
	Gems    []SkillGroupGem `json:"gems,omitempty"`
}

type SkillGroupGem struct {
	ID          string `json:"id"`
	Level       int    `json:"level"`
	Quality     int    `json:"quality"`
	Enabled     *bool  `json:"enabled,omitempty"`
	SourceIndex int    `json:"source_index,omitempty"`
}

func (e SkillGroupEdit) validate() error {
	if e.GroupID < 0 || e.GroupID > 100 || len(e.Gems) > 12 || (e.Label != nil && (len(*e.Label) > 200 || !utf8.ValidString(*e.Label))) {
		return fmt.Errorf("неверная группа умений")
	}
	if e.Delete && e.GroupID > 0 && len(e.Gems) == 0 && e.Label == nil {
		return nil
	}
	if e.Delete || len(e.Gems) == 0 {
		return fmt.Errorf("укажите группу для удаления или список камней для правки")
	}
	for _, gem := range e.Gems {
		if err := gem.validate(e.GroupID > 0); err != nil {
			return err
		}
	}
	return nil
}

func (g SkillGroupGem) validate(existing bool) error {
	if g.ID == "" || len(g.ID) > 256 || !utf8.ValidString(g.ID) || g.Level < 1 || g.Level > 40 || g.Quality < 0 || g.Quality > 100 {
		return fmt.Errorf("некорректный ID, уровень или качество камня")
	}
	if g.SourceIndex < 0 || g.SourceIndex > 12 || (g.SourceIndex != 0 && !existing) {
		return fmt.Errorf("неверный индекс исходного камня")
	}
	return nil
}

// WithResolvedSkillGroup переносит только активные группы из результата PoB.
func (d *Document) WithResolvedSkillGroup(result *Document) (*Document, error) {
	updated, err := d.copyForEdit()
	if err != nil {
		return nil, err
	}
	if err := updated.replaceActiveSkills(result); err != nil {
		return nil, err
	}
	updated.clearCachedResults()
	if _, err := updated.Model(); err != nil {
		return nil, err
	}
	return updated, nil
}

func (d *Document) replaceActiveSkills(result *Document) error {
	oldSet, err := d.active("Skills", "SkillSet", "activeSkillSet")
	if err != nil {
		return err
	}
	newSet, err := result.active("Skills", "SkillSet", "activeSkillSet")
	if err != nil {
		return err
	}
	for _, group := range oldSet.children("Skill") {
		oldSet.remove(group)
	}
	for _, group := range newSet.children("Skill") {
		oldSet.parts = append(oldSet.parts, part{element: group})
	}
	newBuild, _ := result.root.child("Build")
	main, err := mainGroup(newBuild)
	if err != nil {
		return err
	}
	info, _ := d.root.child("Build")
	setMainGroup(info, main)
	return nil
}
