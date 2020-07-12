package monitor

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"mysql-monitor/common"
	task "mysql-monitor/pb/monitor/task"
	"net"
	"os"
	"sync"
	"time"
)

type MySQLMonitorService struct {
	task.FlowServiceServer
}

// 构建Server
func NewRpcServer() (grpcServer *grpc.Server){
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
func StartRpcServer(grpcServer *grpc.Server,waitGroup *sync.WaitGroup) {
	// todo ip、port 从配置文件中读
	lis, err:= net.Listen("tcp", ":8082")
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
