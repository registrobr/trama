package trama

import (
	"net/http"
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
	SetTemplate(name string)
	SetTemplateData(data interface{})
}

type TramaResponse struct {
	RedirectURL        string
	RedirectStatusCode int
	TemplateName       string
	TemplateData       interface{}
}

func (t *TramaResponse) Redirect(url string, statusCode int) {
	t.RedirectURL = url
	t.RedirectStatusCode = statusCode
}

func (t *TramaResponse) SetTemplate(name string) {
	t.TemplateName = name
}

func (t *TramaResponse) SetTemplateData(data interface{}) {
	t.TemplateData = data
}
