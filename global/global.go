package global

import (
	"fmt"
	"mysql-monitor/connector"
)

// 常量
const (
	CPUITEM = "CPUITEM"
)

// 数据库
var DB *connector.Conenctor

// 初始化数据库连接
func init() {
	fmt.Println("Ready to init DB...")
	DB = &connector.Conenctor{}
	DB.Open()
	fmt.Println("Ready to init DB Done")
}


