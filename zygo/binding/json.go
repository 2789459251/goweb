package binding

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
)

type JSONBinding struct {
	DisallowUnknownFields bool
	IsValidate            bool
}

func (b *JSONBinding) Name() string {
	return "json"
}

func (b *JSONBinding) Bind(req *http.Request, obj any) error {
	//json传参内容
	body := req.Body
	if body == nil {
		return errors.New("invalid request,body is nil")
	}
	decoder := json.NewDecoder(body)
	if b.DisallowUnknownFields {
		//非结构体字段校验
		decoder.DisallowUnknownFields()
	}
	if b.IsValidate {
		err := validateRequireParam(obj, decoder)
		if err != nil {
			return err
		}
	} else {

		err := decoder.Decode(obj)
		if err != nil {
			return err
		}
	}
	return validate(obj)
}

func validateRequireParam(obj any, decoder *json.Decoder) error {
	//解析为map根据key比对
	//判断为结构体才能解析为map
	//反射
	valueOf := reflect.ValueOf(obj)
	//判断，需要为指针类型
	if valueOf.Kind() != reflect.Ptr {
		return errors.New("obj is not a pointer")
	}
	elem := valueOf.Elem().Interface()
	of := reflect.ValueOf(elem)
	switch of.Kind() {
	case reflect.Struct:
		return checkParam(of, decoder, obj)
	case reflect.Slice, reflect.Array:
		elem_ := of.Type().Elem()
		if elem_.Kind() == reflect.Struct {
			return checkParamSlice(elem_, decoder, obj)
		}
		//指针没有支持
		//if elem_.Kind() == reflect.Ptr {
		//	return checkParam(elem_.Elem(), decoder, elem_)
		//}
	default:
		return decoder.Decode(obj)
	}
	//数据在mapValue中需要的是读给obj
	return nil

}

func checkParam(of reflect.Value, decoder *json.Decoder, obj any) error {
	//解析为map
	mapValue := make(map[string]interface{})
	err := decoder.Decode(&mapValue)
	if err != nil {
		return err
	}
	for i := 0; i < of.NumField(); i++ {
		//obj的反射
		field := of.Type().Field(i)
		jsonName := field.Tag.Get("json")
		var name string = field.Name
		if jsonName != "" {
			name = jsonName

		}
		value := mapValue[name]
		required := field.Tag.Get("zygo")
		if value == nil && required == "required" {
			return errors.New("json field " + jsonName + " not exist,but should exist")
		}
	}
	objJson, _ := json.Marshal(mapValue)
	json.Unmarshal(objJson, obj)
	return nil
}

func checkParamSlice(of reflect.Type, decoder *json.Decoder, obj any) error {
	//将数据返回到切片:json -> []map -> []byte(json序列化) -> 反序列化为struct
	//需要验证每一个map符合要求
	mapValue := make([]map[string]interface{}, 0)
	err := decoder.Decode(&mapValue)
	if err != nil {
		return err
	}
	for i := 0; i < of.NumField(); i++ {
		//obj的反射
		field := of.Field(i)
		var name string = field.Name
		jsonName := field.Tag.Get("json")

		if jsonName != "" {
			name = jsonName

		}
		required := field.Tag.Get("zygo")
		for _, v := range mapValue {
			value := v[name]
			if value == nil && required == "required" {
				return errors.New("json field " + jsonName + " not exist,but should exist")
			}
		}

	}
	objJson, _ := json.Marshal(mapValue)
	json.Unmarshal(objJson, obj)
	return nil
}

// 第三方校验
func validate(obj any) error {
	return Validator_.ValidateStruct(obj)
}
