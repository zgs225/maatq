package maatq

import (
	"encoding/json"
	"errors"
	"reflect"
	"runtime"
	"sort"
	"strconv"
)

type Int8Slice []int8

func (s Int8Slice) Len() int {
	return len(s)
}

func (s Int8Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Int8Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

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

func makeRangeOfInt8(min, max int8, step int) []int8 {
	s := make([]int8, 0)
	for i := min; i <= max; i += int8(step) {
		s = append(s, i)
	}
	return s
}

func inInt8Slice(n int8, data []int8) bool {
	i := sort.Search(len(data), func(i int) bool { return data[i] >= n })
	return i < len(data) && data[i] == n
}

// 生成完整的队列名称，例如 default => maatq:default
func queueName(q string) string {
	return "maatq:" + q
}
