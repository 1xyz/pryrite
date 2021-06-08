package executor

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Register struct {
	sync.Map
}

func NewRegister() (*Register, error) {
	r := &Register{}

	e, err := NewBashExecutor()
	if err != nil {
		return nil, err
	}

	r.Register(e)

	return r, nil
}

func (r *Register) Register(executor Executor) error {
	for _, c := range executor.ContentTypes() {
		if entry, loaded := r.LoadOrStore(c, executor); loaded {
			return fmt.Errorf("an entry %s for content-type exists %v", c, entry.(Executor).Name())
		}
	}

	// do our best to kill/reap children when interrupted
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdown
		signal.Stop(shutdown)
		r.Cleanup()
	}()

	return nil
}

func (r *Register) Get(contentType ContentType) (Executor, error) {
	executor, ok := r.Load(contentType)
	if !ok {
		return nil, fmt.Errorf("no executor found for contentType=%s", contentType)
	}
	return executor.(Executor), nil
}

func (r *Register) Execute(ctx context.Context, req *ExecRequest) *ExecResponse {
	contentType := req.ContentType
	executor, err := r.Get(contentType)
	if err != nil {
		return &ExecResponse{
			Hdr: &ResponseHdr{RequestID: req.Hdr.ID},
			Err: err,
		}
	}

	return executor.Execute(ctx, req)
}

func (r *Register) Cleanup() {
	r.Range(func(_ interface{}, executor interface{}) bool {
		executor.(Executor).Cleanup()
		return true
	})
}
