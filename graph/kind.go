package graph

type Kind string

const (
	Unknown       Kind = "Unknown"
	Command       Kind = "Command"
	AsciiCast     Kind = "AsciiCast"
	PageOpen      Kind = "PageOpen"
	PageClose     Kind = "PageClose"
	TextSelect    Kind = "TextSelect"
	ClipboardCopy Kind = "ClipboardCopy"
	Text          Kind = "Text"
)
