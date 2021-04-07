package events

type Kind string

const (
	Unknown       Kind = "Unknown"
	Console            = "Console"
	AsciiCast          = "AsciiCast"
	PageOpen           = "PageOpen"
	PageClose          = "PageClose"
	TextSelect         = "TextSelect"
	ClipboardCopy      = "ClipboardCopy"
)
