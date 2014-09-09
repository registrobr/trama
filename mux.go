package trama

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"
)

type trama struct {
	sync.RWMutex
	router          router
	templateDelims  []string
	Recover         func(interface{})
	Error           func(error)
	GlobalTemplates []string
}

func New(errorFunc func(error)) *trama {
	t := &trama{router: newRouter()}

	if errorFunc != nil {
		t.Error = errorFunc
	} else {
		t.Error = func(e error) { println(e.Error()) }
	}

	return t
}

type webHandlerConstructor func() WebHandler
type ajaxHandlerConstructor func() AJAXHandler

func (t *trama) SetTemplateDelims(open, clos string) {
	t.templateDelims = []string{open, clos}
}

func (t *trama) RegisterPage(uri string, h webHandlerConstructor) {
	t.Lock()
	defer t.Unlock()

	a := adapter{webHandler: h, err: t.Error}
	templ := template.New(uri)

	if len(t.templateDelims) == 2 {
		templ.Delims(t.templateDelims[0], t.templateDelims[1])
	}

	handlerTemplates := h().Templates()

	if len(handlerTemplates) > 0 {
		files := make([]string, 0, len(t.GlobalTemplates)+len(handlerTemplates))
		files = append(files, t.GlobalTemplates...)
		files = append(files, handlerTemplates...)
		templ = template.Must(templ.ParseFiles(files...))
	}

	a.template = templ

	if err := t.router.appendRoute(uri, a); err != nil {
		panic("Cannot append route: " + err.Error())
	}
}

func (t *trama) RegisterService(uri string, h ajaxHandlerConstructor) {
	t.Lock()
	defer t.Unlock()

	a := adapter{ajaxHandler: h, err: t.Error}

	if err := t.router.appendRoute(uri, a); err != nil {
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
				t.Error(fmt.Errorf("%s", r))
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
