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
 * 用户空间、内核空间对CPU占用的百分比，用户空间内改变过的优先级的进程占用CPU的百分比
 * CPU空闲百分比
 *
 */
type CpuUsageRateInfo struct {
	ID       int
	CurDate  time.Time // 格式 2020-1-1 10:20:21
	CurTime  string    // 格式 10:20:21
	ItemName string    // 监控项名称
	Usr      string    // 用户空间
	Nice     string    // 用户空间内改变过优先级的进程
	Sys      string    // 内核空间
	Idle     string    // 空闲空间
	CpuNum   string    // CPU号
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
 * 各个空间对CPU占用情况
 */
func NewCpuUsageRateInfo(date time.Time, time string, itemName string, usr, nice, sys, idle, cpuNum string) *CpuUsageRateInfo {
	return &CpuUsageRateInfo{
		CurDate:  date,
		CurTime:  time,
		ItemName: itemName,
		Usr:      usr,    // 用户空间
		Nice:     nice,   // 用户空间内改变过优先级的进程
		Sys:      sys,    // 内核空间
		Idle:     idle,   // 空闲空间
		CpuNum:   cpuNum, // 空闲空间
	}
}

/**
 * 插入一条数据
 */
func (c *CpuLoadAvgInfo) InsertOneCord() (qr *connector.QueryResults, err error) {
	connector := global.DB
	timeOut, _ := strconv.Atoi(connector.BaseInfo.ConnTimeOut)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(timeOut)*time.Second)
	sqlText := "insert into cpu (cur_date, cur_time, item_name, users, one_minute, five_minute, fifteen_minute, sys_runtime, cpu_num, load_avg) values (?,?,?,?,?,?,?,?,?,?);"
	qr = connector.Exec(ctx, sqlText, c.CurDate, c.CurTime, c.ItemName, c.Users, c.OneMinute, c.FiveMinute, c.FifteenMinute, c.SysRuntime, c.CpuNum, c.LoadAvg)
	if nil != qr.Err {
		common.Error("Fail to inset cpuInfo ,sqlTest:[%v] err:[%v]", sqlText, err.Error())
		cancelFunc()
	}
	// 使用global的
	return qr, err
}

/**
 * 保存或者更新cpuusage相关信息
 */
func (c *CpuUsageRateInfo) SaveOrUpdateCpuUsageInfo() (qs *connector.QueryResults) {
	// 先查询，再修改
	timeOut, _ := strconv.Atoi(global.DB.BaseInfo.ConnTimeOut)
	ctx, canncel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeOut))
	sqlText := "select id from cpu_usage where cpu_num = ?"
	qs = global.DB.Query(ctx, sqlText, c.CpuNum)
	var id = -1

	if qs.Rows.Next() {
		err := qs.Rows.Scan(&id)
		// 如果查询不到的化会报什么错误，默认我们认为他是查询到的
		if nil != err || id == -1 {
			common.Error("Fail to select cpu_usage where cpu_num:[%v] ", c.CpuNum)
			canncel()
			return
		}
	}
	timeOut, _ = strconv.Atoi(global.DB.BaseInfo.ConnTimeOut)
	ctx, canncel = context.WithTimeout(context.Background(), time.Second*time.Duration(timeOut))
	// 存在就更新
	if id == -1 {
		sqlText = "insert into cpu_usage (cur_date,cur_time,item_name,usr, nice, sys, idle, cpu_num) values(?,?,?,?,?,?,?,?)"
		qs = global.DB.Exec(ctx, sqlText, c.CurDate, c.CurTime, c.ItemName, c.Usr, c.Nice,c.Sys, c.Idle, c.CpuNum)
		// 如果查询不到的化会报什么错误，默认我们认为他是查询到的
		if nil != qs.Err {
			common.Error("Fail to insert cpu_usage where id:[%v]", id)
			canncel()
		}
	} else {
		// 如果现在还不存在，就插入
		// 查询到之后就更新当前的记录
		timeOut, _ = strconv.Atoi(global.DB.BaseInfo.ConnTimeOut)
		ctx, canncel = context.WithTimeout(context.Background(), time.Second*time.Duration(timeOut))
		sqlText = "update cpu_usage set cur_date=?,cur_time=?,item_name=?,usr=?, nice=?, sys=?, idle=? where id=?"
		qs = global.DB.Exec(ctx, sqlText, c.CurDate, c.CurTime, c.ItemName, c.Usr, c.Nice, c.Idle, id)
		// 如果查询不到的化会报什么错误，默认我们认为他是查询到的
		if nil != qs.Err {
			common.Error("Fail to update disk where id:[%v]", id)
			canncel()
		}
	}
	return
}
