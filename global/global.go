package global

import (
	"mysql-monitor/common"
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
	common.Info("Ready to init DB...")
	DB = &connector.Conenctor{}
	DB.Open()
	common.Info("DB Init successful")
}


