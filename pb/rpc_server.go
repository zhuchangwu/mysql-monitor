package pb

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"mysql-monitor/common"
	"mysql-monitor/dao"
	"mysql-monitor/global"
	task "mysql-monitor/pb/monitor/task"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)



type MySQLMonitorService struct {
	task.FlowServiceServer
}


// grpc 使用的响应状态码

// 构建Server
func NewRpcServer() (grpcServer *grpc.Server) {
	var serverOptions []grpc.ServerOption
	// todo 这里的时间从配置文件中读取
	serverOptions = append(serverOptions, grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: 10 * time.Second,
		Time:              10 * time.Second,
		Timeout:           10 * time.Second,
	}))
	common.Info("Prepare to build Grpc-Server")
	// 构建server
	grpcServer = grpc.NewServer(serverOptions...)
	return grpcServer
}

// 启动Server
func StartRpcServer(grpcServer *grpc.Server, waitGroup *sync.WaitGroup) {
	// todo ip、port 从配置文件中读
	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		common.Warn("Grpc-Server start fail error:[%v]", err)
		os.Exit(1)
		return
	}
	// 注册server
	task.RegisterFlowServiceServer(grpcServer, new(MySQLMonitorService))
	err = grpcServer.Serve(lis)
	if err != nil {
		common.Warn("Grpc-Server fail to exec grpcServer.Serve(lis) err:[%v]", err)
		return
	}
	waitGroup.Done()
	common.Info("Grpc Server successful")

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
			res.Status = global.RPC_RES_FAIL
			res.ResponseMsg = "Fail to update sys_monitor table err: " + qs.Err.Error()
			return res, qs.Err
		}
		res.Status = global.RPC_RES_SUCCESS
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
