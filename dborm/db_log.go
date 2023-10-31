package dborm

import (
	"fmt"
	"io"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

type dbLog struct {
	writer io.Writer
}

func newDBLog(logBase string) *dbLog {
	if logBase == "" {
		logBase = "log"
	}
	filename := strings.Join([]string{logBase, "db.log"}, "/")

	writer := &lumberjack.Logger{
		Filename:   filename,
		LocalTime:  true,
		MaxSize:    10,   // the maximum size of each log file (in megabytes)
		MaxBackups: 5,    // the maximum number of old log files to retain
		MaxAge:     30,   // the maximum number of days to retain old log files
		Compress:   true, // compress old log files with gzip
	}
	return &dbLog{
		writer: writer,
	}
}

func (l *dbLog) Printf(format string, args ...interface{}) {
	if l == nil || l.writer == nil {
		return
	}

	_, _ = l.writer.Write([]byte(fmt.Sprintf(format, args...) + "\n"))
}
