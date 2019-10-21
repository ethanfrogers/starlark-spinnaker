package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func includedModules() starlark.StringDict {
	return starlark.StringDict{
		"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
	}
}

func predeclaredModules() (starlark.StringDict) {
	return starlark.StringDict{
		"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
	}
}

func main() {
	mainConfig := flag.String("config", "main.sky", "entrypoint filename")
	flag.Parse()

	type entry struct {
		globals starlark.StringDict
		err error
	}

	cache := make(map[string]*entry)

	var load func(_ *starlark.Thread, modulePath string) (starlark.StringDict, error)
	load = func(_ *starlark.Thread, modulePath string) (starlark.StringDict, error)  {
		e, ok := cache[modulePath]
		if e == nil {
			if ok {
				return nil, fmt.Errorf("cycle in load graph for module %s", modulePath)
			}

			cache[modulePath] = nil
			// read file from disk
			wd, _ := os.Getwd()
			pth := path.Join(wd, modulePath)
			moduleSource, err := ioutil.ReadFile(pth)
			if err != nil {
				cache[modulePath] = &entry{nil, err}
				return nil, err
			}
			thread := starlark.Thread{ Load: load }
			globals, err := starlark.ExecFile(&thread, modulePath, moduleSource, predeclaredModules())
			e = &entry{globals, err}
			cache[modulePath] = e
		}

		return e.globals, e.err
	}

	thread := starlark.Thread{
		Name: "spinlark",
		Load: load,
	}

	globals, err := load(&thread, *mainConfig)

	if err != nil {
		panic(err)
	}

	mainFunc, ok := globals["main"]
	if !ok {
		panic(errors.New("no main function found"))
	}

	v, err := starlark.Call(&thread, mainFunc, nil , nil)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	if err := write(buf, v); err != nil {
		panic(err)
	}

	fmt.Println(buf.String())

}


type writer interface {
	io.Writer
	io.ByteWriter
	io.StringWriter
}

func write(out writer, v starlark.Value) error {

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
			if err := write(out, v.Index(i)); err != nil {
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
			if err := write(out, key); err != nil {
				return err
			}
			out.WriteString(": ")
			if err := write(out, value); err != nil {
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
