package goskeleton

import (
	"github.com/gin-gonic/gin"
	"strings"
)






func New(groups map[string]GroupDefine, ctx interface {}, ctxFilePath string) (*gin.Engine){


	LoadDataFromFile(ctx, ctxFilePath)


	g := gin.Default()

	//
	// 初始化定义的路由器
	//
	for prefix, groupDefine := range groups {
		if strings.TrimSpace(prefix)  == "" || strings.TrimSpace(prefix) == "/" {
			for _, route := range groupDefine.Routes {
				g.GET(route.Pattern, route.Handler)
				g.POST(route.Pattern, route.Handler)
			}
		} else {
			group := g.Group(prefix, groupDefine.Middlwares...)
			for _, route := range groupDefine.Routes {
				group.GET(route.Pattern, route.Handler)
				group.POST(route.Pattern, route.Handler)
			}
		}
	}

	return g
}
