package trama

import "net/http"

type WebInterceptor interface {
	Before(Response, *http.Request)
	After(Response, *http.Request)
}

type WebInterceptorChain []WebInterceptor

func (c WebInterceptorChain) Chain(i WebInterceptor) WebInterceptorChain {
	return append(c, i)
}

func NewWebInterceptorChain(is ...WebInterceptor) WebInterceptorChain {
	return is
}

type AJAXInterceptor interface {
	Before(w http.ResponseWriter, r *http.Request)
	After(w http.ResponseWriter, r *http.Request)
}

type AJAXInterceptorChain []AJAXInterceptor

func (c AJAXInterceptorChain) Chain(i AJAXInterceptor) AJAXInterceptorChain {
	return append(c, i)
}

func NewAJAXInterceptorChain(is ...AJAXInterceptor) AJAXInterceptorChain {
	return is
}

type NopWebInterceptorChain struct{}

func (n *NopWebInterceptorChain) Interceptors() WebInterceptorChain {
	return NewWebInterceptorChain()
}

type NopWebInterceptor struct{}

func (n *NopWebInterceptor) Before(Response, *http.Request) {}

func (n *NopWebInterceptor) After(Response, *http.Request) {}

type NopAJAXInterceptorChain struct{}

func (n *NopAJAXInterceptorChain) Interceptors() AJAXInterceptorChain {
	return NewAJAXInterceptorChain()
}

type NopAJAXInterceptor struct{}

func (n *NopAJAXInterceptor) Before(w http.ResponseWriter, r *http.Request) {}
func (n *NopAJAXInterceptor) After(w http.ResponseWriter, r *http.Request)  {}
