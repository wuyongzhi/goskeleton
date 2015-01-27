package goskeleton

import (
	"net/http/httptest"
	"github.com/gin-gonic/gin"
)


type responseWriter struct {
	gin.ResponseWriter
	recorder *httptest.ResponseRecorder
}

func (me *responseWriter) WriteHeader(s int) {
	me.ResponseWriter.WriteHeader(s)
	me.recorder.WriteHeader(s)
}

func (me *responseWriter) Write(bytes []byte) (int, error) {
	i, e := me.ResponseWriter.Write(bytes)
	me.recorder.Write(bytes)
	return i, e
}

