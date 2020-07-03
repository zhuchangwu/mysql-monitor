package sys

import (
	"context"
	"sync"
)

/**
 * describe: 系统监控层,对操作系统如下指标进行监控
 */

type itemName string
type cycle int64

type SystemMonitor struct {
	// 上下文
	context context.Context

	// 读写锁
	rwLock sync.RWMutex

	// 存放监控项周期的map
	cycleMap map[itemName]cycle

	// 收集:子goroutine中的错误信息
	errChan chan itemName
}

// 简单工厂模式
func GenerateSingletonSystemMonitor() *SystemMonitor {
	return &SystemMonitor{
		context:  context.Background(),
		cycleMap: make(map[itemName]cycle, 128),
		errChan:  make(chan itemName, 128),
	}
}

/**
 * 监控CPU的利用率
 */
func (s *SystemMonitor) SysCPUUsageRate() {

}

/**
 * 监控系统IO的利用率
 */
func (s *SystemMonitor) SysIOUsageRate() {

}

/**
 * 将采集到的信息落盘
 */
func(s *SystemMonitor) PersistenceItem(){

}
