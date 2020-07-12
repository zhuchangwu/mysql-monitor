package monitor

import (
	"context"
	"fmt"
	"mysql-monitor/common"
	"mysql-monitor/dao"
	"mysql-monitor/global"
	task "mysql-monitor/pb/monitor/task"
	"strconv"
)

type MySQLMonitorService struct {
	task.FlowServiceServer
}

// 处理请求
func (m *MySQLMonitorService) SendTaskToMysqlMonitor(c context.Context, req *task.RpcSysTask) (*task.Response, error) {

	// 解析request
	// printFlowInfo(request)
	switch req.ItemType {
	case task.Type_Sys:
		// 响应的对象
		res := &task.Response{}
		common.Info("Recieve client msg ， msyType:[%v]", task.Type_Sys)
		// 对于操作系统的监控指标来说，beego发送过来的任务目的时更新阙值或者是更新采集周期
		monitor := dao.NewMonitor(req.Sysmsg.ItemName)
		monitor.SCycle, _ = strconv.Atoi(req.Sysmsg.Cycle)
		monitor.Threshold = req.Sysmsg.Threshold
		qs := monitor.UpdateMonitorCycleOrThresholdInfo()
		if qs.Err != nil {
			res.Status = common.RPC_RES_FAIL
			res.ResponseMsg = "Fail to update sys_monitor table err: " + qs.Err.Error()
			return res, qs.Err
		}
		res.Status = common.RPC_RES_SUCCESS
		res.ResponseMsg = "Update sys_monitor table successful"
		//将消息写入chan，Sys.go中有协程消费这个chan
		global.SysHotLoadChan <- req.Sysmsg
		// 将消息中的内容更写进Sys监控使用chan
		return res, qs.Err
	case task.Type_App:
		fmt.Println("接受到了app消息")
	case task.Type_MySQL:
		fmt.Println("接受到了mysql消息")

	}

	// 返回res
	r := &task.Response{
		Status:      1,
		ResponseMsg: "successfully",
	}
	return r, nil
}
