package sys

import (
	"fmt"
	"mysql-monitor/global"
	_ "mysql-monitor/global"
	"strconv"
	"testing"
	"time"
)

func TestSysStoreUsageRate(t *testing.T)  {
	monitor := GenerateSingletonSystemMonitor()
	monitor.SysStoreUsageRate()
}

func TestSysDiskRandomIORate(t *testing.T)  {
	monitor := GenerateSingletonSystemMonitor()
	monitor.SysDiskRandomIORate()
}


func TestSysMemoryUsageRate(t *testing.T){
	monitor := GenerateSingletonSystemMonitor()
	monitor.SysMemoryUsageRate()
}


func TestMonitorChildGroutine(t *testing.T) {
	type Task struct {
		ItemName string
		FuncName string
		ErrorMsg string
	}
	tasks := make(chan *Task, 10)

	handlePanic := func() {
		if err := recover(); err != nil {
			fmt.Println("gorountine panic will send msg to parent gorutine")
			tasks <- &Task{
				ItemName: "cpuMonitor",
				FuncName: "cpuMonitorFunc",
				ErrorMsg: "error~~~",
			}
		}
	}

	go func() {
		defer handlePanic()
		// 模拟工作10秒后panic退出了
		for {
			fmt.Println("child goroutine start work～～～")
			time.Sleep(5 * time.Second)
			panic("wow! panic")
		}

		fmt.Println("child goroutine panic, send msg to parent")
		tasks <- &Task{
			ItemName: "cpuMonitor",
			FuncName: "cpuMonitorFunc",
		}
	}()

	go func() {
		for {
			select {
			case task := <-tasks:
				fmt.Println("有子goroutine退出了～")
				fmt.Printf("%#v", *task)
				fmt.Println("报警～")
			default:
				fmt.Println("nothingtodo～")
				time.Sleep(2 * time.Second)
			}

		}
	}()

	time.Sleep(60 * time.Second)
}

func TestMonitorCpu(t *testing.T) {
	monitor := GenerateSingletonSystemMonitor()
	monitor.SysLoadAvgUsageRate()
}

/**
 * 定时操作
 */
func TestTypeParse(t *testing.T) {
	monitor := GenerateSingletonSystemMonitor()
	referce := monitor.referMap[global.ITEM_CPUITEM]
	ticker := time.NewTicker(time.Second * time.Duration(referce.Cycle))
	select {
	case <-ticker.C:
		fmt.Println("Hello")
	}

	fmt.Println("done")
}

func TestTime(t *testing.T) {
	now := time.Now().Hour()
	M := time.Now().Minute()
	S := time.Now().Second()
	fmt.Println(strconv.Itoa(now) + ":" +strconv.Itoa(M) + ":"+strconv.Itoa(S))
}
