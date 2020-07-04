package sys

import "testing"

func TestMonitorCpu(t *testing.T){
	monitor := GenerateSingletonSystemMonitor()
	monitor.SysCPUUsageRate()
}