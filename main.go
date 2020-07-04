package main

import "runtime"
import _"mysql-monitor/global"

func main() {
	runtime.GOMAXPROCS(1)
	// 启动gprc服务器，通过	wg := new(sync.WaitGroup) 控制启动顺序

	// 启动工作启动工作写成
}
