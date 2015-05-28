package trama

import "net/http"

type Interceptor interface {
	Before(Response, *http.Request) error
	After(Response, *http.Request, error)
}

type InterceptorChain []Interceptor

func (c InterceptorChain) Chain(i Interceptor) InterceptorChain {
	return append(c, i)
}

func NewInterceptorChain(is ...Interceptor) InterceptorChain {
	return is
}

type NopInterceptorChain struct{}

func (n *NopInterceptorChain) Interceptors() InterceptorChain {
	return NewInterceptorChain()
}

type NopInterceptor struct{}

func (n *NopInterceptor) Before(Response, *http.Request) error { return nil }

func (n *NopInterceptor) After(Response, *http.Request, error) {}
