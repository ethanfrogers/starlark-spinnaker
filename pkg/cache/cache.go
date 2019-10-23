package cache

import (
	"fmt"
	"go.starlark.net/starlark"
	"io/ioutil"
	"os"
	"path"
)

// ModuleLoader is implemented as an interface to support multiple
// possible implementations in the future
// possibilities include in-memory, local file, remote file, etc
type ModuleLoader interface {
	// Handles determines whether the module pattern can be handled
	// by this ModuleLoader
	Handles(modulePath string) (bool, error)
	// Load handles loading the module and returning the source
	Load(modulePath string) ([]byte, error)
}

type LocalModuleLoader struct {}

func (l *LocalModuleLoader) Handles(modulePath string) (bool, error) {
	// local modules are the only thing supported for now we we'll just
	// return true
	return true, nil
}

func (l *LocalModuleLoader) Load(modulePath string) ([]byte, error) {
	wd, _ := os.Getwd()
	pth := path.Join(wd, modulePath)
	return ioutil.ReadFile(pth)
}

type cacheEntry struct {
	globals starlark.StringDict
	err error
}


type Loader struct {
	cache map[string]*cacheEntry
	moduleLoaders []ModuleLoader
	predeclaredModules starlark.StringDict
}

type LoaderOptionFunc func(l *Loader)

func WithModuleLoaders(moduleLoaders ...ModuleLoader) LoaderOptionFunc {
	return func (l *Loader) {
		l.moduleLoaders = moduleLoaders
	}
}

func WithPredeclaredModules(modules starlark.StringDict) LoaderOptionFunc {
	return func(l *Loader) {
		l.predeclaredModules = modules
	}
}

func NewLoader(opts ...LoaderOptionFunc) *Loader {
	l := &Loader{cache: make(map[string]*cacheEntry)}
	for _, o := range opts {
		o(l)
	}
	return l
}

func (l *Loader) addCacheEntry(pth string, globals starlark.StringDict, err error) *cacheEntry {
	l.cache[pth] = &cacheEntry{ globals, err}
	return l.cache[pth]
}

func (l *Loader) Load(_ *starlark.Thread, modulePath string) (starlark.StringDict, error) {
	e, ok := l.cache[modulePath]

	if e == nil {
		if ok {
			return nil, fmt.Errorf("cycle in load graph for module %s", modulePath)
		}

		l.cache[modulePath] = nil
		var moduleLoader ModuleLoader
		for _, ml := range l.moduleLoaders {
			handles, err := ml.Handles(modulePath)
			if err != nil {
				// log error
			}

			if handles {
				moduleLoader = ml
				break
			}
		}
		source, err := moduleLoader.Load(modulePath)
		if err != nil {
			l.addCacheEntry(modulePath, nil, err)
			return nil, err
		}
		thread := starlark.Thread{ Load: l.Load }
		globals, err := starlark.ExecFile(&thread, modulePath, source, l.predeclaredModules)
		e = l.addCacheEntry(modulePath, globals, err)
	}

	return e.globals, e.err
}


