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
	Log             func(error)
	GlobalTemplates []string
}

func New(log func(error)) *trama {
	t := &trama{router: newRouter()}

	if log != nil {
		t.Log = log
	} else {
		t.Log = func(err error) { println(err.Error()) }
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

	a := &adapter{webHandler: h, log: t.Log}
	templ := template.New(uri)

	if len(t.templateDelims) == 2 {
		templ.Delims(t.templateDelims[0], t.templateDelims[1])
	}

	handler := h()

	if funcMap := handler.TemplatesFunc(); funcMap != nil {
		templ = templ.Funcs(funcMap)
	}

	handlerTemplates := handler.Templates()

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

	a := &adapter{ajaxHandler: h, log: t.Log}

	if err := t.router.appendRoute(uri, a); err != nil {
		panic("Cannot append route: " + err.Error())
	}
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
				t.Log(fmt.Errorf("%s", r))
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
