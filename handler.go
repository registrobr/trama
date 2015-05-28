package trama

import "net/http"

type Handler interface {
	Get(Response, *http.Request) error
	Post(Response, *http.Request) error
	Interceptors() InterceptorChain
	Templates() TemplateGroupSet
}

type DefaultHandler struct {
	NopInterceptorChain
}

func (d *DefaultHandler) Get(Response, *http.Request) error { return nil }

func (d *DefaultHandler) Post(Response, *http.Request) error { return nil }

func (d *DefaultHandler) Templates() TemplateGroupSet {
	return NewTemplateGroupSet(nil)
}

type adapter struct {
	handler   func() Handler
	templates TemplateGroupSet
	log       func(error)
}

func (a adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := &webResponse{
		responseWriter: w,
		request:        r,
		templates:      a.templates,
		log:            a.log,
	}

	handler := a.handler()
	interceptors := handler.Interceptors()
	var err error

	for k, interceptor := range interceptors {
		err = interceptor.Before(response, r)

		if err != nil {
			interceptors = interceptors[:k+1]
			goto write
		}
	}

	switch r.Method {
	case "GET":
		err = handler.Get(response, r)
	case "POST":
		err = handler.Post(response, r)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

write:
	for k := len(interceptors) - 1; k >= 0; k-- {
		interceptors[k].After(response, r, err)
	}

	response.write()
}
