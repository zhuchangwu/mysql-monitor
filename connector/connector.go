package connector

import (
	"context"
	"database/sql"
	"fmt"
	"mysql-monitor/common"

	// mysql驱动
	_ "github.com/go-sql-driver/mysql"
)

/**
 * MySQL连接相关的逻辑
 */
type Conenctor struct {
	BaseInfo BaseInfo
	DB       *sql.DB
}

/**
 * 基础信息
 */
type BaseInfo struct {
	RootUserName string
	RootPassword string
	Addr         string
	Port         string
	DBName       string
	ConnTimeOut  string //超时时间，单位秒
}

/**
 * 查询结果
 */
type QueryResults struct {
	LastInsertId int64
	EffectRow    int64
	Err          error
	Rows         *sql.Rows
}

// 读取配置文件
func (c *Conenctor) loadConfig() {
	// todo 模拟读取配置文件
	user := "root"
	password := "root"
	Addr := "localhost"
	Port := "3306"
	DBName := "test"
	ConnTimeOut := "2"
	c.BaseInfo.RootUserName = user
	c.BaseInfo.RootPassword = password
	c.BaseInfo.Addr = Addr
	c.BaseInfo.DBName = DBName
	c.BaseInfo.Port = Port
	c.BaseInfo.ConnTimeOut = ConnTimeOut
}

// 连接Mysql
func (c *Conenctor) Open() {
	// 读取配置
	c.loadConfig()
	dataSource := c.BaseInfo.RootUserName + ":" + c.BaseInfo.RootPassword + "@tcp(" + c.BaseInfo.Addr + ":" + c.BaseInfo.Port + ")/" + c.BaseInfo.DBName
	db, Err := sql.Open("mysql", dataSource)
	if Err != nil {
		common.Error("Fail to opendb dataSource:[%v] Err:[%v]", dataSource, Err.Error())
		return
	}
	db.SetMaxOpenConns(500)
	db.SetMaxIdleConns(200)
	c.DB = db
	Err = db.Ping()
	if Err != nil {
		fmt.Printf("Fail to Ping DB Err :[%v]", Err.Error())
		return
	}
}

// 查询
func (c *Conenctor) Query(ctx context.Context, sqlText string, params ...interface{}) (qr *QueryResults) {
	rows, err := c.DB.QueryContext(ctx, sqlText, params...)
	qr = &QueryResults{}
	defer HandleException()
	if err != nil {
		qr.Err = err
		common.Error("Fail to exec qurey sqlText:[%v] params:[%v] err:[%v]", sqlText, params, err)
		return
	}
	qr.Rows = rows
	return
}

// 插入、更新、删除
func (c *Conenctor) Exec(ctx context.Context, sqlText string, params ...interface{}) (qr *QueryResults) {
	qr = &QueryResults{}
	result, err := c.DB.ExecContext(ctx, sqlText, params...)
	defer HandleException()
	if err != nil {
		qr.EffectRow = 0
		qr.Err = err
		common.Error("Fail to exec qurey sqlText:[%v] params:[%v] err:[%v]", sqlText, params, err)
		return
	}
	qr.EffectRow, _ = result.RowsAffected()
	qr.LastInsertId, _ = result.LastInsertId()
	return
}

// 统一的panic处理
func HandleException() {
	if err := recover(); err != nil {
		common.Warn("Connector.go error:[%v]",err)
	}
}
