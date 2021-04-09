package graph

type Kind string

const (
	Unknown       Kind = "Unknown"
	Command            = "Command"
	AsciiCast          = "AsciiCast"
	PageOpen           = "PageOpen"
	PageClose          = "PageClose"
	TextSelect         = "TextSelect"
	ClipboardCopy      = "ClipboardCopy"
)
