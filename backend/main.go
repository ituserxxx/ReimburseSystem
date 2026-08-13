package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"reimburse-system/config"
	"reimburse-system/models"
	"reimburse-system/routers"
)

// loadEnv 优先加载可执行文件同目录下的 .env，找不到再回退当前工作目录，
// 均失败则静默跳过、使用系统环境变量。
func loadEnv() {
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		_ = godotenv.Load(filepath.Join(dir, ".env"))
	}
	_ = godotenv.Load()
}

func main() {
	// 加载 .env 文件（文件不存在时静默跳过，使用系统环境变量）
	loadEnv()

	// 设置 Gin 模式
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		os.Setenv("GIN_MODE", mode)
	}

	// 初始化数据库
	config.InitDB()

	// 自愈：确保 locations 列存在（兼容旧表，避免 GORM 在已有数据的表上
	// 添加 NOT NULL 列导致迁移失败）。需放在 AutoMigrate 之前。
	ensureLocationColumn(config.DB)

	// 自动迁移表结构
	if err := models.AutoMigrate(config.DB); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed")

	// 启动服务
	r := routers.SetupRouter()
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// ensureLocationColumn 检测 reimbursements 表是否已有 location 列，
// 没有则补加（带默认值，兼容表中已有数据的情况）。
func ensureLocationColumn(db *gorm.DB) {
	// 新库时 reimbursements 表还不存在，AutoMigrate 会直接建好带 location 的表，无需补列
	var tableCount int64
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'reimbursements'").Scan(&tableCount)
	if tableCount == 0 {
		return
	}
	var colCount int64
	db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'reimbursements' AND column_name = 'location'").Scan(&colCount)
	if colCount == 0 {
		if err := db.Exec("ALTER TABLE reimbursements ADD COLUMN location varchar(255) NOT NULL DEFAULT '' COMMENT '地点'").Error; err != nil {
			log.Printf("Warning: 补齐 location 列失败: %v", err)
		} else {
			log.Println("已为 reimbursements 表补齐 location 列")
		}
	}
}
