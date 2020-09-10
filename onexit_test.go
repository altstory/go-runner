package runner

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/huandu/go-assert"
)

func TestOnExit(t *testing.T) {
	touched := false
	a := assert.New(t)

	cwd, err := os.Getwd()
	a.NilError(err)
	defer os.Chdir(cwd)
	a.NilError(os.Chdir(path.Join(cwd, "internal", "testdata")))

	defer func() {
		onExitHandlers = nil
	}()

	onExitHandlers = nil
	OnExit(func(ctx context.Context) {
		touched = true
	})
	a.Equal(run(), ExitCodeOK)
	a.Assert(touched)
}
