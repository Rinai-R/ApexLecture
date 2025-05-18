package goroutine

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/panjf2000/ants/v2"
)

func NewPool(size int) *ants.Pool {
	pool, err := ants.NewPool(
		size,
		ants.WithPreAlloc(true),
		ants.WithNonblocking(true),
		ants.WithPanicHandler(func(err interface{}) {
			klog.Debug("panic: %v\n", err)
		}),
	)
	if err != nil {
		klog.Fatal(err)
	}
	return pool
}
