package trama

import (
	"fmt"
	"net/http"
	"sync"
)

// A Mux is an HTTP multiplexer for trama handlers. It can store global HTML
// templates to be used by any handler.
type Mux struct {
	// Recover specifies an optional function to be called if the goroutine
	// handling the request panics.
	Recover func(interface{})

	// GlobalTemplates stores every HTML template not specific to some handler,
	// such as headers and footers one would use in every page.
	GlobalTemplates TemplateGroupSet

	mutex      sync.RWMutex
	mux        *http.ServeMux
	log        func(error)
	leftDelim  string
	rightDelim string
	handlers   []*adapter
}

// NewMux constructs a new trama multiplexer.
func NewMux() *Mux {
	return &Mux{
		mux: http.NewServeMux(),
		log: func(err error) { println(err.Error()) },
	}
}

// SetLogger sets a function to be called when an internal error happens. If no
// function is set, trama defaults to write to stdout.
func (t *Mux) SetLogger(logger func(error)) {
	t.log = logger
}

func (t *Mux) Register(uri string, h func() Handler) {
	a := &adapter{handler: h, log: t.log}
	t.handlers = append(t.handlers, a)
	t.mux.Handle(uri, a)
}

func (t *Mux) SetTemplateDelims(left, right string) {
	t.leftDelim = left
	t.rightDelim = right
}

func (t *Mux) ParseTemplates() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, h := range t.handlers {
		set := h.handler().Templates()
		err := set.union(t.GlobalTemplates)

		if err != nil {
			return err
		}

		err = set.parse(t.leftDelim, t.rightDelim)

		if err != nil {
			return err
		}

		h.templates = set
	}

	return nil
}

func (t *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			w.WriteHeader(http.StatusInternalServerError)

			if t.Recover != nil {
				t.Recover(r)
			} else {
				t.log(fmt.Errorf("%s", r))
			}
		}
	}()

	t.mux.ServeHTTP(w, r)
}
