package trama

import "net/http"

type WebHandler interface {
	Get(Response, *http.Request)
	Post(Response, *http.Request)
	Interceptors() WebInterceptorChain
	Templates() TemplateGroupSet
}

type AJAXHandler interface {
	Get(http.ResponseWriter, *http.Request)
	Post(http.ResponseWriter, *http.Request)
	Put(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
	Patch(http.ResponseWriter, *http.Request)
	Head(http.ResponseWriter, *http.Request)
	Interceptors() AJAXInterceptorChain
}

type DefaultWebHandler struct {
	NopWebInterceptorChain
}

func (d *DefaultWebHandler) Get(Response, *http.Request) {}

func (d *DefaultWebHandler) Post(Response, *http.Request) {}

func (d *DefaultWebHandler) Templates() TemplateGroupSet {
	return NewTemplateGroupSet(nil)
}

type DefaultAJAXHandler struct {
	NopAJAXInterceptorChain
}

func (s *DefaultAJAXHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *DefaultAJAXHandler) Post(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *DefaultAJAXHandler) Put(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *DefaultAJAXHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *DefaultAJAXHandler) Patch(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *DefaultAJAXHandler) Head(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type adapter struct {
	webHandler    webHandlerConstructor
	ajaxHandler   ajaxHandlerConstructor
	staticHandler http.Handler
	uriVars       map[string]string
	templates     TemplateGroupSet
	log           func(error)
}

func (a adapter) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if a.webHandler != nil {
		a.serveWeb(w, r)
	} else if a.ajaxHandler != nil {
		a.serveAJAX(w, r)
	} else if a.staticHandler != nil {
		a.staticHandler.ServeHTTP(w, r)
	}
}

func (a adapter) serveWeb(w http.ResponseWriter, r *http.Request) {
	response := &webResponse{
		responseWriter: w,
		request:        r,
		templates:      a.templates,
		log:            a.log,
	}

	handler := a.webHandler()
	interceptors := handler.Interceptors()

	for k, interceptor := range interceptors {
		interceptor.Before(response, r)

		if response.written {
			interceptors = interceptors[:k+1]
			goto write
		}
	}

	switch r.Method {
	case "GET":
		handler.Get(response, r)
	case "POST":
		handler.Post(response, r)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}

write:
	for k := len(interceptors) - 1; k >= 0; k-- {
		interceptors[k].After(response, r)
	}

	response.write()
}

func (a adapter) serveAJAX(rw http.ResponseWriter, r *http.Request) {
	w := NewBufferedResponseWriter(rw)
	handler := a.ajaxHandler()
	newParamDecoder(handler, a.uriVars, a.log).decode()
	interceptors := handler.Interceptors()

	for k, interceptor := range interceptors {
		interceptor.Before(w, r)

		if w.status > 0 || w.Body.Len() > 0 {
			interceptors = interceptors[:k+1]
			goto write
		}
	}

	switch r.Method {
	case "GET":
		handler.Get(w, r)
	case "POST":
		handler.Post(w, r)
	case "PUT":
		handler.Put(w, r)
	case "DELETE":
		handler.Delete(w, r)
	case "PATCH":
		handler.Patch(w, r)
	case "HEAD":
		handler.Head(w, r)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}

write:
	for k := len(interceptors) - 1; k >= 0; k-- {
		interceptors[k].After(w, r)
	}

	w.Flush()
}
