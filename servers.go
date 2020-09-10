package runner

import (
	"context"
	"sync"
)

var serverHandlers []handler

// AddServer 注册一个自启动的服务。
// 这个 handler 执行之后必须保持阻塞，直到停止服务为止才应该返回。
func AddServer(section string, handler Handler) {
	const skip = 1
	h, err := parseHandler(section, handler)

	if err != nil {
		h = makeErrorHandler(skip, err)
	}

	serverHandlers = append(serverHandlers, h)
}

func runServers(ctx context.Context) int {
	if len(serverHandlers) == 0 {
		return ExitCodeOK
	}

	sz := len(serverHandlers)
	codes := make([]int, sz)
	wg := sync.WaitGroup{}
	wg.Add(sz)

	for i, h := range serverHandlers {
		go func(ctx context.Context, idx int, h handler) {
			defer wg.Done()
			codes[idx] = h.Call(ctx)
		}(ctx, i, h)
	}

	wg.Wait()

	for _, code := range codes {
		if code != 0 {
			return code
		}
	}

	return ExitCodeOK
}
