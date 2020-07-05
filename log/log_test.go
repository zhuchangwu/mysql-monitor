package log

import (
	"fmt"
	_ "mysql-monitor/util"
	"strings"
	"testing"
)

func TestExample_basic(t *testing.T) {
	Debug("name:[%v] age:[%v]", "tom", 23)
	Info("name:[%v] age:[%v]", "tom", 23)
	Error("name:[%v] age:[%v]", "tom", 23)
}

func TestSpilitFilePath(t *testing.T) {
	msg := "File:[/Users/dxm/go/src/mysql-monitor/log/log_test.go]"
	index := strings.Index(msg, "mysql-monitor")
	index2 := strings.Index(msg, "]")
	msg = msg[index+13:index2]
	fmt.Println(msg)
}

