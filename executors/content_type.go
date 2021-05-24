package executor

type ContentType string

const (
	Empty ContentType = ""
	Shell ContentType = "text/shell"
	Bash  ContentType = "text/bash"
)
