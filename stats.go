package runner

import (
	"sync"

	"github.com/altstory/go-log"
)

// Stats 用于记录运行时的统计信息。
type Stats struct {
	mu   sync.Mutex
	data map[string]int
}

// Add 为一个 key 增加 value 的统计值。
func (stats *Stats) Add(key string, value int) {
	if stats == nil || key == "" {
		return
	}

	stats.mu.Lock()
	defer stats.mu.Unlock()

	if stats.data == nil {
		stats.data = make(map[string]int)
	}

	stats.data[key] += value
}

// Set 将 key 的统计值设置为 value。
func (stats *Stats) Set(key string, value int) {
	if stats == nil || key == "" {
		return
	}

	stats.mu.Lock()
	defer stats.mu.Unlock()

	if stats.data == nil {
		stats.data = make(map[string]int)
	}

	stats.data[key] = value
}

// Info 返回当前记录的所有的统计值，用于记录日志。
func (stats *Stats) Info() []log.Info {
	if stats == nil {
		return nil
	}

	stats.mu.Lock()
	defer stats.mu.Unlock()

	list := make([]log.Info, 0, len(stats.data))

	for k, v := range stats.data {
		list = append(list, log.Info{
			Key:   k,
			Value: v,
		})
	}

	return list
}
