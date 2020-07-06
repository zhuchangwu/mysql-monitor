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
 * IO 相关的信息
 * 	1） 流经网卡的IO
 * 	2） 磁盘随机的IO
 * 	3） 储存IO
 */
type IO struct {
	ID        int
	CurDate   time.Time // 格式 2020-1-1 10:20:21
	CurTime   string    // 格式 10:20:21
	ItemName  string    // 监控项名称
	ReadRate  string    // 每秒读取
	WriteRate string    // 每秒写入
}

func NewIOInfo(date time.Time, time string, itemName string, readRate, writeRate string) *IO {
	return &IO{
		CurDate:   date,
		CurTime:   time,
		ItemName:  itemName,
		ReadRate:  readRate,  // 每秒读取
		WriteRate: writeRate, // 每秒写入
	}
}

/**
 * 插入一条数据
 */
func (i *IO) InsertOneCord() (qr *connector.QueryResults, err error) {
	connector := global.DB
	timeOut, _ := strconv.Atoi(connector.BaseInfo.ConnTimeOut)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(timeOut)*time.Second)
	sqlText := "insert into io (cur_date, cur_time, item_name,read_rate,write_rate ) values (?,?,?,?,?);"
	qr = connector.Exec(ctx, sqlText, i.CurDate, i.CurTime, i.ItemName, i.ReadRate, i.WriteRate)
	if nil != qr.Err {
		common.Error("Fail to insert ioInfo ,sqlText:[%v] err:[%v]", sqlText, err.Error())
		cancelFunc()
	}
	// 使用global的
	return qr, err
}
