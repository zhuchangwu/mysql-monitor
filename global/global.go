package global

import (
	"mysql-monitor/common"
	"mysql-monitor/connector"
)

// 常量：采集项名称
const (
	ITEM_CPUITEM       = "cpu"             // CPU
	ITEM_MEMORY        = "memory"          // 内存
	ITEM_STORE         = "store"           // 存储的IO
	ITEM_DISKRANDOMIO  = "disk—random-io"  // 磁盘随机io数
	ITEM_NETWORKCARDIO = "network-card-io" // 网卡的读写流量
	ITEM_DISKUSAGERATE = "disk-usage-rate" // 磁盘的存储
	ITEM_CPUUSAGERATE  = "cpu-usage-rate"  // CPU的使用率
	ITEM_TASKS         = "tasks"           // 任务队列数
)

// 常量：采集过程中 子goroutine向父goroutine汇报的错误类型
const (
	PANIC                         = "panic"                         // panic
	INSERT_CPUINFO_ERR            = "insert-cpu-info-err"            // 保存采集到的cpu信息时报错
	INSERT_MEMORY_ERR             = "insert-memory-err"             // 保存采集到的内存使用情况信息时报错
	INSERT_TASKS_ERR              = "insert-tasks-err"              // 保存采集到的cpu信息时报错失败
	INSERT_STOREINFO_ERR          = "insert-store-info-err"          // 保存采集到的存储的读写IO时报错
	INSERT_NETWORKCARDIORATE_ERR  = "insert-network-card-io-rate-err"  // 保存采集到的流经网卡的流量失败
	INSERT_DISKRANDOMIO_ERR       = "insert-disk-random-io-err"       // 保存采集到的磁盘随机读写失败
	UPDATE_CPUMONITORINFO_ERR     = "update-cpu-monitor-info_err"     // 更新CPU—Motior表时报错
	UPDATE_CPUUSAGEINFO_ERR       = "update-cpu-usage-info-err"       // 更新CPU使用率报错
	UPDATE_MEMORYINFO_ERR        = "update-memory-info-err"        // 更新内存相关监控时报错
	UPDATE_ITEM_DISKUSAGERATE_ERR = "update-item-disk-usage-rate-err" // 更新磁盘使用率时报错
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
