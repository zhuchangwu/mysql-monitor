package connector

import (
	"context"
	"database/sql"
	"fmt"
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
type QueryResult struct {
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
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		fmt.Printf("Fail to opendb dataSource:[%v] err:[%v]", dataSource, err.Error())
		return
	}
	db.SetMaxOpenConns(500)
	db.SetMaxIdleConns(200)
	c.DB = db
	err = db.Ping()
	if err != nil {
		fmt.Printf("Fail to Ping DB err :[%v]", err.Error())
		return
	}
}

// 查询
func (c *Conenctor) Query(ctx *context.Context, sqlText string, params ...interface{}) (qr *QueryResult, err error) {
	return &QueryResult{}, nil
}

// 插入、更新、删除
func (c *Conenctor) Exec(ctx context.Context, sqlText string, params ...interface{}) (affectedRow int, err error) {
	result, err := c.DB.ExecContext(ctx,sqlText, params...)
	if err!=nil{
		affectedRow = 0
		return
	}
	affected, err := result.RowsAffected()
	return int(affected), err
}
