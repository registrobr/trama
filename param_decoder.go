package trama

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type paramDecoder struct {
	structure interface{}
	uriParams map[string]string
	err       func(error)
}

func newParamDecoder(h interface{}, uriParams map[string]string, errorFunc func(error)) paramDecoder {
	return paramDecoder{structure: h, uriParams: uriParams, err: errorFunc}
}

func (c *paramDecoder) Decode(r *http.Request) {
	st := reflect.ValueOf(c.structure).Elem()
	c.unmarshalURIParams(st)

	m := strings.ToLower(r.Method)
	for i := 0; i < st.NumField(); i++ {
		field := st.Type().Field(i)
		value := field.Tag.Get("request")
		if value == "all" || strings.Contains(value, m) {
			c.unmarshalURIParams(st.Field(i))
		}
	}
}

func (c *paramDecoder) unmarshalURIParams(st reflect.Value) {
	if st.Kind() == reflect.Ptr {
		return
	}

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

		s := st.FieldByName(field.Name)
		if s.IsValid() && s.CanSet() {
			switch field.Type.Kind() {
			case reflect.String:
				s.SetString(param)
			case reflect.Int:
				i, err := strconv.ParseInt(param, 10, 0)
				if err != nil {
					c.err(err)
					continue
				}
				s.SetInt(i)
			case reflect.Int64:
				i, err := strconv.ParseInt(param, 10, 64)
				if err != nil {
					c.err(err)
					continue
				}
				s.SetInt(i)
			}
		}
	}
}
