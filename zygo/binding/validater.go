package binding

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
	"sync"
)

type SliceValidationError []error

var Validator_ StructValidator = &defaultStructValidator{}

//结构体的验证接口

type StructValidator interface {
	//结构体验证，
	ValidateStruct(any) error
	//返回对应使用的验证器
	Engine() any
}
type defaultStructValidator struct {
	one      sync.Once //单例
	validate *validator.Validate
}

func (d *defaultStructValidator) ValidateStruct(obj any) error {
	of := reflect.ValueOf(obj)

	switch of.Kind() {
	case reflect.Ptr:
		return d.ValidateStruct(of.Elem().Interface())
	case reflect.Struct:
		return d.validateStruct(of)
	case reflect.Slice, reflect.Array:
		count := of.Len()
		var errs SliceValidationError
		for i := 0; i < count; i++ {
			err := d.validateStruct(of.Index(i).Interface())
			if err != nil {
				errs = append(errs, err)
			}
		}
		return errs
	}
	return validator.New().Struct(obj)
}

func validateStruct(obj any) error {
	return validator.New().Struct(obj)
}

func (err SliceValidationError) Error() string {
	n := len(err)
	switch n {
	case 0:
		return ""
	default:
		var b strings.Builder
		if err[0] != nil {
			fmt.Fprintf(&b, "[%d]:%s,", 0, err[0].Error())
		}
		if n > 1 {
			for i := 1; i < n; i++ {
				if err[i] != nil {
					b.WriteString("\n")
					fmt.Fprintf(&b, "[%d]:%s,", i, err[i].Error())
				}
			}
		}
		return b.String()
	}

}

func (d *defaultStructValidator) Engine() any {
	d.lazyInit()
	return d.validate
}
func (d *defaultStructValidator) lazyInit() {
	d.one.Do(func() {
		//单例-懒汉式
		d.validate = validator.New()
	})
}

func (d *defaultStructValidator) validateStruct(obj any) error {
	d.lazyInit()
	return d.validate.Struct(obj)
}
