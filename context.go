package goskeleton

import (
	"reflect"
	"github.com/codegangsta/inject"
	"github.com/BurntSushi/toml"
	"fmt"
	"github.com/gin-gonic/gin"
)


type Engine struct {
	*gin.Engine
	Injector inject.Injector
}


func isStruct(t reflect.Type) bool {
	for t != nil {
		switch t.Kind()  {
		case reflect.Struct:
			return true
		case reflect.Ptr:
			t = t.Elem()
		case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
			return false
		default:
			return false
		}
	}
	return false
}

// 递归注入
func recursiveInject(injector inject.Injector, value interface {}) error {

	var err error
	err = injector.Apply(value)
	if err != nil {
		return err
	}

	v := reflect.ValueOf(value)


	for v.Kind() == reflect.Ptr {
		v = v.Elem()

	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {

		f := v.Field(i)
		structField := t.Field(i)

		// 指针类型，或有注入标记的，不再处理
		if f.Kind() == reflect.Ptr && structField.Tag == "inject" {
			continue
		} else if f.CanSet() && isStruct(structField.Type){

			err = recursiveInject(injector, f.Addr().Interface())

			if err != nil {
				break
			}
		}
	}

	return err
}





func LoadDataFromFile(injector inject.Injector, data interface {}, ctxFilePath string) {

	if data == nil {
		return
	}

	_, err := toml.DecodeFile(ctxFilePath, data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	ctxValue := reflect.ValueOf(data)
	for ctxValue.Kind() == reflect.Ptr {
		ctxValue = ctxValue.Elem()
	}

	if ctxValue.Kind() == reflect.Struct {
		for fieldIndex:=0; fieldIndex < ctxValue.NumField(); fieldIndex++ {
			fieldValue := ctxValue.Field(fieldIndex)
			fieldType := fieldValue.Type()

			if isStruct(fieldType)  {
				injector.Map(fieldValue.Interface())
			}
		}
	}



	//
	//	递归注入
	//
	err = recursiveInject(injector, data)
	if err != nil {
		fmt.Println("recursiveInject error:", err)
	}

}
