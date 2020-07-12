package monitor

import (
	"context"
	"fmt"
	task "mysql-monitor/pb/monitor/task"
)

type MySQLMonitorService struct {
	task.FlowServiceServer
}

func (m *MySQLMonitorService)SendTaskToMysqlMonitor(c context.Context,req *task.RpcSysTask) (*task.Response, error){

	// 解析request
	// printFlowInfo(request)
	switch 	req.ItemType {
	case task.Type_Sys:
		fmt.Println("接受到了sys消息")
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