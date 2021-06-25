package inspector

type BlockActionType int

const (
	BlockActionNone BlockActionType = iota
	BlockActionNext
	BlockActionPrev
	BlockActionJump
	BlockActionQuit
)

type BlockAction struct {
	Action BlockActionType
	Args   []string
}
