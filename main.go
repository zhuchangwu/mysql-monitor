package main

import "runtime"
// 初始化DB
import _ "mysql-monitor/global"
// 初始化log
import _ "mysql-monitor/common"

func main() {
	runtime.GOMAXPROCS(1)
	// 启动gprc服务器，通过	wg := new(sync.WaitGroup) 控制启动顺序

	// todo 将 monitor中各个监控项的监控周期、报警阈值加载进map中。
	// todo checkMonitorEnv 在启动的过程中，查询检查一下诸如cpu这种监控项是否在monitor表中已存在，如果不存在的话，启动报错。

	// 启动工作启动工作
}
