package runner

import (
	"context"
)

var clientHandlers handlers

// AddClient 注册一个自注册的 client 工厂。
func AddClient(section string, handler Handler) {
	const skip = 1
	h, err := parseHandler(section, handler)

	if err != nil {
		h = makeErrorHandler(skip, err)
	}

	clientHandlers = append(clientHandlers, h)
}

func runClients(ctx context.Context) int {
	return clientHandlers.Call(ctx)
}
