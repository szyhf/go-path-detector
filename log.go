package detector

import (
	"io"
	logpkg "log"
)

var log *logpkg.Logger = logpkg.New(new(nullLogger), "", 0)

func SetLogger(out io.Writer, prefix string, flag int) {
	log = logpkg.New(out, prefix, flag)
}

type nullLogger struct {
}

func (l *nullLogger) Write([]byte) (int, error) {
	return 0, nil
}
