package trama

import "net/http"

type WebInterceptor interface {
	Before(Response, *http.Request) error
	After(Response, *http.Request) error
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

type NopAJAXInterceptorChain struct{}

func (n *NopAJAXInterceptorChain) Interceptors() AJAXInterceptorChain {
	return NewAJAXInterceptorChain()
}