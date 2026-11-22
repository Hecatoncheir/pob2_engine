package grpcapi

import (
	"encoding/json"
	"fmt"
	"math"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
)

func wireIssue(value build.Issue) *pb.BuildIssue {
	return &pb.BuildIssue{Code: value.Code, Path: value.Path, Message: value.Message}
}
func localIssue(value *pb.BuildIssue) build.Issue {
	return build.Issue{Code: value.GetCode(), Path: value.GetPath(), Message: value.GetMessage()}
}
func wireSets(value build.ActiveSets) *pb.ActiveSets {
	return &pb.ActiveSets{Skills: uint32(value.Skills), Items: uint32(value.Items), Tree: uint32(value.Tree), Config: uint32(value.Config)}
}
func localSets(value *pb.ActiveSets) build.ActiveSets {
	return build.ActiveSets{Skills: int(value.GetSkills()), Items: int(value.GetItems()), Tree: int(value.GetTree()), Config: int(value.GetConfig())}
}

func wireConfig(value build.ConfigValue) *pb.ConfigValue {
	types := map[string]pb.ConfigValueType{"boolean": pb.ConfigValueType_CONFIG_VALUE_TYPE_BOOLEAN, "number": pb.ConfigValueType_CONFIG_VALUE_TYPE_NUMBER, "string": pb.ConfigValueType_CONFIG_VALUE_TYPE_STRING}
	return &pb.ConfigValue{Type: types[value.Type], Value: value.Value}
}
func localConfig(value *pb.ConfigValue) (build.ConfigValue, error) {
	types := map[pb.ConfigValueType]string{pb.ConfigValueType_CONFIG_VALUE_TYPE_BOOLEAN: "boolean", pb.ConfigValueType_CONFIG_VALUE_TYPE_NUMBER: "number", pb.ConfigValueType_CONFIG_VALUE_TYPE_STRING: "string"}
	kind, ok := types[value.GetType()]
	if !ok {
		return build.ConfigValue{}, fmt.Errorf("неизвестный тип условия")
	}
	return build.ConfigValue{Type: kind, Value: value.GetValue()}, nil
}

func wireModel(m build.Model) *pb.BuildModel {
	result := &pb.BuildModel{TargetVersion: m.TargetVersion, TreeVersion: m.TreeVersion, ClassName: m.Class, ClassId: uint32(m.ClassID), Ascendancy: m.Ascendancy,
		AscendancyId: uint32(m.AscendancyID), Level: uint32(m.Level), MainGroup: uint32(m.MainGroup), WeaponSet: uint32(m.WeaponSet), ActiveSets: wireSets(m.ActiveSets), Slots: map[string]uint32{}, Config: map[string]*pb.ConfigValue{}}
	for _, group := range m.Groups {
		out := &pb.SkillGroup{Id: uint32(group.ID), Label: group.Label, Enabled: group.Enabled, MainSkill: uint32(group.MainSkill)}
		for _, gem := range group.Gems {
			out.Gems = append(out.Gems, &pb.BuildGem{Id: gem.ID, Name: gem.Name, SkillId: gem.SkillID, VariantId: gem.VariantID, Level: uint32(gem.Level), Quality: uint32(gem.Quality), Enabled: gem.Enabled, StatSet: uint32(gem.StatSet), Corrupted: gem.Corrupted})
		}
		result.Groups = append(result.Groups, out)
	}
	for _, item := range m.Items {
		result.Items = append(result.Items, &pb.BuildItem{Id: uint32(item.ID), Raw: item.Raw})
	}
	for name, id := range m.Slots {
		result.Slots[name] = uint32(id)
	}
	for _, id := range m.Nodes {
		result.Nodes = append(result.Nodes, uint32(id))
	}
	for name, value := range m.Config {
		result.Config[name] = wireConfig(value)
	}
	for _, issue := range m.Issues {
		result.Issues = append(result.Issues, wireIssue(issue))
	}
	return result
}

func wireResult(value calculation.Result) *pb.CalculationResult {
	states := map[string]pb.CalculationStatus{"complete": pb.CalculationStatus_CALCULATION_STATUS_COMPLETE, "partial": pb.CalculationStatus_CALCULATION_STATUS_PARTIAL, "unsupported": pb.CalculationStatus_CALCULATION_STATUS_UNSUPPORTED}
	result := &pb.CalculationResult{Status: states[value.Status], DataId: value.DataID, DataSha256: value.DataSHA256, PobCommit: value.PoBCommit, InputSha256: value.InputSHA256, Skill: value.Skill, Mode: value.Mode, Stats: value.Stats}
	for _, mod := range value.Modifiers {
		result.Modifiers = append(result.Modifiers, &pb.AppliedModifier{Stat: mod.Stat, Kind: mod.Kind, Value: mod.Value, Source: mod.Source})
	}
	for _, issue := range value.Issues {
		result.Issues = append(result.Issues, wireIssue(issue))
	}
	return result
}

func localResult(value *pb.CalculationResult) (calculation.Result, error) {
	states := map[pb.CalculationStatus]string{pb.CalculationStatus_CALCULATION_STATUS_COMPLETE: "complete", pb.CalculationStatus_CALCULATION_STATUS_PARTIAL: "partial", pb.CalculationStatus_CALCULATION_STATUS_UNSUPPORTED: "unsupported"}
	state, ok := states[value.GetStatus()]
	if !ok {
		return calculation.Result{}, fmt.Errorf("неизвестный статус расчёта")
	}
	result := calculation.Result{Status: state, DataID: value.GetDataId(), DataSHA256: value.GetDataSha256(), PoBCommit: value.GetPobCommit(), InputSHA256: value.GetInputSha256(), Skill: value.GetSkill(), Mode: value.GetMode(), Stats: value.GetStats(), Modifiers: []calculation.Modifier{}, Issues: []build.Issue{}}
	if state == "unsupported" && len(result.Stats) == 0 {
		result.Stats = nil
	}
	for _, mod := range value.GetModifiers() {
		result.Modifiers = append(result.Modifiers, calculation.Modifier{Stat: mod.GetStat(), Kind: mod.GetKind(), Value: mod.GetValue(), Source: mod.GetSource()})
	}
	for _, issue := range value.GetIssues() {
		result.Issues = append(result.Issues, localIssue(issue))
	}
	return result, nil
}

func snapshot(value domain.Snapshot) (*pb.SessionSnapshot, error) {
	view, err := application.ViewSnapshot(value)
	if err != nil {
		return nil, err
	}
	result := &pb.SessionSnapshot{SessionId: value.SessionID, BuildName: value.BuildName, Revision: value.Revision}
	if view.Build != nil {
		result.Build = &pb.BuildDocument{Xml: view.Build.XML, InputSha256: view.Build.InputSHA256, InputRevision: view.Build.InputRevision, Model: wireModel(view.Build.Model), CalculatedFromRevision: view.Build.CalculatedFromRevision}
		if view.Build.Calculation != nil {
			result.Build.Calculation = wireResult(*view.Build.Calculation)
		}
	}
	return result, nil
}

func fromSnapshot(value *pb.SessionSnapshot) (domain.Snapshot, error) {
	result := domain.Snapshot{SessionID: value.GetSessionId(), BuildName: value.GetBuildName(), Revision: value.GetRevision()}
	if document := value.GetBuild(); document != nil {
		result.BuildXML = document.GetXml()
		result.InputSHA256 = document.GetInputSha256()
		result.BuildRevision = document.GetInputRevision()
		result.CalculatedFromRevision = document.GetCalculatedFromRevision()
		if document.GetCalculation() != nil {
			calculated, err := localResult(document.GetCalculation())
			if err != nil {
				return domain.Snapshot{}, err
			}
			encoded, err := json.Marshal(calculated)
			if err != nil {
				return domain.Snapshot{}, err
			}
			result.CalculationJSON = string(encoded)
		}
	}
	if _, err := application.ViewSnapshot(result); err != nil {
		return domain.Snapshot{}, err
	}
	return result, nil
}

func wireEdit(value *build.Edit) (*pb.BuildEdit, error) {
	if value == nil {
		return nil, nil
	}
	if err := build.ValidateEdit(*value); err != nil {
		return nil, err
	}
	if err := checkWireIntegers(value); err != nil {
		return nil, err
	}
	result := &pb.BuildEdit{}
	switch {
	case value.Level != nil:
		result.Operation = &pb.BuildEdit_Level{Level: &pb.LevelEdit{Level: uint32(value.Level.Level)}}
	case value.Config != nil:
		out := &pb.ConfigEdit{Name: value.Config.Name}
		if value.Config.Value != nil {
			out.Value = wireConfig(*value.Config.Value)
		}
		result.Operation = &pb.BuildEdit_Config{Config: out}
	case value.Gems != nil:
		out := &pb.GemsEdit{GroupId: uint32(value.Gems.GroupID), Label: value.Gems.Label, Enabled: value.Gems.Enabled, Delete: value.Gems.Delete}
		for _, gem := range value.Gems.Gems {
			out.Gems = append(out.Gems, &pb.GemInput{Id: gem.ID, Level: uint32(gem.Level), Quality: uint32(gem.Quality), Enabled: gem.Enabled})
		}
		result.Operation = &pb.BuildEdit_Gems{Gems: out}
	case value.Equipment != nil:
		result.Operation = &pb.BuildEdit_Equipment{Equipment: &pb.EquipmentEdit{Slot: value.Equipment.Slot, Raw: value.Equipment.Raw}}
	case value.Passives != nil:
		out := &pb.PassivesEdit{}
		for _, id := range value.Passives.Nodes {
			out.Nodes = append(out.Nodes, uint32(id))
		}
		result.Operation = &pb.BuildEdit_Passives{Passives: out}
	case value.Selection != nil:
		out := &pb.SelectionEdit{MainGroup: uint32(value.Selection.MainGroup), WeaponSet: uint32(value.Selection.WeaponSet)}
		if value.Selection.Sets != nil {
			out.Sets = wireSets(*value.Selection.Sets)
		}
		result.Operation = &pb.BuildEdit_Selection{Selection: out}
	case value.AddSupport != nil:
		edit := value.AddSupport
		result.Operation = &pb.BuildEdit_AddSupport{AddSupport: &pb.AddSupportEdit{
			SupportId: edit.SupportID, Level: uint32(edit.Level), Quality: uint32(edit.Quality),
		}}
	}
	return result, nil
}

func checkWireIntegers(value *build.Edit) error {
	var values []int
	if v := value.Level; v != nil {
		values = append(values, v.Level)
	}
	if v := value.Gems; v != nil {
		values = append(values, v.GroupID)
		for _, gem := range v.Gems {
			values = append(values, gem.Level, gem.Quality)
		}
	}
	if v := value.Passives; v != nil {
		values = append(values, v.Nodes...)
	}
	if v := value.AddSupport; v != nil {
		values = append(values, v.Level, v.Quality)
	}
	if v := value.Selection; v != nil {
		values = append(values, v.MainGroup, v.WeaponSet)
		if v.Sets != nil {
			values = append(values, v.Sets.Skills, v.Sets.Items, v.Sets.Tree, v.Sets.Config)
		}
	}
	for _, number := range values {
		if number < 0 || uint64(number) > math.MaxUint32 {
			return fmt.Errorf("целое значение вне диапазона Protobuf")
		}
	}
	return nil
}

func localEdit(value *pb.BuildEdit) (*build.Edit, error) {
	if value == nil {
		return nil, nil
	}
	result := &build.Edit{}
	switch operation := value.GetOperation().(type) {
	case *pb.BuildEdit_Level:
		result.Level = &build.LevelEdit{Level: int(operation.Level.GetLevel())}
	case *pb.BuildEdit_Config:
		result.Config = &build.ConfigEdit{Name: operation.Config.GetName()}
		if operation.Config.GetValue() != nil {
			local, err := localConfig(operation.Config.GetValue())
			if err != nil {
				return nil, err
			}
			result.Config.Value = &local
		}
	case *pb.BuildEdit_Gems:
		input := operation.Gems
		result.Gems = &build.GemsEdit{GroupID: int(input.GetGroupId()), Label: input.GetLabel(), Enabled: input.GetEnabled(), Delete: input.GetDelete(), Gems: []build.GemInput{}}
		for _, gem := range input.GetGems() {
			result.Gems.Gems = append(result.Gems.Gems, build.GemInput{ID: gem.GetId(), Level: int(gem.GetLevel()), Quality: int(gem.GetQuality()), Enabled: gem.GetEnabled()})
		}
	case *pb.BuildEdit_Equipment:
		result.Equipment = &build.EquipmentEdit{Slot: operation.Equipment.GetSlot(), Raw: operation.Equipment.GetRaw()}
	case *pb.BuildEdit_Passives:
		result.Passives = &build.PassivesEdit{Nodes: []int{}}
		for _, id := range operation.Passives.GetNodes() {
			result.Passives.Nodes = append(result.Passives.Nodes, int(id))
		}
	case *pb.BuildEdit_Selection:
		input := operation.Selection
		result.Selection = &build.SelectionEdit{MainGroup: int(input.GetMainGroup()), WeaponSet: int(input.GetWeaponSet())}
		if input.GetSets() != nil {
			sets := localSets(input.GetSets())
			result.Selection.Sets = &sets
		}
	case *pb.BuildEdit_AddSupport:
		edit := operation.AddSupport
		result.AddSupport = &build.AddSupportEdit{
			SupportID: edit.GetSupportId(), Level: int(edit.GetLevel()), Quality: int(edit.GetQuality()),
		}
	default:
		return nil, fmt.Errorf("не указана операция правки")
	}
	return result, nil
}
