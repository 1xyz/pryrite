package graph

type Kind string

const (
	Unknown       Kind = "Unknown"
	Command       Kind = "Command"
	AsciiCast     Kind = "AsciiCast"
	TextSelect    Kind = "TextSelect"
	ClipboardCopy Kind = "ClipboardCopy"
	Text          Kind = "Text"
)
