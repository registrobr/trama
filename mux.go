package trama

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"
)

type trama struct {
	sync.RWMutex
	router         *Router
	templateDelims []string
	Recover        func(interface{})
	Error          func(error)
	globalTemplate *template.Template
}

func New(errorFunc func(error)) *trama {
	t := &trama{router: NewRouter()}

	if errorFunc != nil {
		t.Error = errorFunc
	} else {
		t.Error = func(e error) { println(e.Error()) }
	}

	return t
}

type webHandlerConstructor func() WebHandler
type ajaxHandlerConstructor func() AJAXHandler

func (t *trama) SetGlobalTemplates(files ...string) {
	t.globalTemplate = template.Must(template.ParseFiles(files...))

	if len(t.templateDelims) == 2 {
		t.globalTemplate.Delims(t.templateDelims[0], t.templateDelims[1])
	}
}

func (t *trama) SetTemplateDelims(open, clos string) {
	if t.globalTemplate != nil {
		t.globalTemplate.Delims(open, clos)
		return
	}

	t.templateDelims = []string{open, clos}
}

func (t *trama) RegisterPage(uri string, h webHandlerConstructor) {
	t.Lock()
	defer t.Unlock()

	a := &adapter{webHandler: h, err: t.Error}
	templ := template.Must(t.globalTemplate.Clone())

	for _, f := range h().Templates() {
		content, err := ioutil.ReadFile(f)

		if err != nil {
			panic(err)
		}

		s := string(content)
		template.Must(templ.Parse(s))
	}

	a.template = templ

	if err := t.router.AppendRoute(uri, a); err != nil {
		panic("Cannot append route: " + err.Error())
	}
}

func (t *trama) RegisterService(uri string, h ajaxHandlerConstructor) {
	t.Lock()
	defer t.Unlock()

	a := &adapter{ajaxHandler: h, err: t.Error}

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
				t.Error(fmt.Errorf("%s", r))
			}
		}
	}()

	handler, err := t.router.Match(r.URL.Path)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	handler.serveHTTP(w, r)
}
