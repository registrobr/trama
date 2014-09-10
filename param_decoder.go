package trama

import (
	"reflect"
	"strconv"
)

type paramDecoder struct {
	structure interface{}
	uriParams map[string]string
	err       func(error)
}

func newParamDecoder(h interface{}, uriParams map[string]string, errorFunc func(error)) paramDecoder {
	return paramDecoder{structure: h, uriParams: uriParams, err: errorFunc}
}

func (c paramDecoder) decode() {
	st := reflect.ValueOf(c.structure).Elem()

	for i := 0; i < st.NumField(); i++ {
		field := st.Type().Field(i)
		value := field.Tag.Get("param")

		if value == "" {
			continue
		}

		param, ok := c.uriParams[value]

		if !ok {
			continue
		}

		s := st.Field(i)

		if s.IsValid() && s.CanSet() {
			switch field.Type.Kind() {
			case reflect.String:
				s.SetString(param)
			case reflect.Int:
				n, err := strconv.ParseInt(param, 10, 0)
				if err != nil {
					c.err(err)
					continue
				}
				s.SetInt(n)
			case reflect.Int64:
				n, err := strconv.ParseInt(param, 10, 64)
				if err != nil {
					c.err(err)
					continue
				}
				s.SetInt(n)
			}
		}
	}
}
