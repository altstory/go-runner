package runner

import "context"

type contextStats struct{}

var keyContextStats contextStats

// WithStats 将 stats 追加到 ctx 中，并返回新的 ctx。
func WithStats(ctx context.Context, stats *Stats) context.Context {
	return context.WithValue(ctx, keyContextStats, stats)
}

// StatsFromContext 返回 ctx 中存放的 stats，如果之前没有设置 stats 则返回 nil。
// 由于 stats 做了 nil 兼容，所有函数在 stats == nil 时候依然可以正常调用，
// 所以业务代码永远不需要检查这个返回值是否为 nil 就可以正常使用。
func StatsFromContext(ctx context.Context) *Stats {
	v := ctx.Value(keyContextStats)

	if v == nil {
		return nil
	}

	return v.(*Stats)
}
