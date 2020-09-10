package runner

import (
	"context"
)

var onExitHandlers handlers

// OnExit 将 handler 注册到 runner 的启动列表里面。
func OnExit(handler func(ctx context.Context)) {
	if handler == nil {
		return
	}

	onExitHandlers = append(onExitHandlers, func(ctx context.Context) int {
		handler(ctx)
		return ExitCodeOK
	})
}

func runExitHandlers(ctx context.Context) int {
	return onExitHandlers.Call(ctx)
}
