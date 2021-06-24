package executor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Register struct {
	sync.Map
}

var (
	ErrUnsupportedContentType = errors.New("unsupported content-type")
	ErrExecInProgress         = errors.New("an execution is already in progress")
)

func NewRegister() (*Register, error) {
	r := &Register{}

	// do our best to kill/reap children when interrupted
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdown
		signal.Stop(shutdown)
		r.Cleanup()
		os.Exit(2)
	}()

	return r, nil
}

func (r *Register) Register(executor Executor) error {
	if entry, loaded := r.LoadOrStore(executor.ContentType(), executor); loaded {
		return fmt.Errorf("an entry for %s content-type already exists (%s)",
			executor.ContentType(), entry.(Executor).Name())
	}

	return nil
}

func (r *Register) Get(content []byte, contentType *ContentType) (Executor, error) {
	var executor Executor
	var ok bool

	// if a prompt is provided, we need to locate an executor of that type...
	isPrompted := false
	var start, stop int
	n, _ := fmt.Sscanf(contentType.Params["prompt-assign"], "%d:%d", &start, &stop)
	if n != 2 {
		n, _ = fmt.Sscanf(contentType.Params["prompt"], "%d:%d", &start, &stop)
		isPrompted = n == 2
	}
	if n == 2 {
		contentType = contentType.Clone()
		contentType.Subtype = string(content[start:stop])
		if isPrompted {
			// only match a prompted content-type request
			contentType.Params["prompt"] = contentType.Subtype
		}
	}

	// try to reuse one already running...
	r.Range(func(key interface{}, val interface{}) bool {
		ct := key.(*ContentType)
		if ct.ParentOf(contentType) {
			executor = val.(Executor)
			ok = true
			return false
		}
		return true
	})

	if !ok {
		// attempt to create a new one if one of our executors supports this type...
		for _, nf := range []func([]byte, *ContentType) (Executor, error){
			NewBashExecutor,
			NewPSQLExecutor,
		} {
			var err error
			executor, err = nf(content, contentType)
			if err != nil {
				if errors.Is(err, ErrUnsupportedContentType) {
					// keep looking...
					continue
				}
				return nil, err
			} else {
				// keep it around for future commands of this type
				r.Register(executor)
				ok = true
				break
			}
		}

		if !ok {
			return nil, fmt.Errorf("no executor found for content-type=%s", contentType)
		}
	}

	return executor, nil
}

func (r *Register) Execute(ctx context.Context, req *ExecRequest) *ExecResponse {
	executor, err := r.Get(req.Content, req.ContentType)
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
