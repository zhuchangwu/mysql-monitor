package dao

import "C"
import (
	"context"
	"mysql-monitor/common"
	"mysql-monitor/connector"
	"mysql-monitor/global"
	"strconv"
	"time"
)

/**
 * CPU负载的信息
 */
type CpuLoadAvgInfo struct {
	ID            int
	CurDate       time.Time // 格式 2020-1-1 10:20:21
	CurTime       string    // 格式 10:20:21
	ItemName      string    // 监控项名称
	Users         string    // 当前在线人数
	OneMinute     float64   // 1分钟
	FiveMinute    float64   // 5分钟
	FifteenMinute float64   // 15分钟
	SysRuntime    string    // 系统运行时长
	CpuNum        int       // cpu个数
	LoadAvg       float64   // 当前平均负载
}



/**
 * 操作系统对进程总数、正在进行进程数、休眠的进程、僵尸进程
 */
func NewCpuLoadAvgInfo(date time.Time, time string, itemName string, users string, one, five, fifteen float64, sysRunTime string, cpuNum int, loadAvg float64) *CpuLoadAvgInfo {
	return &CpuLoadAvgInfo{
		CurDate:       date,
		CurTime:       time,
		ItemName:      itemName,
		Users:         users,
		OneMinute:     one,
		FiveMinute:    five,
		FifteenMinute: fifteen,
		SysRuntime:    sysRunTime,
		CpuNum:        cpuNum,
		LoadAvg:       loadAvg,
	}
}



/**
 * 插入一条数据
 */
func (c *CpuLoadAvgInfo) InsertOneCord() (qr *connector.QueryResults, err error) {
	connector := global.DB
	timeOut, _ := strconv.Atoi(connector.BaseInfo.ConnTimeOut)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(timeOut)*time.Second)
	sqlText := "insert into sys_cpu (cur_date, cur_time, item_name, users, one_minute, five_minute, fifteen_minute, sys_runtime, cpu_num, load_avg) values (?,?,?,?,?,?,?,?,?,?);"
	qr = connector.Exec(ctx, sqlText, c.CurDate, c.CurTime, c.ItemName, c.Users, c.OneMinute, c.FiveMinute, c.FifteenMinute, c.SysRuntime, c.CpuNum, c.LoadAvg)
	if nil != qr.Err {
		common.Error("Fail to inset cpuInfo ,sqlTest:[%v] err:[%v]", sqlText, err.Error())
		cancelFunc()
	}
	// 使用global的
	return qr, err
}

