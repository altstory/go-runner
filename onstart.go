package runner

import (
	"context"

	"github.com/altstory/go-log"
)

var onStartHandlers handlers

// OnStart 将 handler 注册到 runner 的启动列表里面。
func OnStart(handler func(ctx context.Context) error) {
	if handler == nil {
		return
	}

	caller := findCaller(1)
	onStartHandlers = append(onStartHandlers, func(ctx context.Context) int {
		if err := handler(ctx); err != nil {
			log.Errorf(ctx, "err=%v||caller=%v||go-runner: fail to call handler", err, caller)
			return ExitCodeHandlerError
		}

		return ExitCodeOK
	})
}

func runStartHandlers(ctx context.Context) int {
	return onStartHandlers.Call(ctx)
}
