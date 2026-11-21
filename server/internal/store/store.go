// Package store 数据访问层：只写查询与少量业务不变量，不做参数校验。
package store

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
)

// Open 建立 MySQL 连接；autoMigrate=true 时启动期建表（开发环境姿势，
// 生产改走 server/migrations/ 版本化 SQL + golang-migrate，MIGRATE=true）。
func Open(dsn string, autoMigrate bool) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if autoMigrate {
		if err := db.AutoMigrate(&model.User{}); err != nil {
			return nil, err
		}
	}
	return db, nil
}
