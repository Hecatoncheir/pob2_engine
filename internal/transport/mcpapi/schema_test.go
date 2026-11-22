package mcpapi

import (
	"encoding/json"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

func TestBuildInputAcceptsPartialSetSelection(t *testing.T) {
	for _, test := range []struct {
		name   string
		schema *jsonschema.Schema
		input  string
	}{
		{"edit", buildInputSchema[editInput](), `{"expected_revision":2,"request_id":"select","edit":{"selection":{"sets":{"skills":2}}}}`},
		{"compare", buildInputSchema[compareInput](), `{"expected_revision":2,"edit":{"selection":{"sets":{"items":2}}}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			resolved, err := test.schema.Resolve(nil)
			if err != nil {
				t.Fatal(err)
			}
			var input any
			if err := json.Unmarshal([]byte(test.input), &input); err != nil {
				t.Fatal(err)
			}
			if err := resolved.Validate(input); err != nil {
				t.Fatalf("частичный выбор набора отклонён схемой MCP: %v", err)
			}
		})
	}
}
