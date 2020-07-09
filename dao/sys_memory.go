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
 * Memory 相关的信息
 */
type Memory struct {
	ID       int
	CurDate  time.Time // 格式 2020-1-1 10:20:21
	CurTime  string    // 格式 10:20:21
	ItemName string    // 监控项名称
	Total    int       // 总内存数
	Used     int       // 已使用
	Free     int       // 剩余可用内存
	Buff     int       // OS缓存
}

func NewMemory(curDate time.Time, curTime, itemName string, total, used, free, buff int) *Memory {
	return &Memory{
		CurDate :curDate,
		CurTime  :curTime,
		ItemName :itemName,
		Total    :total,
		Used     :used,
		Free     :free,
		Buff     :buff,
	}
}


/**
 * 插入一条数据
 */
func (m *Memory) InsertOneCord() (qr *connector.QueryResults, err error) {
	connector := global.DB
	timeOut, _ := strconv.Atoi(connector.BaseInfo.ConnTimeOut)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(timeOut)*time.Second)
	sqlText := "insert into sys_memory (cur_date, cur_time, item_name, total, used, free, buff) values (?,?,?,?,?,?,?);"
	qr = connector.Exec(ctx, sqlText, m.CurDate, m.CurTime, m.ItemName, m.Total, m.Used, m.Free, m.Buff)
	if nil != qr.Err {
		common.Error("Fail to insert sys_memoryInfo ,sqlTest:[%v] err:[%v]", sqlText, err.Error())
		cancelFunc()
	}
	return qr, err
}
