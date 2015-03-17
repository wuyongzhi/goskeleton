package goskeleton

import (
	"bytes"
	"fmt"
	"github.com/codegangsta/inject"
	"github.com/gin-gonic/gin"
	"log"
	"net/http/httptest"
	"reflect"
	"strings"
)

const (
	CTX_KEY_REQUEST_INJECTOR = "REQUEST_INJECTOR"
)

func GetInjector(c *gin.Context) inject.Injector {
	var injector inject.Injector
	requestInjector, err := c.Get(CTX_KEY_REQUEST_INJECTOR)
	if err == nil && requestInjector != nil {
		injector = requestInjector.(inject.Injector)
	}
	return injector
}

//
// 将 injector 注入至 当前请求的 gin.Context 中
//
func injectorHandler(globalInjector inject.Injector) func(*gin.Context) {
	return func(c *gin.Context) {
		requestInjector := inject.New()
		requestInjector.SetParent(globalInjector)

		requestInjector.Map(c)

		c.Set(CTX_KEY_REQUEST_INJECTOR, requestInjector)
	}
}

func ParseFormHandler(formValue interface{}) func(c *gin.Context) {

	return func(c *gin.Context) {
		requestInjector := c.MustGet(CTX_KEY_REQUEST_INJECTOR)
		if requestInjector != nil {

			injector := requestInjector.(inject.Injector)
			if injector != nil {

				if formValue != nil {
					newForm := reflect.New(reflect.TypeOf(formValue))
					//					fmt.Println("newForm.Interface()", newForm.Interface())
					//					fmt.Println("newForm.Addr().Interface()", newForm.Addr().Interface())
					newForm.Elem().Set(reflect.ValueOf(formValue))

					if strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
						c.Request.ParseMultipartForm(2 << 20)
						c.Request.Header.Set("Content-Type", gin.MIMEPOSTForm)
					}

					if c.Bind(newForm.Interface()) {
						fmt.Println("c.Bind succeed:", newForm.Interface())

						injector.Map(newForm.Elem().Interface())
					} else {
						log.Println("c.Bind failed")
					}
				}
			}
		}

		c.Next()
	}

}

// 生成一个处理器，用于调用我们的自定义处理器
func wrapperCustomHandler(customHandler interface{}, formValue interface{}) func(c *gin.Context) {

	if customHandler == nil {
		panic("rawHandler was nil")
	}

	if reflect.ValueOf(customHandler).Kind() != reflect.Func {
		panic("rawHandler is not a reflect.Func")
	}

	return func(c *gin.Context) {
		requestInjector := c.MustGet(CTX_KEY_REQUEST_INJECTOR)
		if requestInjector != nil {

			injector := requestInjector.(inject.Injector)
			if injector != nil {

				if formValue != nil {

					if strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
						c.Request.ParseMultipartForm(2 << 20)
						c.Request.Header.Set("Content-Type", gin.MIMEPOSTForm)
					}

					
					newForm := reflect.New(reflect.TypeOf(formValue))
					//					fmt.Println("newForm.Interface()", newForm.Interface())
					//					fmt.Println("newForm.Addr().Interface()", newForm.Addr().Interface())
					newForm.Elem().Set(reflect.ValueOf(formValue))
					
					

					if c.Bind(newForm.Interface()) {
						//						fmt.Println("c.Bind succeed")

						injector.Map(newForm.Elem().Interface())
					} else {
						log.Println("c.Bind failed")
					}

				}

				returnValues, err := injector.Invoke(customHandler)
				if err != nil {
					panic(err)
				}

				if len(returnValues) > 0 {
					// TODO 处理返回值
				}
			}
		} else {
			// 否则继续处理下一个
			c.Next()
		}

	}
}

var RecordHttp bool

// 用于记录HTTP 请求/响应内容，以便于开发调试
func RecordHttpHandler(c *gin.Context) {

	if !RecordHttp {
		c.Next()
		return
	}

	wrapper := &responseWriter{
		recorder: httptest.NewRecorder(),
	}
	wrapper.ResponseWriter = c.Writer

	oldWriter := c.Writer

	// 重写 wrapper
	c.Writer = wrapper
	c.Next()
	c.Writer = oldWriter

	// 文本类型的响应，输出其内容
	contentType := wrapper.ResponseWriter.Header().Get("Content-Type")
	contentType = strings.ToLower(contentType)

	if strings.HasPrefix(contentType, "text") ||
		strings.Contains(contentType, "javascript") ||
		strings.Contains(contentType, "json") {

		req := c.Request

		buf := bytes.NewBuffer(make([]byte, 0, 512))

		req.ParseMultipartForm(32 << 20)
		req.ParseForm()

		buf.WriteString(fmt.Sprintf("%s\n", req.RemoteAddr))

		buf.WriteString(fmt.Sprintf("%s %s\n", req.Method, req.RequestURI))
		if len(req.UserAgent()) > 0 {
			buf.WriteString(fmt.Sprintf("User-Agent: %s\n", req.UserAgent()))
		}

		if len(req.Form) > 0 {
			buf.WriteString("\n解析的参数:\n")
			for k, v := range req.Form {
				buf.WriteString("\t")
				buf.WriteString(k)
				buf.WriteString("=")
				buf.WriteString(strings.Join(v, ","))
				buf.WriteString("\n")
			}
		}

		if req.MultipartForm != nil {
			if len(req.MultipartForm.Value) > 0 {
				buf.WriteString("\nMultipartForm:\n")
				for k, v := range req.MultipartForm.Value {
					buf.WriteString("\t")
					buf.WriteString(k)
					buf.WriteString("=")
					buf.WriteString(strings.Join(v, ","))
					buf.WriteString("\n")
				}
			}

			if len(req.MultipartForm.File) > 0 {

				buf.WriteString("MultipartFile:\n")
				for k, v := range req.MultipartForm.File {
					if v != nil {
						buf.WriteString("\t")
						buf.WriteString(k)
						buf.WriteString("=")
						for i, e := range v {
							if i > 0 {
								buf.WriteString(",")
							}
							buf.WriteString("filename:" + e.Filename)

						}

						buf.WriteString("\n")
					}
				}
			}
		}

		buf.WriteString("\nResponseBody:\n")
		buf.Write(wrapper.recorder.Body.Bytes())

		fmt.Println(buf.String())
	}

}

type Middlwares struct {
	handlers []gin.HandlerFunc
}

func NewMiddlwares(handlers ...gin.HandlerFunc) *Middlwares {
	return &Middlwares{
		handlers: handlers,
	}
}

func (me *Middlwares) BindForm(formValue interface{}) *Middlwares {
	bindHandler := func(c *gin.Context) {
		requestInjector := c.MustGet(CTX_KEY_REQUEST_INJECTOR)
		if requestInjector != nil {

			injector := requestInjector.(inject.Injector)
			if injector != nil {

				if formValue != nil {
					newForm := reflect.New(reflect.TypeOf(formValue))
					newForm.Elem().Set(reflect.ValueOf(formValue))

					if strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
						c.Request.ParseMultipartForm(2 << 20)
						c.Request.Header.Set("Content-Type", gin.MIMEPOSTForm)
					}

					if c.Bind(newForm.Interface()) {
						fmt.Println("c.Bind succeed:", newForm.Interface())

						injector.Map(newForm.Elem().Interface())
					} else {
						log.Println("c.Bind failed")
					}
				}
			}
		}

		c.Next()
	}

	me.handlers = append(me.handlers, bindHandler)

	return me
}


func (me *Middlwares) Bind(formValue interface{}) *Middlwares {
	bindHandler := func(c *gin.Context) {
		requestInjector := c.MustGet(CTX_KEY_REQUEST_INJECTOR)
		if requestInjector != nil {

			injector := requestInjector.(inject.Injector)
			if injector != nil {

				if formValue != nil {
					newForm := reflect.New(reflect.TypeOf(formValue))
					newForm.Elem().Set(reflect.ValueOf(formValue))

					if strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
						c.Request.ParseMultipartForm(2 << 20)
						c.Request.Header.Set("Content-Type", gin.MIMEPOSTForm)
					}

					if c.Bind(newForm.Interface()) {
						fmt.Println("c.Bind succeed:", newForm.Interface())

						injector.Map(newForm.Elem().Interface())
					} else {
						log.Println("c.Bind failed")
					}
				}
			}
		}

		c.Next()
	}
	me.handlers = append(me.handlers, bindHandler)

	return me
}
