package kernel

import (
	"context"
	"fmt"
)

type Register map[ContentType]Kernel

func (r Register) Register(kernel Kernel) error {
	for _, c := range kernel.ContentTypes() {
		if entry, found := r[c]; found {
			return fmt.Errorf("an entry %s for content-type exists %v", c, entry.Name())
		}
		r[c] = kernel
	}
	return nil
}

func (r Register) Get(contentType ContentType) (Kernel, error) {
	kernel, found := r[contentType]
	if !found {
		return nil, fmt.Errorf("no kernel found for contentType=%s", contentType)
	}
	return kernel, nil
}

func (r Register) Execute(ctx context.Context, req *ExecRequest) *ExecResponse {
	contentType := req.ContentType
	if contentType == "" {
		contentType = Shell
	}
	k, err := r.Get(contentType)
	if err != nil {
		return &ExecResponse{
			Hdr:     &ResponseHdr{RequestID: "hello"},
			Err:     nil,
			Content: []byte{},
		}
	}
	return k.Execute(ctx, req)
}
