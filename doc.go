/*
Package trama is a very simple web framework that uses the handler to store the
state of a request.

On trama, when you register a web handler, you actually register a constructor
for it. Each time an HTTP request arrives, a new handler object is constructed.
This way, a handler can be an arbitrarily complex structure holding any
information it needs. This comes in handy when used with another trama
functionality: interceptors.

Interceptors are special units that are called before and after every handler
method call. With interceptors, one can automate most of the repetitive tasks
involving a request handling, like the setup and commit of a database
transaction and automatic decode of query string parameters. Since a handler is
an object that can store any kind of information, an interceptor can be used to
setup this information before the handler process the request and to make any
cleanup after it.

If the handler register the interceptor chain [a, b, c], the framework will
call, in order:

	a.Before
	b.Before
	c.Before
	handler.Method (any of Get or Post)
	c.After
	b.After
	a.After

If any of the interceptors' Before returns an error, the chain is interrupted.
Say, for example, that the Before method of the b interceptor returns an
error; then, the execution will be:

	a.Before
	b.Before
	b.After
	a.After
*/
package trama
