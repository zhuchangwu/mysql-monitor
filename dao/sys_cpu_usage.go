package dao

import (
	"context"
	"mysql-monitor/common"
	"mysql-monitor/connector"
	"mysql-monitor/global"
	"strconv"
	"time"
)

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
 * 保存或者更新cpuusage相关信息
 */
func (c *CpuUsageRateInfo) SaveOrUpdateCpuUsageInfo() (qs *connector.QueryResults) {
	// 先查询，再修改
	timeOut, _ := strconv.Atoi(global.DB.BaseInfo.ConnTimeOut)
	ctx, canncel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeOut))
	sqlText := "select id from sys_cpu_usage where cpu_num = ?"
	qs = global.DB.Query(ctx, sqlText, c.CpuNum)
	var id = -1

	if qs.Rows.Next() {
		err := qs.Rows.Scan(&id)
		// 如果查询不到的化会报什么错误，默认我们认为他是查询到的
		if nil != err || id == -1 {
			common.Error("Fail to select sys_cpu_usage where cpu_num:[%v] ", c.CpuNum)
			canncel()
			return
		}
	}
	timeOut, _ = strconv.Atoi(global.DB.BaseInfo.ConnTimeOut)
	ctx, canncel = context.WithTimeout(context.Background(), time.Second*time.Duration(timeOut))
	// 存在就更新
	if id == -1 {
		sqlText = "insert into sys_cpu_usage (cur_date,cur_time,item_name,usr, nice, sys, idle, cpu_num) values(?,?,?,?,?,?,?,?)"
		qs = global.DB.Exec(ctx, sqlText, c.CurDate, c.CurTime, c.ItemName, c.Usr, c.Nice,c.Sys, c.Idle, c.CpuNum)
		// 如果查询不到的化会报什么错误，默认我们认为他是查询到的
		if nil != qs.Err {
			common.Error("Fail to insert sys_cpu_usage where id:[%v]", id)
			canncel()
		}
	} else {
		// 如果现在还不存在，就插入
		// 查询到之后就更新当前的记录
		timeOut, _ = strconv.Atoi(global.DB.BaseInfo.ConnTimeOut)
		ctx, canncel = context.WithTimeout(context.Background(), time.Second*time.Duration(timeOut))
		sqlText = "update sys_cpu_usage set cur_date=?,cur_time=?,item_name=?,usr=?, nice=?, sys=?, idle=? where id=?"
		qs = global.DB.Exec(ctx, sqlText, c.CurDate, c.CurTime, c.ItemName, c.Usr, c.Nice, c.Idle, id)
		// 如果查询不到的化会报什么错误，默认我们认为他是查询到的
		if nil != qs.Err {
			common.Error("Fail to update sys_cpu_usage where id:[%v]", id)
			canncel()
		}
	}
	return
}

