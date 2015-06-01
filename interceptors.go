package trama

import "net/http"

// An Interceptor is a special unity that runs before and after any handler
// method call. It can be used by things like setting up and tearing down
// resources, and to modify the request before it reaches the handler, or the
// response after reaching it.
type Interceptor interface {
	// Method to be called before the handler is called. If it returns an
	// error, any subsequent interceptor won’t be called.
	Before(Response, *http.Request) error

	// Method to be called after the handler is called.
	After(Response, *http.Request, error)
}

// InterceptorChain is a sequence of interceptors. Each interceptor has its
// Before method called, in order, before the handler is executed, and has its
// After method called, in reverse order, after it. Any Before method returning
// an error interrupts the execution of the chain.
type InterceptorChain []Interceptor

// NewInterceptorChain creates an interceptor chain with the sequence provided
// as its input.
func NewInterceptorChain(is ...Interceptor) InterceptorChain {
	return is
}

// Chain creates a new chain with the new interceptor plugged at the end of the
// old one.
func (c InterceptorChain) Chain(i Interceptor) InterceptorChain {
	return append(c, i)
}

// NopInterceptorChain is a facility for writting handlers needing no
// interceptors. It is meant to be embedded in the handler.
type NopInterceptorChain struct{}

// Interceptors returns an empty InterceptorChain
func (n *NopInterceptorChain) Interceptors() InterceptorChain {
	return NewInterceptorChain()
}

// NopInterceptor is a facility for writting interceptors that don’t need
// either a Before or an After method. It is meant to be embedded in the
// interceptor.
type NopInterceptor struct{}

// Before writes nothing on Response and returns no error.
func (n *NopInterceptor) Before(Response, *http.Request) error { return nil }

// Before writes nothing on Response.
func (n *NopInterceptor) After(Response, *http.Request, error) {}
