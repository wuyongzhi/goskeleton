package goskeleton

import "github.com/gin-gonic/gin"

type Route struct {
	Pattern string
	Form    interface{}
	Handler gin.HandlerFunc
	Intro   string

}

type Routes []Route

type GroupDefine struct {
	Routes     Routes
	Middlwares []gin.HandlerFunc
	Intro      string
}
