package app

var Name string
var AppDirectory string
var Version string
var CommitHash string

var atExitFuncs = []func(){}

func AtExit(fn func()) {
	atExitFuncs = append(atExitFuncs, fn)
}

func BeforeExit() {
	for _, fn := range atExitFuncs {
		fn()
	}
}
