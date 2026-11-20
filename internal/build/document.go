// Package build хранит модель сборки и XML независимо от транспортов.
package build

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"slices"
	"strings"
	"unicode/utf8"
)

const MaxXMLSize = 1 << 20

type part struct {
	element *element
	text    string
	comment []byte
}

type element struct {
	name  xml.Name
	attrs []xml.Attr
	parts []part
}

type Document struct{ root *element }

func Parse(data []byte) (*Document, error) {
	if len(data) == 0 || len(data) > MaxXMLSize || !utf8.Valid(data) {
		return nil, fmt.Errorf("XML: нужен корректный UTF-8 размером до 1 МиБ")
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var root *element
	var stack []*element
	count := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("повреждённый XML: %w", err)
		}
		count++
		if count > 20000 {
			return nil, fmt.Errorf("слишком много элементов XML")
		}
		switch value := token.(type) {
		case xml.StartElement:
			if len(stack) >= 64 || len(value.Attr) > 128 {
				return nil, fmt.Errorf("слишком сложный XML")
			}
			node := &element{name: value.Name, attrs: slices.Clone(value.Attr)}
			if err := validateAttributes(node); err != nil {
				return nil, err
			}
			if len(stack) == 0 {
				if root != nil {
					return nil, fmt.Errorf("в XML несколько корней")
				}
				root = node
			} else {
				parent := stack[len(stack)-1]
				parent.parts = append(parent.parts, part{element: node})
			}
			stack = append(stack, node)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) == 0 {
				if strings.TrimSpace(string(value)) != "" {
					return nil, fmt.Errorf("текст вне корня XML")
				}
			} else {
				parent := stack[len(stack)-1]
				parent.parts = append(parent.parts, part{text: string(value)})
			}
		case xml.Comment:
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.parts = append(parent.parts, part{comment: slices.Clone([]byte(value))})
			}
		case xml.Directive:
			return nil, fmt.Errorf("директивы и DTD запрещены")
		case xml.ProcInst:
			if value.Target != "xml" || root != nil {
				return nil, fmt.Errorf("инструкции обработки XML запрещены")
			}
		}
	}
	if root == nil || root.name.Local != "PathOfBuilding2" || root.name.Space != "" {
		return nil, fmt.Errorf("ожидается корень PathOfBuilding2")
	}
	doc := &Document{root: root}
	if _, err := doc.Model(); err != nil {
		return nil, err
	}
	return doc, nil
}

func validateAttributes(node *element) error {
	if node.name.Space != "" {
		return fmt.Errorf("пространства имён XML пока не поддерживаются")
	}
	seen := make(map[xml.Name]bool)
	for _, attr := range node.attrs {
		if attr.Name.Space != "" || attr.Name.Local == "xmlns" {
			return fmt.Errorf("пространства имён XML пока не поддерживаются")
		}
		if seen[attr.Name] {
			return fmt.Errorf("повтор атрибута XML %s", attr.Name.Local)
		}
		seen[attr.Name] = true
	}
	return nil
}

func (n *element) attr(name, fallback string) string {
	for _, a := range n.attrs {
		if a.Name.Local == name && a.Name.Space == "" {
			return a.Value
		}
	}
	return fallback
}

func (n *element) set(name, value string) {
	for i := range n.attrs {
		if n.attrs[i].Name.Local == name && n.attrs[i].Name.Space == "" {
			n.attrs[i].Value = value
			return
		}
	}
	n.attrs = append(n.attrs, xml.Attr{Name: xml.Name{Local: name}, Value: value})
}

func (n *element) children(name string) []*element {
	var result []*element
	for _, p := range n.parts {
		if p.element != nil && p.element.name.Local == name && p.element.name.Space == "" {
			result = append(result, p.element)
		}
	}
	return result
}

func (n *element) child(name string) (*element, error) {
	values := n.children(name)
	if len(values) > 1 {
		return nil, fmt.Errorf("повтор секции XML %s", name)
	}
	if len(values) == 0 {
		return nil, nil
	}
	return values[0], nil
}

func (n *element) content() string {
	var result strings.Builder
	for _, p := range n.parts {
		result.WriteString(p.text)
	}
	return result.String()
}

func encodeElement(encoder *xml.Encoder, node *element) error {
	attrs := slices.Clone(node.attrs)
	slices.SortFunc(attrs, func(a, b xml.Attr) int {
		if result := strings.Compare(a.Name.Space, b.Name.Space); result != 0 {
			return result
		}
		return strings.Compare(a.Name.Local, b.Name.Local)
	})
	start := xml.StartElement{Name: node.name, Attr: attrs}
	if err := encoder.EncodeToken(start); err != nil {
		return err
	}
	hasChildren := false
	for _, p := range node.parts {
		hasChildren = hasChildren || p.element != nil
	}
	for _, p := range node.parts {
		if p.element != nil {
			if err := encodeElement(encoder, p.element); err != nil {
				return err
			}
		} else if p.comment != nil {
			if err := encoder.EncodeToken(xml.Comment(p.comment)); err != nil {
				return err
			}
		} else if _, structural := layouts[node.name.Local]; !hasChildren || !structural || strings.TrimSpace(p.text) != "" {
			if err := encoder.EncodeToken(xml.CharData(p.text)); err != nil {
				return err
			}
		}
	}
	return encoder.EncodeToken(start.End())
}

func (d *Document) canonical() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := xml.NewEncoder(&buffer)
	if err := encodeElement(encoder, d.root); err != nil {
		return nil, err
	}
	if err := encoder.Flush(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (d *Document) Export() ([]byte, error) {
	content, err := d.canonical()
	if err != nil {
		return nil, err
	}
	if len(content)+len(xml.Header)+1 > MaxXMLSize {
		return nil, fmt.Errorf("экспортированный XML превышает 1 МиБ")
	}
	return append([]byte(xml.Header), append(content, '\n')...), nil
}

func (d *Document) InputSHA256() (string, error) {
	content, err := d.canonical()
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:]), nil
}
