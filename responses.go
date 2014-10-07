package trama

import (
	"fmt"
	"net/http"
	"path"
)

type ResponseWriter struct {
	http.ResponseWriter
	status  int
	Written bool
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	if !w.Written {
		// note: the first call to Write will trigger an
		// implicit WriteHeader(http.StatusOK).
		if w.status > 0 {
			w.ResponseWriter.WriteHeader(w.status)
		}
	}

	w.Written = true
	return w.ResponseWriter.Write(b)
}

func (w *ResponseWriter) WriteHeader(s int) {
	w.status = s
}

func (w *ResponseWriter) Status() int {
	return w.status
}

type Response interface {
	Redirect(url string, statusCode int)
	ExecuteTemplate(name string, data interface{})
	SetCookie(cookie *http.Cookie)
	SetTemplateGroup(name string)
}

type WebResponse struct {
	responseWriter       http.ResponseWriter
	request              *http.Request
	redirectURL          string
	redirectStatusCode   int
	templateName         string
	templateData         interface{}
	templates            TemplateGroupSet
	currentTemplateGroup string
	written              bool
	err                  error
	log                  func(error)
}

func NewWebResponse(w http.ResponseWriter, r *http.Request, templ TemplateGroupSet) *WebResponse {
	return &WebResponse{responseWriter: w, request: r, templates: templ}
}

func (r *WebResponse) SetTemplateGroup(name string) {
	r.currentTemplateGroup = name
}

func (r *WebResponse) Redirect(url string, statusCode int) {
	r.written = true
	r.redirectURL = url
	r.redirectStatusCode = statusCode
}

func (r *WebResponse) ExecuteTemplate(name string, data interface{}) {
	r.written = true
	_, filename := path.Split(name)
	r.templateName = filename
	r.templateData = data
}

func (r *WebResponse) SetCookie(cookie *http.Cookie) {
	http.SetCookie(r.responseWriter, cookie)
}

func (r *WebResponse) Written() bool {
	return r.written
}

func (r *WebResponse) Write() {
	if !r.Written() {
		r.responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	if r.redirectStatusCode != 0 {
		http.Redirect(r.responseWriter, r.request, r.redirectURL, r.redirectStatusCode)
	} else {
		group, found := r.templates[r.currentTemplateGroup]

		if !found {
			r.log(fmt.Errorf("No template group named “%s” was found", r.currentTemplateGroup))
			r.responseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := group.executeTemplate(r.responseWriter, r.templateName, r.templateData)

		if err != nil {
			r.log(err)
			r.responseWriter.WriteHeader(http.StatusInternalServerError)
		}
	}
}
