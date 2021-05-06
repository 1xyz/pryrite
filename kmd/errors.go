package kmd

// FlagError is the kind of error raised in flag processing
type FlagError struct{ Err error }

func (fe FlagError) Error() string { return fe.Err.Error() }
func (fe FlagError) Unwrap() error { return fe.Err }
