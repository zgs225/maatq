package main

import (
	"encoding/json"
	"errors"
	"reflect"
	"runtime"
	"strconv"
)

func AtoString(val interface{}) (string, error) {
	sval, ok := val.(string)
	if ok {
		return sval, nil
	}

	fval, ok := val.(float64)
	if ok {
		return strconv.FormatInt(int64(fval), 10), nil
	}

	bval, ok := val.(bool)
	if ok {
		return strconv.FormatBool(bval), nil
	}

	jval, ok := val.(map[string]interface{})
	if ok {
		tmpVal, err := json.Marshal(jval)
		if err == nil {
			return string(tmpVal), nil
		}
		return "", err
	}

	return "", errors.New("Unknown type")
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
