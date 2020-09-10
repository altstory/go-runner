package runner

import (
	"context"
	"errors"
	"os"
	"path"
	"testing"

	"github.com/huandu/go-assert"
)

func TestOnStart(t *testing.T) {
	touched := false
	a := assert.New(t)

	cwd, err := os.Getwd()
	a.NilError(err)
	defer os.Chdir(cwd)
	a.NilError(os.Chdir(path.Join(cwd, "internal", "testdata")))

	defer func() {
		onStartHandlers = nil
	}()

	onStartHandlers = nil
	OnStart(func(ctx context.Context) error {
		touched = true
		return nil
	})
	a.Equal(run(), ExitCodeOK)
	a.Assert(touched)

	onStartHandlers = nil
	OnStart(func(ctx context.Context) error {
		return errors.New("error")
	})
	a.Equal(run(), ExitCodeHandlerError)
}
