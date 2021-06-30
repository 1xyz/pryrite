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

func NewBlockAction(name string) *BlockAction {
	action := map[string]BlockActionType{
		"next": BlockActionNext,
		"prev": BlockActionPrev,
		"jump": BlockActionJump,
		"quit": BlockActionQuit,
	}
	if a, ok := action[name]; ok {
		return &BlockAction{Action: a}
	}
	return nil
}
