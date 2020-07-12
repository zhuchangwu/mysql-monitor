package sys

import (
	"context"
	"fmt"
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
 *  目前有报警策略的监控项有：
 *  1、磁盘使用率：当挂载在/下的文件系统的磁盘使用率超过阙值时，报警。
 *  2、内存使用率：当内存剩余不到总内存的 x 阙值时，报警。
 *  3、cpu：当cpu的load-avg大于指定的阙值时，报警。
 */

// todo 针对每一个监控项写一些介绍
/**
 * IO监控，请确保安装了: yum -y install dstat
 *
 * 遇到如下的报错后: vim /usr/bin/dstat  将python修改成python2
 * [root@139 ~]# dstat -r
 *   File "/usr/bin/dstat", line 120
 *   except getopt.error, exc:
 */

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
	referMap map[ItemName]*Referce

	// 收集:子goroutine中的错误信息
	errChan chan *ChildGoroutineErrInfo
}

// 简单工厂模式
func GenerateSingletonSystemMonitor() *SystemMonitor {
	// todo 模拟从数据库中将初始化的信息读取出来
	m := LoadSysMonitorItemCycleAndThresholdFromDB()
	return &SystemMonitor{
		context:  context.Background(),
		referMap: m,
		errChan:  make(chan *ChildGoroutineErrInfo, 256),
	}
}

// 从数据库中加载初始性的信息
func LoadSysMonitorItemCycleAndThresholdFromDB() map[ItemName]*Referce {
	// 1、读取DB，加载默认的采集周期和报警阈值到内存中
	m := make(map[ItemName]*Referce, 24)

	// todo 这些数据从mysql-monitor表中读取加载
	m[global.SYS_ITEM_CPU] = &Referce{
		10,
		0.00,
	}

	// 2、内存报警阈值，当free小于 total*Threhold时触发报警
	m[global.SYS_ITEM_MEMORY] = &Referce{
		10,
		0.1,
	}

	// 3、存储的IO使用率
	// 内存报警阈值，磁盘已使用的空间大于80%时触发报警
	m[global.SYS_ITEM_STORE] = &Referce{
		10,
		0.2,
	}

	// 4、磁盘随机IO次数
	m[global.SYS_ITEM_DISKRANDOMIO] = &Referce{
		10,
		0,
	}

	// 5、流经网卡的流量
	m[global.SYS_ITEM_NETWORKCARDIO] = &Referce{
		10,
		0,
	}

	// 6、磁盘使用情况监控
	m[global.SYS_ITEM_DISKUSAGERATE] = &Referce{
		2,
		0.8,
	}

	// 7、CPU使用率采集时间
	m[global.SYS_ITEM_CPUUSAGERATE] = &Referce{
		2,
		0.9,
	}

	// 8、系统上的Task情况
	m[global.SYS_ITEM_TASKS] = &Referce{
		2,
		0.0,
	}
	return m
}

// 启动
func (s *SystemMonitor) StartSysMonitor() {
	go s.SysTasks()
	go s.SysDiskUsageRate()
	go s.SysNetworkCardIORate()
	go s.SysDiskRandomIORate()
	go s.SysStoreUsageRate()
	go s.SysMemoryUsageRate()
	go s.SysLoadAvgUsageRate()
	go s.SysCPUUsageRate()
	go s.handlePanicAndAlarm()
	go s.LoadNewestDataFromHotChan()
}

/**
 * df -h
 * 磁盘使用情况：总大小、已使用、未使用
 * [root@139 ~]# df -h
 * Filesystem      Size  Used Avail Use% Mounted on
 * devtmpfs        1.9G     0  1.9G   0% /dev
 * tmpfs           1.9G  144K  1.9G   1% /dev/shm
 * tmpfs           1.9G  185M  1.7G  10% /run
 * tmpfs           1.9G     0  1.9G   0% /sys/fs/cgroup
 * /dev/vda1        40G   36G  1.6G  96% /
 * tmpfs           379M     0  379M   0% /run/user/0
 * overlay          40G   36G  1.6G  96% /var/lib/docker/overlay2/f3b7a056768d1a5a87844f0aad0ccd6ee0c178a486480a56fe632f2eb642f291/merged
 *
 * 报警：当挂载在/下的文件系统的磁盘使用率超过阙值时，报警
 */
func (s *SystemMonitor) SysDiskUsageRate() {
	// 当前goroutine panic后，父任务可以收到通知
	defer s.handleException(global.SYS_ITEM_DISKUSAGERATE, global.PANIC)
	for {
		// 获取采集周期和采集时间
		referce := s.referMap[global.SYS_ITEM_DISKUSAGERATE]
		ticker := time.NewTicker(time.Second * time.Duration(referce.Cycle))
		common.Info("SysDiskUsageRateMonitor cycle:[%v] s", referce.Cycle)
		// 定时采集
		select {
		case <-ticker.C:
			// 取三次，求平均值
			// -net/total-
			// recv  send
			//   0     0
			//  66B    0
			//   0     0
			var loadShell = "df -h"
			memory, status, err := util.SyncExecShell(loadShell)
			if status == 127 { // todo -1，表示命令找不到
				common.Error("Fail to exec shell:[%v] err:[%v]", loadShell, err.Error()) // todo 这种地方应该退出，然后报警
			}
			// todo 假数据
			memory = "Filesystem      Size  Used Avail Use% Mounted on\ndevtmpfs        1.9G     0  1.9G   0% /dev\ntmpfs           1.9G  144K  1.9G   1% /dev/shm\ntmpfs           1.9G  185M  1.7G  10% /run\ntmpfs           1.9G     0  1.9G   0% /sys/fs/cgroup\n/dev/vda1        40G   36G  1.6G  96% /\ntmpfs           379M     0  379M   0% /run/user/0\noverlay          40G   36G  1.6G  96% /var/lib/docker/overlay2/f3b7a056768d1a5a87844f0aad0ccd6ee0c178a486480a56fe632f2eb642f291/merged"
			// 当前时间
			currentTime := time.Now()
			// 获取当前采集的时间点: 09:44:36
			time := util.GetTimeString(currentTime)
			// 获取各个分区的磁盘使用情况

			diskInfos := make([]*dao.Disk, 0)
			split := strings.Split(memory, "\n")
			for i := 1; i < len(split); i++ {
				item := util.SpilitStringBySpace(split[i])
				info := dao.NewDiskInfo(currentTime, time, global.SYS_ITEM_DISKUSAGERATE, item[0], item[1], item[2], item[3], item[4], item[5])
				diskInfos = append(diskInfos, info)
			}

			// 批量更新本次的采集项
			for i := 0; i < len(diskInfos); i++ {
				diskInfo := diskInfos[i]
				qr := diskInfo.SaveOrUpdateDiskInfo()
				if qr.Err != nil {
					common.Warn("Fail to update diskInfo err:[%v]", qr.Err.Error())
					s.handleException(global.SYS_ITEM_DISKUSAGERATE, global.SYS_UPDATE_ITEM_DISKUSAGERATE_ERR)
				}
				if qr.EffectRow == 0 {
					common.Warn("Fail to update diskInfo EffectRow 0")
					s.handleException(global.SYS_ITEM_DISKUSAGERATE, global.SYS_UPDATE_ITEM_DISKUSAGERATE_ERR)
				} else {
					common.Info("Update diskInfo itemName:[%v] ", diskInfo.ItemName)
				}
				// 报警；当挂在在/下的磁盘使用率大于阙值时，报警
				usage, err := strconv.ParseFloat(strings.Split(diskInfo.Usage, "%")[0], 64)
				if err != nil {
					common.Warn("fotmat [%v]  to int", strings.Split(diskInfo.Usage, "%")[0])
					return
				}
				if diskInfo.MountedOn == "/" && usage > referce.Threshold*100 {

					common.Warn("Warning file_system:[%v] has been greater than referce.Threshold:[%v]", diskInfo.FileSystem, referce.Threshold)
					monitor := dao.NewMonitor(global.SYS_ITEM_DISKUSAGERATE)
					qr := monitor.SaveOrUpdateMonitorInfo()
					if qr.Err != nil {
						common.Warn("Fail to update monitor err:[%v]", qr.Err.Error())
						s.handleException(global.SYS_ITEM_DISKUSAGERATE, global.SYS_UPDATE_DISKUSAGEMONITORINFO_ERR)
					}
					if qr.EffectRow == 0 {
						common.Warn("Fail to update monitor EffectRow 0")
						s.handleException(global.SYS_ITEM_DISKUSAGERATE, global.SYS_UPDATE_DISKUSAGEMONITORINFO_ERR)
					} else {
						common.Info("Update to monitor successful itemName:[%v] ", global.SYS_ITEM_DISKUSAGERATE)
					}
				}
			}
		}
	}
}

/**
 * 流经网卡转发的流量
 */
func (s *SystemMonitor) SysNetworkCardIORate() {
	// 当前goroutine panic后，父任务可以收到通知
	defer s.handleException(global.SYS_ITEM_NETWORKCARDIO, global.PANIC)
	for {
		// 获取采集周期和采集时间
		referce := s.referMap[global.SYS_ITEM_NETWORKCARDIO]
		ticker := time.NewTicker(time.Second * time.Duration(referce.Cycle))
		common.Info("SysDiskRandomIOMonitor cycle:[%v] s", referce.Cycle)
		// 定时采集
		select {
		case <-ticker.C:
			// 取三次，求平均值
			// -net/total-
			// recv  send
			//   0     0
			//  66B    0
			//   0     0
			var loadShell = "dstat -n 1 0"
			memory, status, err := util.SyncExecShell(loadShell)
			if status == 127 { // todo -1，表示命令找不到
				common.Error("Fail to exec shell:[%v] err:[%v]", loadShell, err.Error()) // todo 这种地方应该退出，然后报警
			}
			// todo 假数据
			memory = "-net/total-\n recv  send\n   0     0"
			// 当前时间
			currentTime := time.Now()
			// 获取当前采集的时间点: 09:44:36
			time := util.GetTimeString(currentTime)
			// 获取储存的IO读写使用情况
			space := util.SpilitStringBySpace(strings.Split(memory, "\n")[2])
			readRate := space[0]
			writRate := space[1]
			// 落库
			randIOInfo := dao.NewIOInfo(currentTime, time, global.SYS_ITEM_NETWORKCARDIO, readRate, writRate)
			qr, err := randIOInfo.InsertOneCord()
			if err != nil {
				common.Error("Fail to insert network card io err:[%v]", err.Error())
				// 向父goroutine汇报
				s.handleException(global.SYS_ITEM_NETWORKCARDIO, global.SYS_INSERT_NETWORKCARDIORATE_ERR)
			}
			if qr.LastInsertId == 0 {
				common.Error("Fail to insert network card io  LastInsertId:[%v]", qr.LastInsertId)
				s.handleException(global.SYS_ITEM_NETWORKCARDIO, global.SYS_INSERT_NETWORKCARDIORATE_ERR)
			} else {
				common.Info("Insert to network card io successful id:[%v] ", qr.LastInsertId)
			}
		}
	}
}

/**
 * 磁盘的随机读写 次数
 */
func (s *SystemMonitor) SysDiskRandomIORate() {
	// 当前goroutine panic后，父任务可以收到通知
	defer s.handleException(global.SYS_ITEM_DISKRANDOMIO, global.PANIC)
	for {
		// 获取采集周期和采集时间
		referce := s.referMap[global.SYS_ITEM_DISKRANDOMIO]
		ticker := time.NewTicker(time.Second * time.Duration(referce.Cycle))
		common.Info("SysDiskRandomIOMonitor cycle:[%v] s", referce.Cycle)
		// 定时采集
		select {
		case <-ticker.C:
			// 取三次，求平均值
			// --io/total-
			// read  writ
			// 0.01  2.15
			// 0     0
			// 0     0
			var loadShell = "dstat -r 1 2"
			memory, status, err := util.SyncExecShell(loadShell)
			if status == 127 { // todo -1，表示命令找不到
				common.Error("Fail to exec shell:[%v] err:[%v]", loadShell, err.Error()) // todo 这种地方应该退出，然后报警
			}
			// todo 假数据
			memory = "--io/total-\n read  writ\n0.01  2.15\n   0     0\n   0     0"
			// 当前时间
			currentTime := time.Now()
			// 获取当前采集的时间点: 09:44:36
			time := util.GetTimeString(currentTime)
			// 获取储存的IO读写使用情况
			space1 := util.SpilitStringBySpace(strings.Split(memory, "\n")[2])
			space2 := util.SpilitStringBySpace(strings.Split(memory, "\n")[3])
			space3 := util.SpilitStringBySpace(strings.Split(memory, "\n")[4])
			f1, err := strconv.ParseFloat(space1[0], 64)
			f2, err := strconv.ParseFloat(space2[0], 64)
			f3, err := strconv.ParseFloat(space3[0], 64)
			f4, err := strconv.ParseFloat(space1[1], 64)
			f5, err := strconv.ParseFloat(space2[1], 64)
			f6, err := strconv.ParseFloat(space3[1], 64)
			readRate := (f1 + f2 + f3) / 3
			sprintf1 := fmt.Sprintf("%.2f", readRate)
			writRate := (f4 + f5 + f6) / 3
			sprintf2 := fmt.Sprintf("%.2f", writRate)
			// 落库,
			randIOInfo := dao.NewIOInfo(currentTime, time, global.SYS_ITEM_DISKRANDOMIO, sprintf1, sprintf2)
			qr, err := randIOInfo.InsertOneCord()
			if err != nil {
				common.Error("Fail to insert randomIOInfo err:[%v]", err.Error())
				// 向父goroutine汇报
				s.handleException(global.SYS_ITEM_DISKRANDOMIO, global.SYS_INSERT_DISKRANDOMIO_ERR)
			}
			if qr.LastInsertId == 0 {
				common.Error("Fail to insert randomIOInfo LastInsertId:[%v]", qr.LastInsertId)
				s.handleException(global.SYS_ITEM_DISKRANDOMIO, global.SYS_INSERT_DISKRANDOMIO_ERR)
			} else {
				common.Info("Insert to randomIOInfo successful id:[%v] ", qr.LastInsertId)
			}
		}
	}
}

/**
 * 磁盘的存储的IO吞咽量：每秒读、每秒写
 */
func (s *SystemMonitor) SysStoreUsageRate() {
	// 当前goroutine panic后，父任务可以收到通知
	defer s.handleException(global.SYS_ITEM_STORE, global.PANIC)
	for {
		// 获取采集周期和采集时间
		referce := s.referMap[global.SYS_ITEM_STORE]
		ticker := time.NewTicker(time.Second * time.Duration(referce.Cycle))
		common.Info("SysDiskMonitor cycle:[%v] s", referce.Cycle)
		// 定时采集
		select {
		case <-ticker.C:
			// -dsk/total-
			// read  writ
			// 770B   13k
			var loadShell = "dstat -d 1 0"
			memory, status, err := util.SyncExecShell(loadShell)
			if status == 127 { // todo -1，表示命令找不到
				common.Error("Fail to exec shell:[%v] err:[%v]", loadShell, err.Error())
			}
			// todo 假数据
			memory = "-dsk/total-\n read  writ\n 770B   13k"
			// 当前时间
			currentTime := time.Now()
			// 获取当前采集的时间点: 09:44:36
			time := util.GetTimeString(currentTime)
			// 获取储存的IO读写使用情况
			space := util.SpilitStringBySpace(strings.Split(memory, "\n")[2])
			readRate := space[0]
			writRate := space[1]
			// 落库
			storeIOInfo := dao.NewIOInfo(currentTime, time, global.SYS_ITEM_STORE, readRate, writRate)
			qr, err := storeIOInfo.InsertOneCord()
			if err != nil {
				common.Error("Fail to insert storeIOInfo err:[%v]", err.Error())
				// 向父goroutine汇报
				s.handleException(global.SYS_ITEM_STORE, global.SYS_INSERT_STOREINFO_ERR)
			}
			if qr.LastInsertId == 0 {
				common.Error("Fail to insert storeIOInfo LastInsertId:[%v]", qr.LastInsertId)
				s.handleException(global.SYS_ITEM_STORE, global.SYS_INSERT_STOREINFO_ERR)
			} else {
				common.Info("Insert to storeIOInfo successful id:[%v] ", qr.LastInsertId)
			}
		}
	}
}

/**
 * 内存使用率监控
 */
func (s *SystemMonitor) SysMemoryUsageRate() {
	// 当前goroutine panic后，父任务可以收到通知
	defer s.handleException(global.SYS_ITEM_MEMORY, global.PANIC)
	for {
		// 获取采集周期和采集时间
		referce := s.referMap[global.SYS_ITEM_MEMORY]
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
			memoryInfo := dao.NewMemory(currentTime, time, global.SYS_ITEM_MEMORY, total, used, free, buff)
			qr, err := memoryInfo.InsertOneCord()
			if err != nil {
				common.Error("Fail to insert memoryInfo err:[%v]", err.Error())
				// 向父goroutine汇报
				s.handleException(global.SYS_ITEM_MEMORY, global.SYS_INSERT_MEMORY_ERR)
			}
			if qr.LastInsertId == 0 {
				common.Error("Fail to insert memoryInfo LastInsertId:[%v]", qr.LastInsertId)
				s.handleException(global.SYS_ITEM_MEMORY, global.SYS_INSERT_MEMORY_ERR)
			} else {
				common.Info("Insert to memoryInfo successful id:[%v] ", qr.LastInsertId)
			}
			// 剩余可用内存小于总内存的%10，报警  referce.Threshold
			if float64(free) < referce.Threshold*float64(total) {
				common.Warn("Warning freeMemory:[%v] has been smaller than total * referce.Threshold:[%v]", free, referce.Threshold)
				monitor := dao.NewMonitor(global.SYS_ITEM_MEMORY)
				qr := monitor.SaveOrUpdateMonitorInfo()
				if qr.Err != nil {
					common.Warn("Fail to update monitor err:[%v]", qr.Err.Error())
					s.handleException(global.SYS_ITEM_MEMORY, global.SYS_UPDATE_MEMORYINFO_ERR)
				}
				if qr.EffectRow == 0 {
					common.Warn("Fail to update monitor EffectRow 0")
					s.handleException(global.SYS_ITEM_MEMORY, global.SYS_UPDATE_MEMORYINFO_ERR)
				} else {
					common.Info("Update to monitor successful itemName:[%v] ", global.SYS_ITEM_MEMORY)
				}
			}
		}
	}
}

/**
 * 监控系统负载
 */
func (s *SystemMonitor) SysLoadAvgUsageRate() {
	// 当前goroutine panic后，父任务可以收到通知
	defer s.handleException(global.SYS_ITEM_CPU, global.PANIC)
	for {
		// 获取采集周期和采集时间
		referce := s.referMap[global.SYS_ITEM_CPU]
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
			if status == 127 || status == -1 {
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
			if status == 127 || status == -1 {
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
			cpuInfo := dao.NewCpuLoadAvgInfo(currentTime, time, global.SYS_ITEM_CPU, users, loadNum[0], loadNum[1], loadNum[2], sysRunTime, num, avgLoad)
			qr, err := cpuInfo.InsertOneCord()
			if err != nil {
				common.Error("Fail to insert cpuInfo err:[%v]", err.Error())
				// 向父goroutine汇报
				s.handleException(global.SYS_ITEM_CPU, global.SYS_INSERT_CPUINFO_ERR)
			}
			if qr.LastInsertId == 0 {
				common.Error("Fail to insert cpuInfo LastInsertId:[%v]", qr.LastInsertId)
				s.handleException(global.SYS_ITEM_CPU, global.SYS_INSERT_CPUINFO_ERR)
			} else {
				common.Info("Insert to cpuInfo successful id:[%v] ", qr.LastInsertId)
			}

			// 如果平均负载大于等于报警项，落库,计数+1
			if avgLoad >= referce.Threshold {
				common.Warn("Warning avgLoad:[%v] has been greater than referce.Threshold:[%v]", avgLoad, referce.Threshold)
				monitor := dao.NewMonitor(global.SYS_ITEM_CPU)
				qr := monitor.SaveOrUpdateMonitorInfo()
				if qr.Err != nil {
					common.Warn("Fail to update monitor err:[%v]", qr.Err.Error())
					s.handleException(global.SYS_ITEM_CPU, global.SYS_UPDATE_CPUMONITORINFO_ERR)
				}
				if qr.EffectRow == 0 {
					common.Warn("Fail to update monitor EffectRow 0")
					s.handleException(global.SYS_ITEM_CPU, global.SYS_UPDATE_CPUMONITORINFO_ERR)
				} else {
					common.Info("Update to monitor successful itemName:[%v] ", global.SYS_ITEM_CPU)
				}
			}
		}
	}
}

/**
 *  CPU调度运行队列长度
 *  本监控指标相关优秀的博客：https://www.cnblogs.com/makelu/p/11169270.html
 */
func (s *SystemMonitor) SysTasks() {
	// 当前goroutine panic后，父任务可以收到通知
	defer s.handleException(global.SYS_ITEM_TASKS, global.PANIC)
	for {
		// 获取采集周期和采集时间
		referce := s.referMap[global.SYS_ITEM_TASKS]
		ticker := time.NewTicker(time.Second * time.Duration(referce.Cycle))
		common.Info("SysTasksMonitor cycle:[%v] s", referce.Cycle)
		// 定时采集
		select {
		case <-ticker.C:
			// var loadShell = "top -n 1 | grep Tasks"
			// memory, status, err := util.SyncExecShell(loadShell)
			// if status == 127 { // todo -1，表示命令找不到
			// 	common.Error("Fail to exec shell:[%v] err:[%v]", loadShell, err.Error()) // todo 这种地方应该退出，然后报警
			// }
			// todo 假数据
			memory := "Tasks:  92 total,   1 running,  90 sleeping,   1 stopped,   0 zombie"
			// 当前时间
			currentTime := time.Now()
			// 获取当前采集的时间点: 09:44:36
			time := util.GetTimeString(currentTime)
			// 解析task总数，正在运行的进程...
			item := util.SpilitStringBySpace(memory)
			total, err := strconv.Atoi(item[1])
			running, err := strconv.Atoi(item[3])
			sleeping, err := strconv.Atoi(item[5])
			stoped, err := strconv.Atoi(item[7])
			zombie, err := strconv.Atoi(item[10])
			// 获取CPU的占用情况
			info := dao.NewTasksInfo(currentTime, time, global.SYS_ITEM_TASKS, total, running, sleeping, stoped, zombie)
			qr, err := info.InsertOneCord()
			// 批量更新本次的采集项
			if err != nil {
				common.Error("Fail to insert tasks err:[%v]", err.Error())
				// 向父goroutine汇报
				s.handleException(global.SYS_ITEM_TASKS, global.SYS_INSERT_TASKS_ERR)
			}
			if qr.LastInsertId == 0 {
				common.Error("Fail to insert tasks LastInsertId:[%v]", qr.LastInsertId)
				s.handleException(global.SYS_ITEM_TASKS, global.SYS_INSERT_TASKS_ERR)
			} else {
				common.Info("Insert to tasks successful id:[%v] ", qr.LastInsertId)
			}
		}
	}
}

/**
 *  监控：用户空间、内核空间、用户空间内改变过的优先级的进程占用CPU的百分比，以及空闲空间占用CPU的百分比
 *  [root@139 ~]# mpstat -P ALL
 *  Linux 3.10.0-1062.4.1.el7.x86_64 (139.9.92.235) 	07/08/2020 	_x86_64_	(2 CPU)
 * 					    用户空间 	用户空间内改变过优先级的进程    内核空间												   空闲空间
 *  12:48:03 PM  CPU    %usr    	%nice   					 %sys 	 %iowait    %irq   %soft  %steal  %guest  %gnice   %idle
 *  12:48:03 PM  all    0.21    	 0.00   					 0.07 	    0.10    0.00    0.00    0.00    0.00    0.00   99.61
 *  12:48:03 PM    0    0.22    	 0.00   					 0.07 	    0.07    0.00    0.00    0.00    0.00    0.00   99.64
 *  12:48:03 PM    1    0.21    	 0.00   					 0.08 	    0.13    0.00    0.00    0.00    0.00    0.00   99.58
 */
func (s *SystemMonitor) SysCPUUsageRate() {
	// 当前goroutine panic后，父任务可以收到通知
	defer s.handleException(global.SYS_ITEM_CPUUSAGERATE, global.PANIC)
	for {
		// 获取采集周期和采集时间
		referce := s.referMap[global.SYS_ITEM_CPUUSAGERATE]
		ticker := time.NewTicker(time.Second * time.Duration(referce.Cycle))
		common.Info("SysCPUUsageRateMonitor cycle:[%v] s", referce.Cycle)
		// 定时采集
		select {
		case <-ticker.C:
			var loadShell = "mpstat -P ALL"
			memory, status, err := util.SyncExecShell(loadShell)
			if status == 127 { // todo -1，表示命令找不到
				common.Error("Fail to exec shell:[%v] err:[%v]", loadShell, err.Error()) // todo 这种地方应该退出，然后报警
			}
			// todo 假数据
			memory = "Linux 3.10.0-1062.4.1.el7.x86_64 (139.9.92.235) \t07/08/2020 \t_x86_64_\t(2 CPU)\n\n09:18:41 PM  CPU    %usr   %nice    %sys %iowait    %irq   %soft  %steal  %guest  %gnice   %idle\n09:18:41 PM  all    0.21    0.00    0.07    0.10    0.00    0.00    0.00    0.00    0.00   99.61\n09:18:41 PM    0    0.22    0.00    0.07    0.07    0.00    0.00    0.00    0.00    0.00   99.64\n09:18:41 PM    1    0.21    0.00    0.08    0.13    0.00    0.00    0.00    0.00    0.00   99.58"
			// 当前时间
			currentTime := time.Now()
			// 获取CPU的占用情况
			cpuUsageInfos := make([]*dao.CpuUsageRateInfo, 0)
			split := strings.Split(memory, "\n")
			for i := 4; i < len(split); i++ {
				item := util.SpilitStringBySpace(split[i])
				info := dao.NewCpuUsageRateInfo(currentTime, item[0], global.SYS_ITEM_CPUUSAGERATE, item[3], item[4], item[5], item[12], item[2])
				cpuUsageInfos = append(cpuUsageInfos, info)
			}
			// 批量更新本次的采集项
			for i := 0; i < len(cpuUsageInfos); i++ {
				cpuUsageInfo := cpuUsageInfos[i]
				qr := cpuUsageInfo.SaveOrUpdateCpuUsageInfo()
				if qr.Err != nil {
					common.Warn("Fail to update cpu usage rate err:[%v]", qr.Err.Error())
					s.handleException(global.SYS_ITEM_CPUUSAGERATE, global.SYS_UPDATE_CPUUSAGEINFO_ERR)
				}
				if qr.EffectRow == 0 {
					common.Warn("Fail to update cpu usage rate  EffectRow 0")
					s.handleException(global.SYS_ITEM_CPUUSAGERATE, global.SYS_UPDATE_CPUUSAGEINFO_ERR)
				} else {
					common.Info("Update cpu usage rate  itemName:[%v] ", global.SYS_ITEM_CPUUSAGERATE)
				}
			}
		}
	}
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
 * 将报警消息写入chan
 */
func (s *SystemMonitor) handlePanicAndAlarm() {
	for {
		select {
		case panicInfo := <-s.errChan:
			switch panicInfo.ErrType {
			case global.PANIC:
				common.Warn("recieve child goroutine panic msg：panicInfo:[%v]", panicInfo)
				common.Warn("restart goroutine")

			case global.SYS_INSERT_CPUINFO_ERR:
				// todo

			case global.SYS_UPDATE_CPUMONITORINFO_ERR:
				// todo

			}
		default:
			common.Info("nothing todo will sleep 2 second")
			time.Sleep(time.Second * 2)
		}
	}
}

// 监控热加载chan
func (s *SystemMonitor) LoadNewestDataFromHotChan() {
	for {
		select {
		case newestData := <-global.SysHotLoadChan:
			cycle, _ := strconv.Atoi(newestData.Cycle)
			threshold, _ := strconv.ParseFloat(newestData.Threshold, 64)
			referce := &Referce{
				Cycle:     cycle,
				Threshold: threshold,
			}
			s.rwLock.Lock()
			// 更新采集周期、采集阙值
			s.referMap[ItemName(newestData.ItemName)] = referce
			s.rwLock.Unlock()
		default:
			common.Info("nothing in  Sys Hot Chan will sleep 5 seconds")
			time.Sleep(time.Second * 5)
		}
	}
}
