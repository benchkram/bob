package reflectutil

import "reflect"

// CountExportedFields returns number of exported fields of a
func CountExportedFields(a interface{}) int {
	rt := reflect.TypeOf(a) // take type
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem() // use Elem to get the pointed-to-type
	}
	if rt.Kind() == reflect.Slice {
		rt = rt.Elem() // use Elem to get type of slice's element
	}
	if rt.Kind() == reflect.Ptr { // handle input of type like []*StructType
		rt = rt.Elem() // use Elem to get the pointed-to-type
	}
	if rt.Kind() != reflect.Struct {
		return -1
	}

	var count int
	for _, f := range reflect.VisibleFields(rt) {
		if f.IsExported() {
			count++
		}
	}

	return count
}
