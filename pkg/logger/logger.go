package logger

import (
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Errorf(action string, err interface{}) {
	var filename, function string
	pc, file, _, ok := runtime.Caller(1)
	if ok {
		filename = filepath.Base(file)
		toSplit := runtime.FuncForPC(pc).Name()
		splited := strings.Split(toSplit, ".")
		function = splited[len(splited)-1]
	}
	log.Errorf("%s --> func %s --> action %v --> error: %v", filename, function, action, err)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}
