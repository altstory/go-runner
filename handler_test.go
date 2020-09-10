package runner

import (
	"context"
	"os"
	"testing"

	"github.com/huandu/go-assert"
)

type testHandlerConfig struct{ Foo int }
type testError interface{ error }

func testValid1(ctx context.Context)                                 {}
func testValid2(ctx context.Context) error                           { return nil }
func testValid3(ctx context.Context, c *testHandlerConfig)           {}
func testValid4(ctx context.Context, c *testHandlerConfig) error     { return nil }
func testValid5(ctx context.Context, c *testHandlerConfig) testError { return nil }

func TestParseValidHandler(t *testing.T) {
	handlers := []Handler{testValid1, testValid2, testValid3, testValid4, testValid5}

	for _, h := range handlers {
		_, err := parseHandler("", h)

		if err != nil {
			t.Fatalf("invalid handler. [err:%v]", err)
		}
	}
}

func testInvalid1()                                                             {}
func testInvalid2(c testHandlerConfig)                                          {}
func testInvalid3(c *testHandlerConfig) (i int, err error)                      { return }
func testInvalid4(c *testHandlerConfig) (i interface{})                         { return }
func testInvalid5(c *testHandlerConfig) (s string)                              { return }
func testInvalid6(ctx context.Context, c testHandlerConfig)                     {}
func testInvalid7(ctx context.Context, c *testHandlerConfig) (i int, err error) { return }
func testInvalid8(ctx context.Context, c *testHandlerConfig) (i interface{})    { return }
func testInvalid9(ctx context.Context, c *testHandlerConfig) (s string)         { return }

func TestParseInvalidHandler(t *testing.T) {
	handlers := []Handler{nil, 1, testInvalid1, testInvalid2, testInvalid3, testInvalid4, testInvalid5,
		testInvalid6, testInvalid7, testInvalid8, testInvalid9}

	for _, h := range handlers {
		_, err := parseHandler("", h)

		if err == nil {
			t.Fatalf("handler should be invalid.")
		}
	}
}

type fooConfig struct {
	Bar int `config:"bar"`
}

func TestUnmarshalConfig(t *testing.T) {
	a := assert.New(t)
	all := make([]*fooConfig, 4)

	old, err := os.Getwd()
	a.NilError(err)
	defer os.Chdir(old)
	a.NilError(os.Chdir("./internal/testdata"))

	AddServer("foo", func(ctx context.Context, config *fooConfig) error {
		all[0] = config
		return nil
	})
	AddClient("foo", func(ctx context.Context, config *fooConfig) error {
		all[1] = config
		return nil
	})
	AddServer("not.exist", func(ctx context.Context, config *fooConfig) error {
		all[2] = config
		return nil
	})
	AddClient("not_exist", func(ctx context.Context, config *fooConfig) error {
		all[3] = config
		return nil
	})
	a.Assert(run() == 0)

	a.Equal(all[0].Bar, 123)
	a.Equal(all[1].Bar, 123)
	a.Assert(all[2] == nil)
	a.Assert(all[3] == nil)
}
