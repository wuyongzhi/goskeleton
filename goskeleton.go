package goskeleton

import (
	"github.com/codegangsta/inject"
	"github.com/gin-gonic/gin"
	"strings"
)

func New(groups map[string]GroupDefine, ctx interface{}, ctxFilePath string, middlewares ...gin.HandlerFunc) *Engine {

	e := Engine{}
	e.Engine = gin.Default()
	e.Injector = inject.New()

	LoadDataFromFile(e.Injector, ctx, ctxFilePath)

	// 安装中间件
	e.Engine.Use(injectorHandler(e.Injector))
	e.Engine.Use(middlewares...)

	//
	// 初始化定义的路由器
	//
	for prefix, groupDefine := range groups {
		if strings.TrimSpace(prefix) == "" || strings.TrimSpace(prefix) == "/" {
			for _, route := range groupDefine.Routes {

				customHandler := wrapperCustomHandler(route.Handler, route.Form)
				routeMiddlwares := []gin.HandlerFunc{}

				if len(route.Middlwares) > 0 {
					routeMiddlwares = append(routeMiddlwares, route.Middlwares...)
				}
				if customHandler != nil {
					routeMiddlwares = append(routeMiddlwares, customHandler)
				}
				e.GET(route.Pattern, routeMiddlwares...)
				e.POST(route.Pattern, routeMiddlwares...)
			}
		} else {
			group := e.Group(prefix, groupDefine.Middlwares...)
			for _, route := range groupDefine.Routes {

				customHandler := wrapperCustomHandler(route.Handler, route.Form)

				routeMiddlwares := []gin.HandlerFunc{}

				if len(route.Middlwares) > 0 {
					routeMiddlwares = append(routeMiddlwares, route.Middlwares...)
				}
				if customHandler != nil {
					routeMiddlwares = append(routeMiddlwares, customHandler)
				}

				group.GET(route.Pattern, routeMiddlwares...)
				group.POST(route.Pattern, routeMiddlwares...)
			}
		}
	}

	return &e
}
