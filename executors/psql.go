package executor

type PSQLExecutor struct {
	BaseExecutor
}

var PSQL = &ContentType{"text", "psql", map[string]string{}}

func NewPSQLExecutor(content []byte, contentType *ContentType) (Executor, error) {
	pe := &PSQLExecutor{}
	pe.setDefaults()

	err := pe.processContentType(content, PSQL, contentType)
	if err != nil {
		return nil, err
	}

	if pe.command == "" {
		pe.command = "psql"
	}

	return pe, nil
}
