package main

import (
	"context"

	"go.starlark.net/starlark"
)

func main() {
	ctx := context.Background()
	thread := &starlark.Thread{Name: "spinlark"}
	globals, err := starlark.ExecFile(thread, "main.sky", nil, nil)

	if err != nil {
		panic(err)
	}

}
