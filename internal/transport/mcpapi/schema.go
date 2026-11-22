package mcpapi

import (
	"reflect"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/google/jsonschema-go/jsonschema"
)

// Неподдержанный расчёт возвращает stats:null. Схема должна сохранять этот контракт.
func calculationOutputSchema[T any]() *jsonschema.Schema {
	return toolSchema[T](&jsonschema.ForOptions{TypeSchemas: map[reflect.Type]*jsonschema.Schema{
		reflect.TypeFor[map[string]float64](): {Types: []string{"object", "null"}, AdditionalProperties: &jsonschema.Schema{Type: "number"}},
	}})
}

// В правке можно передать только меняющиеся наборы. Снимок по-прежнему содержит все идентификаторы.
func buildInputSchema[T any]() *jsonschema.Schema {
	sets := toolSchema[build.ActiveSets](nil)
	sets.Required = nil
	return toolSchema[T](&jsonschema.ForOptions{TypeSchemas: map[reflect.Type]*jsonschema.Schema{
		reflect.TypeFor[build.ActiveSets](): sets,
	}})
}

func toolSchema[T any](options *jsonschema.ForOptions) *jsonschema.Schema {
	schema, err := jsonschema.For[T](options)
	if err != nil {
		panic(err)
	}
	return schema
}
