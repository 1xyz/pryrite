package run

type StatusLevel string

const (
	StatusInfo  StatusLevel = "Info"
	StatusError StatusLevel = "Error"
)

type Status struct {
	Level   StatusLevel
	Message string
}
