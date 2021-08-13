package executor

import "runtime"

type WinBashExecutor struct {
	BaseExecutor
}

func NewWinBashExecutor(content []byte, contentType *ContentType) (Executor, error) {
	if runtime.GOOS != "windows" {
		return nil, ErrUnsupportedContentType
	}

	we := &WinBashExecutor{}
	we.setDefaults()

	err := we.processContentType(content, Bash, contentType)
	if err != nil {
		return nil, err
	}

	we.name = "win-" + we.name
	if we.command == "" {
		we.setExecCommand("bash", nil)
	}

	return we, nil
}
