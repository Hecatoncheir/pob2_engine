package mcpapi

import (
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
)

// Неподдержанный расчёт возвращает stats:null. Схема должна сохранять этот контракт.
func calculationOutputSchema[T any]() *jsonschema.Schema {
	schema, err := jsonschema.For[T](&jsonschema.ForOptions{TypeSchemas: map[reflect.Type]*jsonschema.Schema{
		reflect.TypeFor[map[string]float64](): {Types: []string{"object", "null"}, AdditionalProperties: &jsonschema.Schema{Type: "number"}},
	}})
	if err != nil {
		panic(err)
	}
	return schema
}
