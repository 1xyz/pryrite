package graph

type Kind string

const (
	Unknown       Kind = "Unknown"
	Command       Kind = "Command"
	Slurp         Kind = "Slurp"
	AsciiCast     Kind = "AsciiCast"
	TextSelect    Kind = "TextSelect"
	ClipboardCopy Kind = "ClipboardCopy"
	Text          Kind = "Text"
)
