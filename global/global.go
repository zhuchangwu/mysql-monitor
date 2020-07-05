package global

import (
	"mysql-monitor/common"
	"mysql-monitor/connector"
)

// 常量：采集项名称
const (
	CPUITEM = "CPUITEM"
)

// 常量：采集过程中 子goroutine向父goroutine汇报的错误类型
const (
	PANIC = "PANIC" // panic
	INSERT_CPUINFO_ERR = "INSERT_CPUINFO_ERR" // 保存采集到的cpu信息时报错失败
	UPDATE_CPUMONITORINFO_ERR = "UPDATE_CPUMONITORINFO_ERR"  // 更新CPU—Motior表时报错失败
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


