package mcpapi

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type equipmentDetailsInput struct {
	ExpectedRevision uint64 `json:"expected_revision,omitempty" jsonschema:"Ноль читает последний снимок сессии"`
	build.EquipmentDetailsQuery
}

type equipmentDetailsOutput struct {
	Revision uint64         `json:"revision"`
	Details  map[string]any `json:"details"`
}

func addEquipmentDetailsTool(server *mcp.Server, reader application.EquipmentDetailsCommands, id string) {
	openWorld := false
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_equipment_details",
		Description: "Прочитать описание, совместимость и подробное сравнение предмета из сборки или текста",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &openWorld},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input equipmentDetailsInput) (*mcp.CallToolResult, equipmentDetailsOutput, error) {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		view, err := reader.EquipmentDetails(ctx, id, input.ExpectedRevision, input.EquipmentDetailsQuery)
		if err != nil {
			return nil, equipmentDetailsOutput{}, err
		}
		result := equipmentDetailsOutput{Revision: view.Revision}
		err = json.Unmarshal(view.Details, &result.Details)
		return nil, result, err
	})
}
