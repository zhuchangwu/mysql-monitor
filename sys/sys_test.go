package sys

import (
	"fmt"
	"mysql-monitor/global"
	_ "mysql-monitor/global"
	"testing"
	"time"
)

func TestMonitorCpu(t *testing.T){
	monitor := GenerateSingletonSystemMonitor()
	monitor.SysCPUUsageRate()
}


/**
 * 定时操作
 */
func TestTypeParse(t *testing.T) {
	monitor := GenerateSingletonSystemMonitor()
	referce := monitor.referMap[global.CPUITEM]
	ticker := time.NewTicker(time.Second * time.Duration(referce.Cycle))
	select {
	case <-ticker.C:
		fmt.Println("Hello")
	}

	fmt.Println("done")
}

/**
 *
 */
func TestTime(t *testing.T)  {
	now := time.Now()
	fmt.Println(now)
}