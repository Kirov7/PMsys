package gorms

import (
	"context"
	"fmt"
	"github.com/Kirov7/project-user/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _db *gorm.DB

func init() {
	//配置MySQL连接参数
	username := config.AppConf.MysqlConfig.Username //账号
	password := config.AppConf.MysqlConfig.Password //密码
	host := config.AppConf.MysqlConfig.Host         //数据库地址，可以是Ip或者域名
	port := config.AppConf.MysqlConfig.Port         //数据库端口
	Dbname := config.AppConf.MysqlConfig.Db         //数据库名
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
	var err error
	_db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("连接数据库失败, error=" + err.Error())
	}
}

func GetDB() *gorm.DB {
	return _db
}

type GormConn struct {
	DB *gorm.DB
	tx *gorm.DB
}

func New() *GormConn {
	return &GormConn{DB: GetDB()}
}

func NewTransaction() *GormConn {
	return &GormConn{DB: GetDB(), tx: GetDB()}
}

func (g *GormConn) Session(ctx context.Context) *gorm.DB {
	return g.DB.Session(&gorm.Session{Context: ctx})
}

func (g *GormConn) Begin() {
	g.tx = GetDB().Begin()
}

func (g *GormConn) Rollback() {
	g.tx.Rollback()
}

func (g *GormConn) Commit() {
	g.tx.Commit()
}

func (g *GormConn) Tx(ctx context.Context) *gorm.DB {
	return g.tx.WithContext(ctx)
}
