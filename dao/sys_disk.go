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
 * 磁盘空间
 * 	Filesystem      Size  Used 		Avail		Use%      MountedOn
 *  文件系统			大小  已使用 	剩余可用		使用率	  挂在路径
 */
type Disk struct {
	ID         int
	CurDate    time.Time // 格式 2020-1-1 10:20:21
	CurTime    string    // 格式 10:20:21
	ItemName   string    // 监控项名称
	FileSystem string    // 文件系统
	Size       string    // 总大小
	Used       string    // 已使用
	Avail      string    // 已使用
	Usage      string    // 使用率
	MountedOn  string    // 挂在路径
}

func NewDiskInfo(date time.Time, time string, itemName string, filesystem, size, use, avail, usage, mountedOn string) *Disk {
	return &Disk{
		CurDate:    date,
		CurTime:    time,
		ItemName:   itemName,
		FileSystem: filesystem, // 文件系统
		Size:       size,       // 总大小
		Used:       use,        // 剩余可用
		Avail:      avail,      // 已使用
		Usage:      usage,      // 挂在路径
		MountedOn:  mountedOn,  // 挂在路径
	}
}

/**
 * 有则更新、无则插入 CPU 监控项
 */
func (d *Disk) SaveOrUpdateDiskInfo() (qs *connector.QueryResults) {
	// 先查询，再修改
	timeOut, _ := strconv.Atoi(global.DB.BaseInfo.ConnTimeOut)
	ctx, canncel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeOut))
	sqlText := "select id from sys_disk where file_system = ? and mounted_on = ?"
	qs = global.DB.Query(ctx, sqlText, d.FileSystem, d.MountedOn)
	var id = -1

	if qs.Rows.Next() {
		err := qs.Rows.Scan(&id)
		// 如果查询不到的化会报什么错误，默认我们认为他是查询到的
		if nil != err || id == -1 {
			common.Error("Fail to select diskInfo where fileSystem:[%v] moundtedOn:[%v]", d.FileSystem, d.MountedOn)
			canncel()
			return
		}
	}
	timeOut, _ = strconv.Atoi(global.DB.BaseInfo.ConnTimeOut)
	ctx, canncel = context.WithTimeout(context.Background(), time.Second*time.Duration(timeOut))
	// 存在就更新
	if id == -1 {
		sqlText = "insert into sys_disk (cur_date,cur_time,item_name,file_system,size,used,avail, `usage`,mounted_on) values(?,?,?,?,?,?,?,?,?)"
		qs = global.DB.Exec(ctx, sqlText, d.CurDate, d.CurTime, d.ItemName, d.FileSystem, d.Size, d.Used, d.Avail, d.Usage, d.MountedOn)
		// 如果查询不到的化会报什么错误，默认我们认为他是查询到的
		if nil != qs.Err {
			common.Error("Fail to insert sys_disk where id:[%v]", id)
			canncel()
		}
	} else {
		// 如果现在还不存在，就插入
		// 查询到之后就更新当前的记录
		timeOut, _ = strconv.Atoi(global.DB.BaseInfo.ConnTimeOut)
		ctx, canncel = context.WithTimeout(context.Background(), time.Second*time.Duration(timeOut))
		sqlText = "update sys_disk set cur_date=?,cur_time=?,item_name=?,file_system=?,size=?,used=?,avail=?, `usage`=?,mounted_on=? where id=?"
		qs = global.DB.Exec(ctx, sqlText, d.CurDate, d.CurTime, d.ItemName, d.FileSystem, d.Size, d.Used, d.Avail, d.Usage, d.MountedOn, id)
		// 如果查询不到的化会报什么错误，默认我们认为他是查询到的
		if nil != qs.Err {
			common.Error("Fail to update sys_disk where id:[%v]", id)
			canncel()
		}
	}
	return
}
