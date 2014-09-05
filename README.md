Trama
=====

Trama is a simple web server for Golang forked from Handy 
(github.com/trajber/handy). Like Handy, it has the concept of interceptors that 
can be freely plugged in the pipeline of a request handling, allowing making 
libraries of small reusable pieces of funcionality. Also like Handy, it uses 
the handler to carry state across the entire request, what enables injecting 
arbitrary type-safe contextual information to the pipeline.

