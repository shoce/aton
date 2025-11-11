/*
kleine markup language
GoFmt GoBuildNull
GoRun
*/

package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

func Marshal(v interface{}) (string, error) {
	var buf bytes.Buffer
	err := encodeValue(&buf, v, 0)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func encodeValue(buf *bytes.Buffer, v interface{}, indent int) error {
	switch val := v.(type) {

	case nil:
		buf.WriteString("<nil>")

	case string:
		if strings.ContainsAny(val, " \t\n{}()<>") {
			buf.WriteString("[" + val + "]")
		} else {
			buf.WriteString(val)
		}

	case bool:
		if val {
			buf.WriteString("<true>")
		} else {
			buf.WriteString("<false>")
		}

	case int, int8, int16, int32, int64:
		buf.WriteString(fmt.Sprintf("<%d>", val))

	case uint, uint8, uint16, uint32, uint64:
		buf.WriteString(fmt.Sprintf("<%d>", val))

	case float32, float64:
		buf.WriteString(fmt.Sprintf("<%v>", val))

	case map[string]interface{}:
		buf.WriteString("{\n")
		for k, v2 := range val {
			writeIndent(buf, indent+2)
			if strings.ContainsAny(k, " \t\n{}()<>") {
				buf.WriteString("[" + k + "] ")
			} else {
				buf.WriteString(k + " ")
			}
			if err := encodeValue(buf, v2, indent+2); err != nil {
				return err
			}
			buf.WriteString("\n")
		}
		writeIndent(buf, indent)
		buf.WriteString("}")

	case []interface{}:
		buf.WriteString("(")
		for i, elem := range val {
			if i > 0 {
				buf.WriteString(" ")
			}
			if err := encodeValue(buf, elem, indent); err != nil {
				return err
			}
		}
		buf.WriteString(")")

	default:
		return fmt.Errorf("unsupported type: %T", v)

	}
	return nil
}

func writeIndent(buf *bytes.Buffer, n int) {
	for i := 0; i < n; i++ {
		buf.WriteByte(' ')
	}
}

func Unmarshal(data string) (interface{}, error) {
	tokens := tokenize(data)
	pos := 0
	return parseValue(tokens, &pos)
}

func tokenize(s string) []string {
	var tokens []string
	var buf bytes.Buffer
	inBracket := false

	for _, r := range s {
		switch {

		case r == '[':
			inBracket = true
			buf.Reset()

		case r == ']':
			tokens = append(tokens, buf.String())
			inBracket = false

		case inBracket:
			buf.WriteRune(r)

		case strings.ContainsRune("{}()<>\n\t ", r):
			if buf.Len() > 0 {
				tokens = append(tokens, buf.String())
				buf.Reset()
			}
			if strings.TrimSpace(string(r)) != "" {
				tokens = append(tokens, string(r))
			}

		default:
			buf.WriteRune(r)

		}
	}

	if buf.Len() > 0 {
		tokens = append(tokens, buf.String())
	}
	return tokens
}

func parseValue(tokens []string, pos *int) (interface{}, error) {
	if *pos >= len(tokens) {
		return nil, errors.New("unexpected end of input")
	}
	tok := tokens[*pos]
	*pos++

	switch tok {

	case "{":
		obj := make(map[string]interface{})
		for *pos < len(tokens) && tokens[*pos] != "}" {
			key := tokens[*pos]
			*pos++
			val, err := parseValue(tokens, pos)
			if err != nil {
				return nil, err
			}
			obj[key] = val
		}
		if *pos >= len(tokens) || tokens[*pos] != "}" {
			return nil, errors.New("missing closing }")
		}
		*pos++
		return obj, nil

	case "(":
		var arr []interface{}
		for *pos < len(tokens) && tokens[*pos] != ")" {
			val, err := parseValue(tokens, pos)
			if err != nil {
				return nil, err
			}
			arr = append(arr, val)
		}
		if *pos >= len(tokens) || tokens[*pos] != ")" {
			return nil, errors.New("missing closing )")
		}
		*pos++
		return arr, nil

	default:

		// numeral
		if strings.HasPrefix(tok, "<") && strings.HasSuffix(tok, ">") {
			inner := strings.Trim(tok, "<>")
			switch inner {

			case "true":
				return true, nil

			case "false":
				return false, nil

			case "nil":
				return nil, nil

			}
			if i, err := strconv.ParseInt(inner, 10, 64); err == nil {
				return i, nil
			}
			if u, err := strconv.ParseUint(inner, 10, 64); err == nil {
				return u, nil
			}
			if f, err := strconv.ParseFloat(inner, 64); err == nil {
				return f, nil
			}
			return nil, fmt.Errorf("invalid numeral: %s", inner)
		}

		// string
		return tok, nil
	}
}

func main() {
	data := map[string]any{
		"name":    "Alice",
		"age":     30,
		"active":  true,
		"hobbies": []any{"reading", "gaming", "chess"},
		"address": map[string]any{
			"city":   "Wonderland",
			"street": "Rabbit Hole 42",
		},
		"notes": "Multi\nline\nnote",
		"nilv":  nil,
	}

	encoded, err := Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("=== Marshaled ===")
	fmt.Println(encoded)

	decoded, err := Unmarshal(encoded)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n=== Unmarshaled ===")
	fmt.Printf("%#v\n", decoded)
}
