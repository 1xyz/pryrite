package executor

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"

	"github.com/1xyz/pryrite/app"
	"github.com/1xyz/pryrite/tools"
)

type Register struct {
	sync.Map
}

var (
	ErrUnsupportedContentType = errors.New("unsupported content-type")
	ErrExecInProgress         = errors.New("an execution is already in progress")

	disablePTY = false
)

func NewRegister() (*Register, error) {
	r := &Register{}

	// do our best to kill/reap children before exiting
	app.AtExit(r.Cleanup)

	return r, nil
}

func DisablePTY() {
	disablePTY = true
}

func (r *Register) Register(executor Executor) error {
	if entry, loaded := r.LoadOrStore(executor.ContentType(), executor); loaded {
		return fmt.Errorf("an entry for %s content-type already exists (%s)",
			executor.ContentType(), entry.(Executor).Name())
	}

	tools.Log.Debug().Str("contentType", executor.ContentType().String()).Str("name", executor.Name()).
		Msg("Registered executor")

	return nil
}

func (r *Register) Get(content []byte, contentType *ContentType) (Executor, error) {
	var executor Executor
	var ok bool

	// convert any a:b positions into their string counterparts from the content
	contentType = translatePositions(content, contentType)

	// if a prompt is provided, we need to locate an executor of that type...
	prompt := contentType.Params["prompt-assign"]
	isAssigned := prompt != ""
	isPrompted := false
	if !isAssigned {
		prompt = contentType.Params["prompt"]
		isPrompted = prompt != ""
	}
	if isAssigned || isPrompted {
		contentType = contentType.Clone()
		contentType.Subtype = prompt
		if isPrompted {
			// only match a prompted content-type request
			contentType.Params["prompt"] = contentType.Subtype
		}
	}

	if !isAssigned {
		// try to reuse one already running, but only if it is NOT a prompt assignment...
		r.Range(func(key interface{}, val interface{}) bool {
			ct := key.(*ContentType)
			var requiredKeys []string
			if isPrompted {
				requiredKeys = []string{"prompt"}
			}
			tools.Trace("register", "parent-of (running)", ct.String(), contentType.String(), requiredKeys)
			if ct.ParentOf(contentType, requiredKeys) {
				executor = val.(Executor)
				ok = true
				return false
			}
			return true
		})

		if !ok && isPrompted {
			return nil, fmt.Errorf("no running prompt found for content-type=%s", contentType)
		}
	}

	if !ok {
		// attempt to create a new one if one of our executors supports this type...
		for _, nf := range []func([]byte, *ContentType) (Executor, error){
			NewWinBashExecutor,
			NewBashExecutor,
			NewPSQLExecutor,
			NewMySQLExecutor,
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
	tools.Trace("register", "execute requested", req.ContentType.String(), string(req.Content))
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

// Be very strict about what we "find" as a postion
var positionRE = regexp.MustCompile(`^\s*(\d+):(\d+)?\s*$`)

func translatePositions(content []byte, contentType *ContentType) *ContentType {
	var clone *ContentType

	for key, val := range contentType.Params {
		results := positionRE.FindAllStringSubmatch(val, -1)
		if len(results) != 1 {
			continue
		}

		var err error
		start := -1
		stop := -1

		result := results[0]
		if result[2] == "" {
			start, err = strconv.Atoi(result[1])
			if err != nil {
				continue
			}
		} else {
			start, err = strconv.Atoi(result[1])
			if err != nil {
				continue
			}
			stop, err = strconv.Atoi(result[2])
			if err != nil {
				continue
			}
		}

		if clone == nil {
			clone = contentType.Clone()
		}

		var newVal string
		if stop == -1 {
			newVal = string(content[start:])
		} else {
			newVal = string(content[start:stop])
		}

		clone.Params[key] = newVal
	}

	if clone != nil {
		tools.Trace("register", "translated content-type", contentType)
		return clone
	}

	return contentType
}
