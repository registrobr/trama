package trama

import (
	"net/http"
	"sync"
)

type trama struct {
	sync.RWMutex
	router  *Router
	Recover func(interface{})
}

func New() *trama {
	return &trama{router: NewRouter()}
}

func (t *trama) RegisterPage(uri string, h webHandlerConstructor) {
	t.Lock()
	defer t.Unlock()

	a := &adapter{webHandler: h}

	if err := t.router.AppendRoute(uri, a); err != nil {
		panic("Cannot append route: " + err.Error())
	}
}

func (t *trama) RegisterService(uri string, h ajaxHandlerConstructor) {
	t.Lock()
	defer t.Unlock()

	a := &adapter{ajaxHandler: h}

	if err := t.router.AppendRoute(uri, a); err != nil {
		panic("Cannot append route: " + err.Error())
	}
}

func (t *trama) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.RLock()
	defer t.RUnlock()

	defer func() {
		if r := recover(); r != nil {
			if t.Recover != nil {
				t.Recover(r)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				// TODO Logar erro
			}
		}
	}()

	handler, err := t.router.Match(r.URL.Path)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	handler.ServeHTTP(w, r)
}

type webHandlerConstructor func() WebHandler
type ajaxHandlerConstructor func() AJAXHandler
