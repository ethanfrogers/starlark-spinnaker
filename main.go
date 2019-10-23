package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/ethanfrogers/starlark-spinnaker/pkg/cache"
	"github.com/ethanfrogers/starlark-spinnaker/pkg/writer"
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


	starlarkLoader := cache.NewLoader(
		cache.WithModuleLoaders(&cache.LocalModuleLoader{}),
		cache.WithPredeclaredModules(predeclaredModules()))

	thread := starlark.Thread{
		Name: "spinlark",
		Load: starlarkLoader.Load,
	}

	globals, err := starlarkLoader.Load(&thread, *mainConfig)

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
	if err := writer.ToJson(buf, v); err != nil {
		panic(err)
	}

	fmt.Println(buf.String())

}
