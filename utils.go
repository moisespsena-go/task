package task

import "reflect"

func validStoper(s Stoper) bool {
	rv := reflect.Indirect(reflect.ValueOf(s))
	return rv.IsValid() && !rv.IsZero()
}
