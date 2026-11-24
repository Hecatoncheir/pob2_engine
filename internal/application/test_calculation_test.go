package application

import (
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
	"github.com/Hecatoncheir/pob2_engine/internal/testreference"
)

func serviceWithTestCalculation(t testing.TB) *Service {
	t.Helper()
	return NewServiceWithCalculator(testCalculator(t))
}

func testCalculator(t testing.TB) calculation.Calculator {
	t.Helper()
	data, err := gamedata.Embedded()
	if err != nil {
		t.Fatal(err)
	}
	calculator, err := testreference.New(data)
	if err != nil {
		t.Fatal(err)
	}
	return calculator
}
