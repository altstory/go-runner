package runner

import (
	"os"
	"path"
	"testing"

	"github.com/huandu/go-assert"
)

type FooConfig struct {
	Bar int `config:"bar"`
}

func TestLoadConfig(t *testing.T) {
	var foo *FooConfig
	var notExist *FooConfig
	a := assert.New(t)

	cwd, err := os.Getwd()
	a.NilError(err)
	defer os.Chdir(cwd)
	a.NilError(os.Chdir(path.Join(cwd, "internal", "testdata")))

	defer func() {
		configHandlers = nil
	}()

	configHandlers = nil
	LoadConfig("foo", &foo)
	LoadConfig("not_exist", &notExist)
	a.Equal(run(), ExitCodeOK)

	a.Equal(foo, &FooConfig{
		Bar: 123,
	})
	a.Equal(notExist, nil)
}
