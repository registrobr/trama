package trama

import (
	"errors"
	"html/template"
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
	Error(error)
}

type WebResponse struct {
	responseWriter     http.ResponseWriter
	request            *http.Request
	redirectURL        string
	redirectStatusCode int
	templateName       string
	templateData       interface{}
	typ                responseType
	errorTemplate      string
	template           *template.Template
	err                error
	log                func(error)
}

func NewWebResponse(w http.ResponseWriter, r *http.Request, templ *template.Template, errorTemplate string) *WebResponse {
	return &WebResponse{
		responseWriter: w,
		request:        r,
		template:       templ,
		errorTemplate:  errorTemplate,
	}
}

type responseType string

const (
	TypeTemplate responseType = "template"
	TypeRedirect responseType = "redirect"
	TypeError    responseType = "error"
)

func (r *WebResponse) Redirect(url string, statusCode int) {
	r.setType(TypeRedirect)
	r.redirectURL = url
	r.redirectStatusCode = statusCode
}

func (r *WebResponse) ExecuteTemplate(name string, data interface{}) {
	r.setType(TypeTemplate)
	_, filename := path.Split(name)
	r.templateName = filename
	r.templateData = data
}

func (r *WebResponse) Error(err error) {
	r.setType(TypeError)
	r.err = err
}

func (r *WebResponse) SetCookie(cookie *http.Cookie) {
	http.SetCookie(r.responseWriter, cookie)
}

func (r *WebResponse) setType(t responseType) {
	if r.typ != TypeError {
		r.typ = t
	}
}

func (r *WebResponse) Written() bool {
	return r.typ != ""
}

func (r *WebResponse) Write() {
	switch r.typ {
	case TypeTemplate:
		err := r.template.ExecuteTemplate(r.responseWriter, r.templateName, r.templateData)

		if err != nil {
			r.log(err)
		}

	case TypeError:
		err := r.template.ExecuteTemplate(r.responseWriter, r.errorTemplate, r.err)

		if err != nil {
			r.log(err)
		}

	case TypeRedirect:
		http.Redirect(r.responseWriter, r.request, r.redirectURL, r.redirectStatusCode)

	default:
		r.log(errors.New("Unknown WebResponse type"))
	}
}
