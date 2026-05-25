package db

import (
	"context"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open 使用 DATABASE_URL 连接 PostgreSQL 并返回 GORM 实例。
func Open(ctx context.Context, connString string) (*gorm.DB, error) {
	_ = ctx
	gdb, err := gorm.Open(postgres.Open(connString), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}
	return gdb, nil
}
