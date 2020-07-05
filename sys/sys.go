package sys

import (
	"context"
	"mysql-monitor/common"
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

type ChildGoroutineErrInfo struct {
	ItemName  string // 采集项的名称
	ErrType   string // 反馈的错误信息类型
	ErrorInfo string // 具体的错误信息
}

type SystemMonitor struct {
	// 上下文
	context context.Context

	// 读写锁
	rwLock sync.RWMutex

	// 存放监控项采集周期以及报警参考值的map
	referMap map[ItemName]Referce

	// 收集:子goroutine中的错误信息
	errChan chan *ChildGoroutineErrInfo
}

// 简单工厂模式
func GenerateSingletonSystemMonitor() *SystemMonitor {
	// 读取DB，加载默认的采集周期和报警阈值到内存中
	m := make(map[ItemName]Referce, 24)
	m[global.ITEM_CPUITEM] = Referce{
		10,
		0.00,
	}
	// 内存报警阈值，当free小于 total*Threhold时触发报警
	m[global.ITEM_MEMORTY] = Referce{
		10,
		0.1,
	}
	return &SystemMonitor{
		context:  context.Background(),
		referMap: m,
		errChan:  make(chan *ChildGoroutineErrInfo, 256),
	}
}

/**
 * 内存使用率
 */
func (s *SystemMonitor) SysMemoryUsageRate() {
	// 当前goroutine panic后，父任务可以收到通知
	defer s.handleException(global.ITEM_MEMORTY, global.PANIC)
	for {
		// 获取采集周期和采集时间
		referce := s.referMap[global.ITEM_MEMORTY]
		ticker := time.NewTicker(time.Second * time.Duration(referce.Cycle))
		common.Info("SysMemoryMonitor cycle:[%v] s", referce.Cycle)
		// 定时采集
		select {
		case <-ticker.C:
			// [root@139 ~]# free -m
			//				 total(总共)  used（已使用  free（空闲） shared    buff/cache（OS缓存）   available
			// Mem:           3788        1091         161         184        2535                 2270
			// Swap:             0           0           0
			//
			// [root@139 ~]# free -m | grep Mem
			//  Mem:           3788        1092         161         184        2535        2269
			var loadShell = "free -m | grep Mem"
			memory, status, err := util.SyncExecShell(loadShell)
			if status == 127 { // todo -1，表示命令找不到
				common.Error("Fail to exec shell:[%v] err:[%v]", loadShell, err.Error())
			}
			// todo 假数据
			memory = "Mem:           3788        1092         161         184        2535        2269"
			// 当前时间
			currentTime := time.Now()
			// 获取当前采集的时间点: 09:44:36
			time := util.GetTimeString(currentTime)
			// 获取总内存、已使用内存、空闲内存、OS缓存
			memorys := util.SpilitStringBySpace(memory)
			total, _ := strconv.Atoi(memorys[1])
			used, _ := strconv.Atoi(memorys[2])
			free, _ := strconv.Atoi(memorys[3])
			buff, _ := strconv.Atoi(memorys[5])
			// 落库
			memoryInfo := dao.NewMemory(currentTime, time, global.ITEM_MEMORTY, total, used, free, buff)
			qr, err := memoryInfo.InsertOneCord()
			if err != nil {
				common.Error("Fail to insert memoryInfo err:[%v]", err.Error())
				// 向父goroutine汇报
				s.handleException(global.ITEM_MEMORTY, global.INSERT_CPUINFO_ERR)
			}
			if qr.LastInsertId == 0 {
				common.Error("Fail to insert memoryInfo LastInsertId:[%v]", qr.LastInsertId)
				s.handleException(global.ITEM_MEMORTY, global.INSERT_CPUINFO_ERR)
			} else {
				common.Info("Insert to memoryInfo successful id:[%v] ", qr.LastInsertId)
			}
			// 剩余可用内存小于总内存的%10，报警  referce.Threshold
			if float64(free) < referce.Threshold*float64(total) {
				common.Warn("Warning freeMemory:[%v] has been smaller than total * referce.Threshold:[%v]", free, referce.Threshold)
				monitor := dao.NewMonitor(global.ITEM_MEMORTY)
				qr := monitor.SaveOrUpdateMonitorInfo()
				if qr.Err != nil {
					common.Warn("Fail to update monitor err:[%v]", qr.Err.Error())
					s.handleException(global.ITEM_MEMORTY, global.UPDATE_CPUMONITORINFO_ERR)
				}
				if qr.EffectRow == 0 {
					common.Warn("Fail to update monitor EffectRow 0")
					s.handleException(global.ITEM_MEMORTY, global.UPDATE_CPUMONITORINFO_ERR)
				} else {
					common.Info("Update to monitor successful itemName:[%v] ", global.ITEM_MEMORTY)
				}
			}
		}
	}
}

// todo
// 3.2 us : 用户空间占用cpu百分比3.2
// 0.0 sy : 内核空间占用cpu的百分百
// 0.0 ni : 用户空间内改变过优先级的进程占用cpu的百分比
// 96.8 id: 空闲的cpu百分比
// 0.0 wa: 等待输入输出的进程占用cpu 的百分比
// 0.0 hi: 硬件cpu占用百分比
// 0.0 si: 软中断占用cpu百分比
// 0.0 st: 虚拟机占有cpu百分比
/**
 * 监控系统负载
 */
func (s *SystemMonitor) SysLoadAvgUsageRate() {
	// 当前goroutine panic后，父任务可以收到通知
	defer s.handleException(global.ITEM_CPUITEM, global.PANIC)
	for {
		// 获取采集周期和采集时间
		referce := s.referMap[global.ITEM_CPUITEM]
		ticker := time.NewTicker(time.Second * time.Duration(referce.Cycle))
		common.Info("SysLoadAvgUsageRateMonitor cycle:[%v]", referce.Cycle)
		// 定时采集
		select {
		case <-ticker.C:
			currentTime := time.Now()
			// 09:44:36 up 198 days, 21:31,  2 users,  load average: 0.00, 0.01, 0.05
			// 获取系统在1分钟、5分钟、15分钟内的负载值
			var loadShell = "uptime"
			loadAvg, status, err := util.SyncExecShell(loadShell)
			if status == 127 || status == -1{
				common.Error("Fail to exec shell:[%v] err:[%v]", loadShell, err.Error())
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
					common.Error("Fail to parse %v to flaot , err:[%v]", v, err.Error())
					return
				}
				loadNum[i] = float
			}
			// 获取系统的CPU核心数
			var cpuNumShell = "cat /proc/cpuinfo | grep process | wc -l"
			cpuNum, status, err := util.SyncExecShell(cpuNumShell)
			// 假数据
			cpuNum = "2"
			if status == 127 ||status == -1{
				common.Error("Fail to exec shell:[%v] err:[%v]", cpuNumShell, err.Error())
			}
			num, err := strconv.Atoi(cpuNum)
			if err != nil {
				common.Error("Fail to cpuNum:[%v]  to int err:[%v]", cpuNum, err)
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
			cpuInfo := dao.NewCpuInfo(currentTime, time, global.ITEM_CPUITEM, users, loadNum[0], loadNum[1], loadNum[2], sysRunTime, num, avgLoad)
			qr, err := cpuInfo.InsertOneCord()
			if err != nil {
				common.Error("Fail to insert cpuInfo err:[%v]", err.Error())
				// 向父goroutine汇报
				s.handleException(global.ITEM_CPUITEM, global.INSERT_CPUINFO_ERR)
			}
			if qr.LastInsertId == 0 {
				common.Error("Fail to insert cpuInfo LastInsertId:[%v]", qr.LastInsertId)
				s.handleException(global.ITEM_CPUITEM, global.INSERT_CPUINFO_ERR)
			} else {
				common.Info("Insert to cpuInfo successful id:[%v] ", qr.LastInsertId)
			}

			// 如果平均负载大于等于报警项，落库,计数+1
			if avgLoad >= referce.Threshold {
				common.Warn("Warning avgLoad:[%v] has been greater than referce.Threshold:[%v]", avgLoad, referce.Threshold)
				monitor := dao.NewMonitor(global.ITEM_CPUITEM)
				qr := monitor.SaveOrUpdateMonitorInfo()
				if qr.Err != nil {
					common.Warn("Fail to update monitor err:[%v]", qr.Err.Error())
					s.handleException(global.ITEM_CPUITEM, global.UPDATE_CPUMONITORINFO_ERR)
				}
				if qr.EffectRow == 0 {
					common.Warn("Fail to update monitor EffectRow 0")
					s.handleException(global.ITEM_CPUITEM, global.UPDATE_CPUMONITORINFO_ERR)
				} else {
					common.Info("Update to monitor successful itemName:[%v] ", global.ITEM_CPUITEM)
				}
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

/**
 *  统一panic处理，当负责采集信息当goroutinpanic后在此函数中重新启动向父goroutine中发送信号
 */
func (s *SystemMonitor) handleException(itemName string, errorType string) {
	if err := recover(); err != nil {
		common.Warn("gorountine panic will send msg to parent gorutine, msg:[%v]", err)
		s.errChan <- &ChildGoroutineErrInfo{
			ItemName:  itemName,
			ErrType:   errorType,
			ErrorInfo: err.(error).Error(),
		}
	}
}

/**
 * 将
 */
func (s *SystemMonitor) handlePanicAndAlarm() {
	for {
		select {
		case panicInfo := <-s.errChan:
			switch panicInfo.ErrType {
			case global.PANIC:
				common.Warn("recieve child goroutine panic msg：panicInfo:[%v]", panicInfo)
				common.Warn("restart goroutine")

			case global.INSERT_CPUINFO_ERR:
				// todo

			case global.UPDATE_CPUMONITORINFO_ERR:
				// todo

			}
		default:
			common.Info("nothing todo will sleep 2 second")
			time.Sleep(time.Second * 2)
		}
	}
}
