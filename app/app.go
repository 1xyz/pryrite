package app

var Name string
var Version string
var BuildTime string
var CommitHash string
var UsageName string // Display name used in usage strings

var atExitFuncs = []func(){}

func AtExit(fn func()) {
	atExitFuncs = append(atExitFuncs, fn)
}

func BeforeExit() {
	for _, fn := range atExitFuncs {
		fn()
	}
}
