package graph

// HttpError is the kind of error raised in graph processing
type HttpError struct {
	Err      error
	HTTPCode int
}

func (ge HttpError) Error() string { return ge.Err.Error() }
func (ge HttpError) Unwrap() error { return ge.Err }

type Error struct {
	Context string
	Err     error
}

func (ge Error) Error() string { return ge.Err.Error() }
func (ge Error) Unwrap() error { return ge.Err }
