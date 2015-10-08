package confobject

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func sliceFromStrings(value interface{}, type_ interface{}) (vs interface{}, err error) {
	switch value.(type) {
	case string:
		sv := value.(string)
		strs := strings.Split(sv, ",")
		switch type_.(type) {
		case string:
			return strs, err
		case int:
			var ivs []int
			for _, s := range strs {
				var iv int64
				iv, err = strconv.ParseInt(s, 10, 64)
				if err != nil {
					err = fmt.Errorf("Cannot parse %v", s)
					return
				}
				ivs = append(ivs, int(iv))
			}
			return ivs, err
		case float64:
			var fvs []float64
			for _, s := range strs {
				var fv float64
				fv, err = strconv.ParseFloat(s, 64)
				if err != nil {
					err = fmt.Errorf("Cannot parse %v", s)
					return
				}
				fvs = append(fvs, fv)
			}
			return fvs, err
		default:
			err = fmt.Errorf("Cannot parse %v", value)
			return
		}
	case []string:
		// var values interface{}
		for _, sval := range value.([]string) {
			val, err := sliceFromStrings(sval, type_)
			if err != nil {
				return vs, err
			}
			// vs = append(vs, val)

			if reflect.TypeOf(vs) == nil {
				vs = val
			} else {
				vs = reflect.AppendSlice(
					reflect.ValueOf(vs),
					reflect.ValueOf(val),
				).Interface()
			}

		}
	default:
		err = fmt.Errorf("Cannot parse %v", value)
	}
	return
}

func boolFromInterface(value interface{}) (v bool, err error) {
	if reflect.ValueOf(value).Kind() == reflect.Slice {
		v, err = boolFromInterface(reflect.ValueOf(value).Index(0).Interface())
		return
	}
	switch value.(type) {
	case string:
		bs := strings.ToLower(value.(string))
		fmt.Println("+++", bs)
		switch bs {
		case "t", "true":
			v = true
		case "f", "false":
			v = false
		default:
			err = fmt.Errorf("Invalid string representation of a bool: %v", bs)
		}
	default:
		err = fmt.Errorf("Cannot parse %v", value)
	}
	fmt.Println("+++", err)
	return
}

func intFromInterface(value interface{}) (v int64, err error) {
	if reflect.ValueOf(value).Kind() == reflect.Slice {
		// fmt.Println("intFromInterface got slice", reflect.ValueOf(value).Index(0).Interface())
		v, err = intFromInterface(reflect.ValueOf(value).Index(0).Interface())
		return
	}
	switch value.(type) {
	case int:
		v = int64(value.(int))
	case int8:
		v = int64(value.(int8))
	case int16:
		v = int64(value.(int16))
	case int32:
		v = int64(value.(int32))
	case string:
		v, err = strconv.ParseInt(value.(string), 10, 64)
	default:
		err = fmt.Errorf("Cannot parse %v", value)
	}

	return
}

func floatFromInterface(value interface{}) (v float64, err error) {
	if reflect.ValueOf(value).Kind() == reflect.Slice {
		// fmt.Println("floatFromInterface got slice", reflect.ValueOf(value).Index(0).Interface())
		v, err = floatFromInterface(reflect.ValueOf(value).Index(0).Interface())
		return
	}
	switch value.(type) {
	case float32:
		v = float64(value.(float32))
	case string:
		v, err = strconv.ParseFloat(value.(string), 64)
	default:
		err = fmt.Errorf("Cannot parse %v", value)
	}

	return
}
