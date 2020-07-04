package sys

import (
	"context"
	"fmt"
	"mysql-monitor/dao"
	"mysql-monitor/global"
	"mysql-monitor/util"
	"strconv"
	"strings"
	"sync"
	"time"
)

/**
 * describe: 系统监控层,对操作系统如下指标进行监控
 */
type ItemName string

type Referce struct {
	Cycle     int     //默认单位秒
	Threshold float64 //默认单位秒
}
type SystemMonitor struct {
	// 上下文
	context context.Context

	// 读写锁
	rwLock sync.RWMutex

	// 存放监控项采集周期以及报警参考值的map
	referMap map[ItemName]Referce

	// 收集:子goroutine中的错误信息
	errChan chan ItemName
}

// 简单工厂模式
func GenerateSingletonSystemMonitor() *SystemMonitor {
	// 读取DB，加载默认的采集周期和报警阈值到内存中
	m := make(map[ItemName]Referce, 128)
	m[global.CPUITEM] = Referce{
		10,
		0.95,
	}
	return &SystemMonitor{
		context:  context.Background(),
		referMap: m,
		errChan:  make(chan ItemName, 128),
	}
}

/**
 * 监控CPU的利用率
 */
func (s *SystemMonitor) SysCPUUsageRate() {
	// todo 测试，子goroutinepanic后，父任务可以收到通知

	for {
		// 获取采集周期和采集时间
		referce := s.referMap[global.CPUITEM]
		ticker := time.NewTicker(time.Second * time.Duration(referce.Cycle))
		// 定时采集
		select {
		case <-ticker.C:
			currentTime := time.Now()
			// 09:44:36 up 198 days, 21:31,  2 users,  load average: 0.00, 0.01, 0.05
			// 获取系统在1分钟、5分钟、15分钟内的负载值
			var loadShell = "uptime"
			loadAvg, status, err := util.SyncExecShell(loadShell)
			if status == 127 {
				fmt.Printf("Fail to exec shell:[%v] err:[%v]", loadShell, err.Error())
			}

			// todo 假数据
			loadAvg = "09:44:36 up 198 days, 21:31,  2 users,  load average: 0.04, 0.01, 0.05"

			// 获取当前采集的时间点: 09:44:36
			time := util.SubStringWithStartEnd(loadAvg, 0, 7)
			// 系统运行时常
			sysRunTime := util.SubStringBettweenSub1Sub2(loadAvg, "up", ",")
			// 当前在线人数:   2 users,  load average: 0.04, 0.01, 0.05"
			tempStr := util.SubFirstString(util.SubFirstString(loadAvg, ","), ",")
			index := strings.Index(tempStr, ",")
			users := util.SubStringWithStartEnd(tempStr, 0, index-1)
			// 获取负载信息
			subString := util.SubLastString(loadAvg, ": ") //获取:0.00, 0.01, 0.05
			arr := strings.Split(subString, ", ")          //截取:[0.00, 0.01, 0.05] 转换成float类型
			loadNum := make([]float64, 3)
			for i, v := range arr {
				float, err := strconv.ParseFloat(v, 64)
				if err != nil {
					fmt.Printf("Fail to parse %v to flaot , err:[%v]", v, err.Error())
					return
				}
				loadNum[i] = float
			}
			// 获取系统的CPU核心数
			var cpuNumShell = "cat /proc/cpuinfo | grep process | wc -l"
			cpuNum, status, err := util.SyncExecShell(cpuNumShell)
			// 假数据
			cpuNum = "2"
			if status == 127 {
				fmt.Printf("Fail to exec shell:[%v] err:[%v]", cpuNumShell, err.Error())
			}
			num, err := strconv.Atoi(cpuNum)
			if err != nil {
				fmt.Printf("Fail to cpuNum:[%v]  to int err:[%v]", cpuNum, err)
				return
			}
			// 计算平均负载情况：
			var avgLoad = 0.0
			for _, v := range loadNum {
				avgLoad += v
			}
			// 大于0.6 == 繁忙
			avgLoad = avgLoad / 10
			// 落库
			cpuInfo := dao.NewCpuInfo(currentTime, time, global.CPUITEM, users, loadNum[0], loadNum[1], loadNum[2], sysRunTime, num, avgLoad)
			id, err := cpuInfo.InsertOneCord()
			if err != nil || id == 0 {
				fmt.Printf("Fail to insert cpuInfo err:[%v]", err.Error())
				// todo 考虑下，当子goroutine执行出错时，下一步的动作是什么？
			}
			// 如果平均负载大于等于报警项，落库,计数+1
			if avgLoad >= referce.Threshold {
				monitor := dao.NewMonitor(global.CPUITEM)
				monitor.SaveOrUpdate()
			}
		}
	}
}

/**
 * 监控系统IO的利用率
 */
func (s *SystemMonitor) SysIOUsageRate() {

}

/**
 * 将采集到的信息落盘
 */
func (s *SystemMonitor) PersistenceItem() {

}
