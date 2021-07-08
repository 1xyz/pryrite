package inspector

type BlockActionType int

const (
	BlockActionNone BlockActionType = iota
	BlockActionNext
	BlockActionPrev
	BlockActionJump
	BlockActionQuit
	BlockActionExecutionDone
)

type BlockAction struct {
	Action BlockActionType
	Args   []string
}

var (
	action = map[string]BlockActionType{
		"next": BlockActionNext,
		"prev": BlockActionPrev,
		"jump": BlockActionJump,
		"quit": BlockActionQuit,
	}
)

func NewBlockAction(name string) *BlockAction {
	if a, ok := action[name]; ok {
		return &BlockAction{Action: a}
	}
	return nil
}
