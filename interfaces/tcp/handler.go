package tcp

import (
	"context"
	"net"
)


// 统一函数格式
type HandleFunc func(ctx context.Context, conn net.Conn)

// 方法清单
type Handler interface{
	Handle(ctx context.Context, conn net.Conn)
	Close() error
} 