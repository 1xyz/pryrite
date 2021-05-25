package kernel

type ContentType string

const (
	Empty ContentType = ""
	Shell ContentType = "shell"
	Bash  ContentType = "bash"
)
