package runner

import (
	"os"
	"testing"

	"github.com/huandu/go-assert"
)

func TestMeta(t *testing.T) {
	a := assert.New(t)

	cwd, err := os.Getwd()
	a.NilError(err)
	defer os.Chdir(cwd)
	a.NilError(os.Chdir("./internal/testdata"))

	parseMetaInfo()
	meta := Meta()
	a.Equal(meta, &MetaInfo{
		Project:     "go-runner",
		Namespace:   "altstory-framework",
		Env:         "production",
		Type:        "release",
		GitRefName:  "release-version",
		GitRevision: "12345678",
	})
}
