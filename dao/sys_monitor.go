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
 * 监控项&报警
 */
type Monitor struct {
	Id                  int
	ItemName            string
	Threshold           string // 报警的阈值
	ThresholdNum        int    // 累计超过报警的阈值多少次后报警
	CurrentThresholdNum int    // 当前累计超过报警的阈值次数（当监控项恢复正常后，这个值被摸为0）
	ThresholdHistoryNum int    // 历史累计超过报警的阈值次数（只会累加，不会被抹掉）
	UrgentAction        string // 报警执行的动作
}

func NewMonitor(itemName string) *Monitor {
	return &Monitor{
		ItemName: itemName,
	}
}

/**
 * 有则更新、无则插入 CPU 监控项
 */
func (m *Monitor) SaveOrUpdateMonitorInfo() (qs *connector.QueryResults) {
	// 先查询，再修改
	timeOut, _ := strconv.Atoi(global.DB.BaseInfo.ConnTimeOut)
	ctx, canncel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeOut))
	sqlText := "select id from sys_monitor where item_name = ?"
	qs = global.DB.Query(ctx, sqlText, m.ItemName)
	var id = -1

	if qs.Rows.Next() {
		err := qs.Rows.Scan(&id)
		// 如果查询不到的化会报什么错误，默认我们认为他是查询到的
		if nil != err || id == -1 {
			common.Error("Fail to select sys_monitor where item_name:[%v]", m.ItemName)
			canncel()
			return
		}

	}

	// 查询到之后就更新当前的记录
	timeOut, _ = strconv.Atoi(global.DB.BaseInfo.ConnTimeOut)
	ctx, canncel = context.WithTimeout(context.Background(), time.Second*time.Duration(timeOut))
	sqlText = "update sys_monitor set current_threshold_num=current_threshold_num+1,threshold_history_num=threshold_history_num+1 where id=?"
	qs = global.DB.Exec(ctx, sqlText, id)
	// 如果查询不到的化会报什么错误，默认我们认为他是查询到的
	if nil != qs.Err {
		common.Error("Fail to update sys_monitor where id:[%v]", id)
		canncel()
	}
	return
}
