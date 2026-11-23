package application

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func modelFields(doc *build.Document) (map[string]json.RawMessage, error) {
	model, err := doc.Model()
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

func changedBuildFields(previous string, current *build.Document) (string, error) {
	before, err := modelFieldsFromXML(previous)
	if err != nil {
		return "", err
	}
	after, err := modelFields(current)
	if err != nil {
		return "", err
	}
	fields := []string{"build.xml", "build.input_sha256", "build.input_revision", "calculation"}
	for name, value := range after {
		if !bytes.Equal(before[name], value) {
			fields = append(fields, "build.model."+name)
		}
	}
	for name := range before {
		if _, present := after[name]; !present {
			fields = append(fields, "build.model."+name)
		}
	}
	sort.Strings(fields)
	return strings.Join(fields, ","), nil
}

func modelFieldsFromXML(xml string) (map[string]json.RawMessage, error) {
	if xml == "" {
		return nil, nil
	}
	doc, err := build.Parse([]byte(xml))
	if err != nil {
		return nil, err
	}
	return modelFields(doc)
}
