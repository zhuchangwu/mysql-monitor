package global

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"mysql-monitor/common"
	"mysql-monitor/connector"
	monitor "mysql-monitor/pb/monitor"
	task "mysql-monitor/pb/monitor/task"
	"net"
	"os"
	"sync"
	"time"
)

// 数据库连接
var DB *connector.Conenctor

// 常量：（操作系统系统相关）采集项名称
const (
	SYS_ITEM_CPU           = "cpu"             // CPU
	SYS_ITEM_MEMORY        = "memory"          // 内存
	SYS_ITEM_STORE         = "store"           // 存储的IO
	SYS_ITEM_DISKRANDOMIO  = "disk—random-io"  // 磁盘随机io数
	SYS_ITEM_NETWORKCARDIO = "network-card-io" // 网卡的读写流量
	SYS_ITEM_DISKUSAGERATE = "disk-usage-rate" // 磁盘的存储
	SYS_ITEM_CPUUSAGERATE  = "cpu-usage-rate"  // CPU的使用率
	SYS_ITEM_TASKS         = "tasks"           // 任务队列数
)

// 常量：（操作系统系统相关）采集过程中，子goroutine向父goroutine汇报的错误类型
const (
	PANIC                             = "panic"                           // panic
	SYS_INSERT_CPUINFO_ERR            = "insert-cpu-info-err"             // 保存采集到的cpu信息时报错
	SYS_INSERT_MEMORY_ERR             = "insert-memory-err"               // 保存采集到的内存使用情况信息时报错
	SYS_INSERT_TASKS_ERR              = "insert-tasks-err"                // 保存采集到的cpu信息时报错失败
	SYS_INSERT_STOREINFO_ERR          = "insert-store-info-err"           // 保存采集到的存储的读写IO时报错
	SYS_INSERT_NETWORKCARDIORATE_ERR  = "insert-network-card-io-rate-err" // 保存采集到的流经网卡的流量失败
	SYS_INSERT_DISKRANDOMIO_ERR       = "insert-disk-random-io-err"       // 保存采集到的磁盘随机读写失败
	SYS_UPDATE_CPUMONITORINFO_ERR     = "update-cpu-monitor-info_err"     // 更新CPU—Motior表时报错
	SYS_UPDATE_CPUUSAGEINFO_ERR       = "update-cpu-usage-info-err"       // 更新CPU使用率报错
	SYS_UPDATE_MEMORYINFO_ERR         = "update-memory-info-err"          // 更新内存相关监控时报错
	SYS_UPDATE_ITEM_DISKUSAGERATE_ERR = "update-item-disk-usage-rate-err" // 更新磁盘使用率时报错
)

// 收集:子goroutine中的错误信息
var SysHotLoadChan chan *task.SysMsg
var AppHotLoadChan chan *task.AppMsg
var MysqlHotLoadChan chan *task.MySQLMsg

func init() {
	// 初始化数据库连接
	common.Info("Ready to init DB...")
	DB = &connector.Conenctor{}
	DB.Open()
	common.Info("DB Init successful")

	// 初始化热加载用于 Sys、App、Mysql 能监控的配置chan
	SysHotLoadChan = make(chan *task.SysMsg, 256)
	AppHotLoadChan = make(chan *task.AppMsg, 256)
	MysqlHotLoadChan = make(chan *task.MySQLMsg, 256)
	common.Info("Sys、App、Mysql Hot Load chan make successfully")
}


// 构建Server
func NewRpcServer() (grpcServer *grpc.Server){
	var serverOptions []grpc.ServerOption
	// todo 这里的时间从配置文件中读取
	serverOptions = append(serverOptions, grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: 10 * time.Second,
		Time:              10 * time.Second,
		Timeout:           10 * time.Second,
	}))
	common.Info("Prepare to build Grpc-Server")
	// 构建server
	grpcServer = grpc.NewServer(serverOptions...)
	return grpcServer
}

// 启动Server
func StartRpcServer(grpcServer *grpc.Server,waitGroup *sync.WaitGroup) {
	// todo ip、port 从配置文件中读
	lis, err:= net.Listen("tcp", ":8082")
	if err != nil {
		common.Warn("Grpc-Server start fail error:[%v]", err)
		os.Exit(1)
		return
	}
	// 注册server
	task.RegisterFlowServiceServer(grpcServer, new(monitor.MySQLMonitorService))
	err = grpcServer.Serve(lis)
	if err != nil {
		common.Warn("Grpc-Server fail to exec grpcServer.Serve(lis) err:[%v]", err)
		return
	}
	waitGroup.Done()
	common.Info("Grpc Server successful")

}

