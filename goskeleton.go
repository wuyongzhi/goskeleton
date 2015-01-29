package goskeleton

import (
	"github.com/gin-gonic/gin"
	"strings"
	"github.com/codegangsta/inject"
)





func New(groups map[string]GroupDefine, ctx interface {}, ctxFilePath string, middlewares ... gin.HandlerFunc) (*Engine) {


	e := Engine {}
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
		if strings.TrimSpace(prefix)  == "" || strings.TrimSpace(prefix) == "/" {
			for _, route := range groupDefine.Routes {


				customHandler := wrapperCustomHandler(route.Handler, route.Form)
				e.GET(route.Pattern, customHandler )
				e.POST(route.Pattern, customHandler )
			}
		} else {
			group := e.Group(prefix, groupDefine.Middlwares...)
			for _, route := range groupDefine.Routes {

				customHandler := wrapperCustomHandler(route.Handler, route.Form)
				group.GET(route.Pattern, customHandler )
				group.POST(route.Pattern, customHandler )
			}
		}
	}

	return &e
}
