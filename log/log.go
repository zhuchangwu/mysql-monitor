package log

import (
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Formatter = new(logrus.TextFormatter)                      //default
	log.Formatter.(*logrus.TextFormatter).DisableColors = true     // remove colors
	log.Formatter.(*logrus.TextFormatter).DisableTimestamp = false // remove timestamp from test output
	log.SetLevel(logrus.DebugLevel)
	file, err := os.OpenFile("mysql-monitor.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
}

func Debug(fmtStr string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		file:=parseFilePath(file)
		pre := "File:[" + file + "] line:[" + strconv.Itoa(line) + "] "
		log.Debugf(pre+fmtStr, args...)
		return
	}
	log.Debugf(fmtStr, args...)
}

func Info(fmtStr string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		file:=parseFilePath(file)
		pre := "File:[" + file + "] line:[" + strconv.Itoa(line) + "] "
		log.Infof(pre+fmtStr, args...)
		return
	}
	log.Infof(fmtStr, args...)
}

func Error(fmtStr string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		file:=parseFilePath(file)
		pre := "File:[" + file + "] line:[" + strconv.Itoa(line) + "] "
		log.Errorf(pre+fmtStr, args...)
		return
	}
	log.Errorf(fmtStr, args...)
}

func parseFilePath(pathStr string)string{
	msg := "File:[/Users/dxm/go/src/mysql-monitor/log/log_test.go]"
	index := strings.Index(msg, "mysql-monitor")
	index2 := strings.Index(msg, "]")
	return msg[index+13:index2]
}