/*
at object notation

GoFmt GoBuildNull
GoRun

https://github.com/shoce/aton/actions

*/

package aton

import (
	"fmt"
	"os"
	"strings"
)

const (
	SP   = " "
	SPAC = "    "
	TAB  = "\t"
	NL   = "\n"
)

func main() {

	data := map[string]interface{}{
		"version": 2,
		"models": []interface{}{
			map[string]interface{}{
				"name":         "my_transformation",
				"description":  "This model transforms raw data",
				"database":     "your_database",
				"schema":       "your_schema",
				"materialized": "table",
				"columns": []interface{}{
					map[string]interface{}{
						"name":        "id",
						"description": "A unique identifier",
						"primary_key": true,
					},
					map[string]interface{}{
						"name":        "name",
						"description": "The name of the item",
						"primary_key": false,
					},
				},
				"sql": "SELECT id, name FROM source_data",
			},
		},
	}

	fmt.Printf("%#v"+NL, data)

	if datatext, err := MarshalDocument(data); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR MarshalDocument %v"+NL, err)
		os.Exit(1)
	} else {
		fmt.Println(datatext)
	}

}

type Marshaler struct {
	indent string
	level  int
}

func NewMarshaler() *Marshaler {
	return &Marshaler{indent: TAB}
}

func (m *Marshaler) Marshal(v interface{}) (string, error) {
	var sb strings.Builder
	if err := m.marshalValue(&sb, v); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (m *Marshaler) marshalValue(sb *strings.Builder, v interface{}) error {
	switch val := v.(type) {

	case map[string]interface{}:

		if m.level > 0 {
			sb.WriteString("{")
			m.level++
		}

		dict, _ := v.(map[string]interface{})
		for key, value := range dict {
			sb.WriteString(NL)
			sb.WriteString(strings.Repeat(m.indent, m.level))
			sb.WriteString("@[")
			sb.WriteString(key)
			sb.WriteString("]")
			sb.WriteString(SP)
			if err := m.marshalValue(sb, value); err != nil {
				return err
			}
		}

		if m.level > 0 {
			m.level--
			sb.WriteString(NL)
			sb.WriteString(strings.Repeat(m.indent, m.level))
			sb.WriteString("}")
		}

	case []interface{}:

		sb.WriteString("(")
		m.level++

		list, _ := v.([]interface{})
		for _, item := range list {
			sb.WriteString(NL)
			sb.WriteString(strings.Repeat(m.indent, m.level))
			if err := m.marshalValue(sb, item); err != nil {
				return err
			}
		}

		m.level--
		sb.WriteString(NL)
		sb.WriteString(strings.Repeat(m.indent, m.level))
		sb.WriteString(")")

	case bool:

		sb.WriteString(fmt.Sprintf("<%t>", val))

	case int:

		sb.WriteString(fmt.Sprintf("<%d>", val))

	case string:

		sb.WriteString(fmt.Sprintf("[%s]", EscStr(val)))

	default:

		return fmt.Errorf("unsupported type %T", v)

	}
	return nil
}

func MarshalDocument(v interface{}) (string, error) {
	marshaler := NewMarshaler()
	result, err := marshaler.Marshal(v)
	if err != nil {
		return "", err
	}
	return result + NL, nil
}

func EscStr(text string) string {
	for _, c := range "\\]" {
		text = strings.ReplaceAll(text, string(c), "\\"+string(c))
	}
	return text
}
