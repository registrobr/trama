package trama

import (
	"fmt"
	"net/http"
	"sync"
)

type Mux struct {
	Recover         func(interface{})
	GlobalTemplates TemplateGroupSet

	sync.RWMutex
	mux        *http.ServeMux
	log        func(error)
	leftDelim  string
	rightDelim string
	handlers   []*adapter
}

func NewMux(log func(error)) *Mux {
	t := &Mux{mux: http.NewServeMux()}

	if log != nil {
		t.log = log
	} else {
		t.log = func(err error) { println(err.Error()) }
	}

	return t
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
	t.Lock()
	defer t.Unlock()

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
