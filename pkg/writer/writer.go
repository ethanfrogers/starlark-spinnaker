package writer

import (
	"encoding/json"
	"fmt"
	"go.starlark.net/starlark"
	"io"
)

type OutputWriter interface {
	io.Writer
	io.ByteWriter
	io.StringWriter
}

// ToJson converts Starlark objects to their JSON representation
// writing the converted result to the output object
// copied from the Drone CI Starlark conversion plugin
func ToJson(out OutputWriter, v starlark.Value) error {
	switch v := v.(type) {
	case starlark.NoneType:
		out.WriteString("null")
	case starlark.Bool:
		fmt.Fprintf(out, "%t", v)
	case starlark.Int:
		out.WriteString(v.String())
	case starlark.Float:
		fmt.Fprintf(out, "%g", v)
	case starlark.String:
		s := string(v)
		if isQuoteSafe(s) {
			fmt.Fprintf(out, "%q", s)
		} else {
			data, _ := json.Marshal(s)
			out.Write(data)
		}
	case starlark.Indexable:
		out.WriteByte('[')
		for i, n := 0, starlark.Len(v); i < n; i++ {
			if i > 0 {
				out.WriteString(", ")
			}
			if err := ToJson(out, v.Index(i)); err != nil {
				return err
			}
		}
		out.WriteByte(']')
	case *starlark.Dict:
		out.WriteByte('{')
		for i, itemPair := range v.Items() {
			key := itemPair[0]
			value := itemPair[1]
			if i > 0 {
				out.WriteString(", ")
			}
			if err := ToJson(out, key); err != nil {
				return err
			}
			out.WriteString(": ")
			if err := ToJson(out, value); err != nil {
				return err
			}
		}
		out.WriteByte('}')
	default:
		return fmt.Errorf("value %s (type `%s') can't be converted to JSON", v.String(), v.Type())
	}
	return nil
}

func isQuoteSafe(s string) bool {
	for _, r := range s {
		if r < 0x20 || r >= 0x10000 {
			return false
		}
	}
	return true
}