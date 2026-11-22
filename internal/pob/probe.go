package pob

import (
	"context"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	multigroup "github.com/Hecatoncheir/pob2_engine/testdata/reference/0.0.5"
)

func probeInstallation(ctx context.Context, i Installation) error {
	calculator, err := NewCalculator(i)
	if err != nil {
		return err
	}
	defer calculator.Close()
	if _, err := calculator.CatalogData(ctx); err != nil {
		return fmt.Errorf("чтение каталога: %w", err)
	}
	files, err := multigroup.Builds.ReadDir(".")
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := probeBuild(ctx, calculator, file.Name()); err != nil {
			return fmt.Errorf("расчёт %s: %w", file.Name(), err)
		}
	}
	return nil
}

func probeBuild(ctx context.Context, calculator *Calculator, name string) error {
	xml, err := multigroup.Builds.ReadFile(name)
	if err != nil {
		return err
	}
	doc, err := build.Parse(xml)
	if err != nil {
		return err
	}
	result, err := calculator.Calculate(ctx, doc)
	if err != nil {
		return err
	}
	if result.Skill == "" || result.Stats["Life"] <= 0 || result.Stats["FullDPS"] <= 0 {
		return fmt.Errorf("не получены показатели многогрупповой сборки")
	}
	for _, issue := range result.Issues {
		if issue.Code == "pob_gem_error" {
			return fmt.Errorf("PoB не загрузил камень: %s", issue.Message)
		}
	}
	return nil
}
