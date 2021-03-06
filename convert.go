package convert

import (
	"errors"
	"reflect"
)

// Fn converts from into to.
type Fn func(from interface{}) (interface{}, error)

const (
	tagName = "convert"
)

var converters = map[reflect.Type]map[reflect.Type]Fn{}

// Register registers a callback fn which is called whenever a conversion
// has to happen from type 'from' to type 'to'.
func Register(from, to reflect.Type, fn Fn) {
	if _, ok := converters[from]; !ok {
		c := map[reflect.Type]Fn{}
		converters[from] = c
	}

	converters[from][to] = fn
}

// Convert takes in a strut from and a pointer to struct to and performs
// the conversion.
func Convert(from, to interface{}) error {
	toValue := indirect(reflect.ValueOf(to))
	if !toValue.CanAddr() {
		return errors.New("copy to value is unaddressable")
	}

	fromValue := indirect(reflect.ValueOf(from))
	if !fromValue.IsValid() {
		return nil
	}

	fromType := indirectType(reflect.TypeOf(from))
	toType := indirectType(reflect.TypeOf(to))
	if fromType.Kind() == reflect.Slice {
		if toType.Kind() != reflect.Slice {
			return errors.New("cannot copy slice to a non-slice")
		}
		length := fromValue.Len()
		for i := 0; i < length; i++ {
			fromElem := indirect(fromValue.Index(i))
			toElem := indirect(reflect.New(indirectType(toType.Elem())))

			fields := getFields(fromElem.Type())
			err := setElem(fields, fromElem, toElem)
			if err != nil {
				return err
			}
			if toElem.Addr().Type().AssignableTo(toValue.Type().Elem()) {
				toValue.Set(reflect.Append(toValue, toElem.Addr()))
			} else if toElem.Type().AssignableTo(toValue.Type().Elem()) {
				toValue.Set(reflect.Append(toValue, toElem))
			}
		}
	} else {
		fields := getFields(fromType)
		err := setElem(fields, fromValue, toValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func setElem(fields []reflect.StructField, fromValue, toValue reflect.Value) error {
	for _, field := range fields {
		tag := field.Tag.Get(tagName)
		name := field.Name
		if tag != "" {
			name = tag
		}

		fromField := fromValue.FieldByName(field.Name)
		toField := toValue.FieldByName(name)

		if !fromField.IsValid() || !toField.IsValid() || !toField.CanSet() {
			continue
		}

		if err := set(toField, fromField); err != nil {
			return err
		}
	}
	return nil
}

func getFields(t reflect.Type) []reflect.StructField {
	var fields []reflect.StructField

	for i := 0; i < t.NumField(); i++ {
		v := t.Field(i)
		fields = append(fields, v)
	}

	return fields
}

func indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func indirectType(reflectType reflect.Type) reflect.Type {
	for reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	return reflectType
}

func set(to, from reflect.Value) error {
	if !from.IsValid() {
		return nil
	}
	if to.Kind() == reflect.Ptr {
		// set `to` to nil if from is nil
		if from.Kind() == reflect.Ptr && from.IsNil() {
			to.Set(reflect.Zero(to.Type()))
			return nil
		} else if to.IsNil() {
			to.Set(reflect.New(to.Type().Elem()))
		}
		to = to.Elem()
	}

	if from.Type().ConvertibleTo(to.Type()) {
		to.Set(from.Convert(to.Type()))
	} else if from.Kind() == reflect.Ptr {
		return set(to, from.Elem())
	} else {
		f := from.Type()
		t := to.Type()
		if c, ok := converters[f]; ok {
			if fn, ok := c[t]; ok {
				newVal, err := fn(from.Interface())
				if err != nil {
					return err
				}
				to.Set(reflect.ValueOf(newVal))
			}
		}
	}
	return nil
}
