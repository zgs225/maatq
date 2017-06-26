package maatq

import (
	"reflect"
	"testing"
)

func TestMakeRangeOfInt8(t *testing.T) {
	s := makeRangeOfInt8(1, 5, 1)
	e := []int8{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(s, e) {
		t.Error("makeRangeOfInt8 error, step 1")
	}

	s = makeRangeOfInt8(1, 5, 2)
	e = []int8{1, 3, 5}
	if !reflect.DeepEqual(s, e) {
		t.Error("makeRangeOfInt8 error, step 2")
	}

	s = makeRangeOfInt8(1, 5, 100)
	e = []int8{1}
	if !reflect.DeepEqual(s, e) {
		t.Error("makeRangeOfInt8 error, step 2")
	}
}
