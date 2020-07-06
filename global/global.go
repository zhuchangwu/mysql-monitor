package global

import (
	"mysql-monitor/common"
	"mysql-monitor/connector"
)

// 常量：采集项名称
const (
	ITEM_CPUITEM       = "CPUITEM"            // CPU
	ITEM_MEMORTY       = "MEMORTY"            // 内存
	ITEM_STORE         = "STORE"              // 存储的IO
	ITEM_DISKRANDOMIO  = "DISKRANDOMIO"       // 磁盘随机io数
	ITEM_NETWORKCARDIO = "ITEM_NETWORKCARDIO" // 网卡的读写流量
)

// 常量：采集过程中 子goroutine向父goroutine汇报的错误类型
const (
	PANIC                     = "PANIC"                     // panic
	INSERT_CPUINFO_ERR        = "INSERT_CPUINFO_ERR"        // 保存采集到的cpu信息时报错失败
	INSERT_STOREINFO_ERR      = "INSERT_STOREINFO_ERR"      // 保存采集到的存储的读写IO时报错失败
	UPDATE_CPUMONITORINFO_ERR = "UPDATE_CPUMONITORINFO_ERR" // 更新CPU—Motior表时报错失败
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
