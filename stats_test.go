package runner

import (
	"testing"

	"github.com/altstory/go-log"
	"github.com/huandu/go-assert"
)

func TestStats(t *testing.T) {
	a := assert.New(t)

	var stats *Stats

	// stats 是 nil，什么也不会发生。
	stats.Add("foo", 2)
	stats.Set("bar", 3)
	a.Assert(stats.Info() == nil)

	stats = &Stats{}
	stats.Add("foo", 2)
	stats.Add("foo", 4)
	stats.Set("foo", 3)
	stats.Add("foo", 8)
	a.Equal(stats.Info(), []log.Info{
		{
			Key:   "foo",
			Value: 11,
		},
	})
}
