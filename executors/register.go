package executor

import (
	"context"
	"fmt"
)

type Register map[ContentType]Executor

func NewRegister() (Register, error) {
	r := Register{}

	e, err := NewBashExecutor()
	if err != nil {
		return nil, err
	}

	r.Register(e)

	return r, nil
}

func (r Register) Register(executor Executor) error {
	for _, c := range executor.ContentTypes() {
		if entry, found := r[c]; found {
			return fmt.Errorf("an entry %s for content-type exists %v", c, entry.Name())
		}
		r[c] = executor
	}
	return nil
}

func (r Register) Get(contentType ContentType) (Executor, error) {
	executor, found := r[contentType]
	if !found {
		return nil, fmt.Errorf("no executor found for contentType=%s", contentType)
	}
	return executor, nil
}

func (r Register) Execute(ctx context.Context, req *ExecRequest) *ExecResponse {
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
