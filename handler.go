package trama

import "net/http"

// Handler is the interface a trama handler must implement.
type Handler interface {
	// Get handles HTTP requests with a GET method
	Get(Response, *http.Request) error

	// Post handles HTTP requests with a POST method
	Post(Response, *http.Request) error

	// Interceptors defines the interceptor chain to be called along with the
	// handler when a request arrives.
	Interceptors() InterceptorChain

	// Templates returns a TemplateGroupSet to be registered by the framework.
	// These templates are parsed at once when calling Mux’s ParseTemplates
	// method.
	Templates() TemplateGroupSet
}

// NopHandler is a facility for writing handlers. It is meant to be embedded in
// your handler if you don’t need to implement all Handler methods.
type NopHandler struct {
	NopInterceptorChain
}

// Get writes no response and returns no error
func (n *NopHandler) Get(Response, *http.Request) error { return nil }

// Post writes no response and returns no error
func (n *NopHandler) Post(Response, *http.Request) error { return nil }

// Templates returns an empty TemplateGroupSet
func (n *NopHandler) Templates() TemplateGroupSet {
	return NewTemplateGroupSet(nil)
}

type adapter struct {
	handler   func() Handler
	templates TemplateGroupSet
	log       func(error)
}

func (a adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := &response{
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
