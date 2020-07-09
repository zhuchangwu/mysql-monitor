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
type Tasks struct {
	ID       int
	CurDate  time.Time // 格式 2020-1-1 10:20:21
	CurTime  string    // 格式 10:20:21
	ItemName string    // 监控项名称
	Total    int       // 总任务数
	Running  int       // 正在运行的进程数
	Sleeping int       // 睡眠的进程数
	Stoped   int       // 暂停的进程数
	Zombie   int       // 僵尸进程数
}

func NewTasksInfo(date time.Time, time string, itemName string, total, running, sleeping, stoped, zombie int) *Tasks {
	return &Tasks{
		CurDate:  date,
		CurTime:  time,
		ItemName: itemName,
		Total:    total,
		Running:  running,
		Sleeping: sleeping,
		Stoped:   stoped,
		Zombie:   zombie,
	}
}

/**
 * 插入一条数据
 */
func (t *Tasks) InsertOneCord() (qr *connector.QueryResults, err error) {
	connector := global.DB
	timeOut, _ := strconv.Atoi(connector.BaseInfo.ConnTimeOut)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(timeOut)*time.Second)
	sqlText := "insert into sys_tasks (cur_date, cur_time, item_name,total, running, sleeping, stoped, zombie) values (?,?,?,?,?,?,?,?);"
	qr = connector.Exec(ctx, sqlText, t.CurDate, t.CurTime, t.ItemName, t.Total,t.Running,t.Sleeping,t.Stoped,t.Zombie)
	if nil != qr.Err {
		common.Error("Fail to insert sys_tasks ,sqlText:[%v] err:[%v]", sqlText, err.Error())
		cancelFunc()
	}
	// 使用global的
	return qr, err
}
