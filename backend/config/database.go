package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type DBConfig struct {
	Host      string
	Port      string
	User      string
	Password  string
	DBName    string
	Charset   string
	Collation string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func InitDB() {
	cfg := DBConfig{
		Host:      getEnv("DB_HOST", "127.0.0.1"),
		Port:      getEnv("DB_PORT", "3306"),
		User:      getEnv("DB_USER", "root"),
		Password:  getEnv("DB_PASSWORD", "root"),
		DBName:    getEnv("DB_NAME", "reimburse_system"),
		Charset:   "utf8mb4",
		Collation: "utf8mb4_unicode_ci",
	}

	// 1) 先连到 MySQL 服务本身（不带具体库名），用于自动建库
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=%s&parseTime=True&loc=Local&multiStatements=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Charset)

	rootDB, err := gorm.Open(mysql.Open(rootDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect MySQL server: %v", err)
	}

	// 2) 自动创建数据库（若不存在）
	createSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET %s COLLATE %s",
		cfg.DBName, cfg.Charset, cfg.Collation,
	)
	if err := rootDB.Exec(createSQL).Error; err != nil {
		log.Fatalf("Failed to create database '%s': %v", cfg.DBName, err)
	}
	log.Printf("Database '%s' is ready", cfg.DBName)

	// 关闭临时连接
	if rootSqlDB, err := rootDB.DB(); err == nil && rootSqlDB != nil {
		rootSqlDB.Close()
	}

	// 3) 连接目标数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Database connected successfully")
}
