//go:build !embed

package web

import "github.com/gin-gonic/gin"

// Register 在未启用 embed 构建标签时为空实现。
// 开发模式下请使用前端 dev server（npm run dev），其已配置代理将 /api 转发到后端。
func Register(r *gin.Engine) {}
