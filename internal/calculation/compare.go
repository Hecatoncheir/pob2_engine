package calculation

import (
	"context"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

type Delta struct {
	Stat       string  `json:"stat"`
	Before     float64 `json:"before"`
	After      float64 `json:"after"`
	Difference float64 `json:"difference"`
}

type Comparison struct {
	Before Result  `json:"before"`
	After  Result  `json:"after"`
	Deltas []Delta `json:"deltas"`
}

func (e *Engine) Compare(ctx context.Context, before, after *build.Document) (Comparison, error) {
	old, err := e.Calculate(ctx, before)
	if err != nil {
		return Comparison{}, err
	}
	current, err := e.Calculate(ctx, after)
	if err != nil {
		return Comparison{}, err
	}
	result := Comparison{Before: old, After: current, Deltas: []Delta{}}
	if old.Stats == nil || current.Stats == nil {
		return result, nil
	}
	for _, name := range sortedKeys(old.Stats) {
		value, ok := current.Stats[name]
		if ok {
			result.Deltas = append(result.Deltas, Delta{Stat: name, Before: old.Stats[name], After: value, Difference: value - old.Stats[name]})
		}
	}
	return result, ctx.Err()
}
