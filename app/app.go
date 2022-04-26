package app

var (
	Name         = "pryrite"
	AppDirectory = ".pryrite"
	CommitHash   string
	Version      string
)

var atExitFuncs = []func(){}

func AtExit(fn func()) {
	atExitFuncs = append(atExitFuncs, fn)
}

func BeforeExit() {
	for _, fn := range atExitFuncs {
		fn()
	}
}
