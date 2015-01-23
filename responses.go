package trama

import (
	"fmt"
	"net/http"
	"path"
)

type Response interface {
	SetTemplateGroup(name string)
	SetCookie(cookie *http.Cookie)
	Redirect(url string, statusCode int)
	ExecuteTemplate(name string, data interface{})
	TemplateName() string
}

type webResponse struct {
	redirectURL          string
	redirectStatusCode   int
	templateName         string
	templateData         interface{}
	currentTemplateGroup string
	templates            TemplateGroupSet
	written              bool
	responseWriter       http.ResponseWriter
	request              *http.Request
	log                  func(error)
}

func (r *webResponse) TemplateName() string {
	return r.templateName
}

func (r *webResponse) SetTemplateGroup(name string) {
	r.currentTemplateGroup = name
}

func (r *webResponse) Redirect(url string, statusCode int) {
	r.written = true
	r.redirectURL = url
	r.redirectStatusCode = statusCode
}

func (r *webResponse) ExecuteTemplate(name string, data interface{}) {
	r.written = true
	_, filename := path.Split(name)
	r.templateName = filename
	r.templateData = data
}

func (r *webResponse) SetCookie(cookie *http.Cookie) {
	http.SetCookie(r.responseWriter, cookie)
}

func (r *webResponse) write() {
	if !r.written {
		r.responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	if r.redirectStatusCode != 0 {
		http.Redirect(r.responseWriter, r.request, r.redirectURL, r.redirectStatusCode)
	} else {
		group, found := r.templates.elements[r.currentTemplateGroup]

		if !found {
			r.log(fmt.Errorf("No template group named “%s” was found", r.currentTemplateGroup))
			r.responseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := group.executeTemplate(r.responseWriter, r.templateName, r.templateData)

		if err != nil {
			r.log(err)
		}
	}
}
