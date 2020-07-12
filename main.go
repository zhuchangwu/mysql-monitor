package main

import (
	"mysql-monitor/common"// step2: 初始化log
	"mysql-monitor/global"// step1: 初始化DB、热更新配置需要的chan
	"mysql-monitor/sys"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

import  _ "mysql-monitor/global"

import  _ "mysql-monitor/common"

// 考虑重定向标准输出到日志文件。(因为panic级别的错误会在控制台打印)
func main() {

	runtime.GOMAXPROCS(1)

	//todo 解析配置文件，将配置文件中的配置解析全局的 配置结构体中

	// todo precheck
	// checkMonitorEnv 在启动的过程中，查询检查一下诸如cpu这种监控项是否在monitor表中已存在，如果不存在的话，启动报错。

	// 将Sys相关的各个监控项的初始采集周期和阈值加载进map

	// 创建SysMnoitor
	sysMonitor:= sys.GenerateSingletonSystemMonitor()
	sysMonitor.StartSysMonitor()

	// 启动Grpc-Server
	waitGroup:=new(sync.WaitGroup)
	waitGroup.Add(1)
	server:= global.NewRpcServer()
	go global.StartRpcServer(server,waitGroup)

	// 启动过程中出现异常后优雅的退出
	sc := make(chan os.Signal)
	signal.Notify(sc,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	for {
		time.Sleep(1 * time.Second)
		switch sig := <-sc; sig {
		case syscall.SIGINT, syscall.SIGTERM:
			common.Warn("[MysqlMonitor] closed signal:%v", sig)
			os.Exit(0)
		default:
			continue
		}
	}

}

