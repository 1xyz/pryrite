package executor

type MySQLExecutor struct {
	BaseExecutor
}

var MySQL = &ContentType{"text", "mysql", map[string]string{}}

func NewMySQLExecutor(content []byte, contentType *ContentType) (Executor, error) {

	me := &MySQLExecutor{}
	me.setDefaults()

	err := me.processContentType(content, MySQL, contentType)
	if err != nil {
		return nil, err
	}

	if me.command == "" {
		me.setExecCommand("mysql", nil)
	}

	return me, nil
}
