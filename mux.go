package trama

import (
	"fmt"
	"net/http"
	"sync"
)

type trama struct {
	Recover         func(interface{})
	GlobalTemplates TemplateGroupSet

	sync.RWMutex
	router      router
	log         func(error)
	leftDelim   string
	rightDelim  string
	webHandlers []*adapter
}

func New(log func(error)) *trama {
	t := &trama{router: newRouter()}

	if log != nil {
		t.log = log
	} else {
		t.log = func(err error) { println(err.Error()) }
	}

	return t
}

type webHandlerConstructor func() WebHandler
type ajaxHandlerConstructor func() AJAXHandler

func (t *trama) SetTemplateDelims(left, right string) {
	t.leftDelim = left
	t.rightDelim = right
}

func (t *trama) RegisterPage(uri string, h webHandlerConstructor) {
	t.Lock()
	defer t.Unlock()

	a := &adapter{webHandler: h, log: t.log}

	if err := t.router.appendRoute(uri, a); err != nil {
		panic("Cannot append route: " + err.Error())
	}

	t.webHandlers = append(t.webHandlers, a)
}

func (t *trama) RegisterService(uri string, h ajaxHandlerConstructor) {
	t.Lock()
	defer t.Unlock()

	a := &adapter{ajaxHandler: h, log: t.log}

	if err := t.router.appendRoute(uri, a); err != nil {
		panic("Cannot append route: " + err.Error())
	}
}

func (t *trama) RegisterStatic(uri string, root http.FileSystem) {
	t.Lock()
	defer t.Unlock()

	a := &adapter{staticHandler: http.FileServer(root), log: t.log}

	if err := t.router.appendRoute(uri, a); err != nil {
		panic("Cannot append route: " + err.Error())
	}
}

func (t *trama) ParseTemplates() error {
	t.Lock()
	defer t.Unlock()

	for _, h := range t.webHandlers {
		set := h.webHandler().Templates()
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

func (t *trama) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.RLock()
	defer t.RUnlock()

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

	handler, err := t.router.match(r.URL.Path)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	handler.serveHTTP(w, r)
}
