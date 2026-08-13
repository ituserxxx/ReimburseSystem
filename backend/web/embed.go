//go:build embed

package web

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var distFS embed.FS

// Register 将内嵌的前端静态资源挂载到 gin 引擎，并支持 SPA 回退：
// 当请求路径既不是 /api 接口、也找不到对应静态文件时，返回 index.html。
func Register(r *gin.Engine) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Printf("[web] 内嵌前端资源加载失败: %v", err)
		return
	}
	fileServer := http.FileServer(http.FS(sub))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 接口路径未匹配到路由时返回 JSON 404，而不是前端页面
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "接口不存在"})
			return
		}

		// 尝试作为静态文件返回
		rel := strings.TrimPrefix(path, "/")
		if rel == "" {
			rel = "index.html"
		}
		if _, err := fs.Stat(sub, rel); err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// SPA 回退：返回 index.html
		c.FileFromFS("index.html", http.FS(sub))
	})
}
