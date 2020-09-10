package runner

import (
	"context"
	"errors"
	"os"
	"path"
	"testing"

	"github.com/altstory/go-log"
	"github.com/huandu/go-assert"
)

type testFoo struct {
	Bar int `config:"bar"`
}

type testConfig struct {
	Log log.Config `config:"log"`
	Foo testFoo    `config:"foo"`
}

func TestRun(t *testing.T) {
	a := assert.New(t)
	cases := []struct {
		Handler  Handler
		Cwd      string
		ExitCode int
	}{
		{
			func() {},
			"./internal/testdata",
			ExitCodeInvalidHandler,
		},
		{
			func(ctx context.Context, c *testConfig) {},
			"",
			ExitCodeInvalidConfig,
		},
		{
			func(ctx context.Context, c *testConfig) { log.Fatalf(ctx, "intended") },
			"./internal/testdata",
			ExitCodeHandlerError,
		},
		{
			func(ctx context.Context, c *testConfig) error { return errors.New("intended") },
			"./internal/testdata",
			ExitCodeHandlerError,
		},
		{
			func(ctx context.Context, c *testConfig) {
				a.Equal(c, &testConfig{
					Log: log.Config{
						LogPath:  "./log/test.log",
						LogLevel: "debug",
					},
					Foo: testFoo{
						Bar: 123,
					},
				})
			},
			"./internal/testdata",
			ExitCodeOK,
		},
	}

	cwd, err := os.Getwd()
	a.NilError(err)
	defer os.Chdir(cwd)
	defer func() {
		clientHandlers = nil
	}()

	for _, c := range cases {
		if c.Cwd != "" {
			a.NilError(os.Chdir(path.Join(cwd, c.Cwd)))
		} else {
			a.NilError(os.Chdir(cwd))
		}

		clientHandlers = nil
		AddClient("", c.Handler)
		a.Equal(run(), c.ExitCode)
	}
}

func TestRunWithExtConfig(t *testing.T) {
	a := assert.New(t)

	cwd, err := os.Getwd()
	a.NilError(err)
	defer os.Chdir(cwd)
	defer func() {
		clientHandlers = nil
	}()
	a.NilError(os.Chdir("./internal/testdata"))

	a.NilError(os.Setenv(envRunnerExtConfig, "./conf/service-ext.conf"))
	AddClient("", func(ctx context.Context, c *testConfig) {
		a.Equal(c, &testConfig{
			Log: log.Config{
				LogPath:  "./log/test.log",
				LogLevel: "info",
			},
			Foo: testFoo{
				Bar: 0,
			},
		})
	})
	a.Equal(run(), ExitCodeOK)
}
