package trama

import (
	"fmt"
	"net/http"
	"path"
)

// Response is the interface to write HTTP responses.
type Response interface {
	// SetTemplateGroup specifies which group of templates, among those from
	// the registered group set (see Handler’s method Template) will be used
	// when ExecuteTemplate is called. Useful for set system language for
	// example.
	SetTemplateGroup(name string)

	// SetCookie sets the cookies that will be sent with the response.
	SetCookie(cookie *http.Cookie)

	// Redirect redirects the request to the specified URL.
	Redirect(url string, statusCode int)

	// ExecuteTemplate looks for the named template among those registered in
	// the template group specified with SetTemplateGroup, and prepares it to
	// be parsed using the input data and to be written to the response. The
	// actual writing will only happen after all the interceptors be executed.
	ExecuteTemplate(name string, data interface{})

	// TemplateName returns the name of the template set by a previous call to
	// ExecuteTemplate. It is meant to be used by an interceptor that would
	// wanto to instrospect in its After method the response set by the
	// handler.
	TemplateName() string
}

type response struct {
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
	returnStatus         int
}

func (r *response) TemplateName() string {
	return r.templateName
}

func (r *response) SetTemplateGroup(name string) {
	r.currentTemplateGroup = name
}

func (r *response) Redirect(url string, statusCode int) {
	r.written = true
	r.redirectURL = url
	r.redirectStatusCode = statusCode
}

func (r *response) ExecuteTemplate(name string, data interface{}) {
	r.written = true
	_, filename := path.Split(name)
	r.templateName = filename
	r.templateData = data
}

func (r *response) SetCookie(cookie *http.Cookie) {
	http.SetCookie(r.responseWriter, cookie)
}

func (r *response) write() {
	if !r.written {
		if r.returnStatus != 0 {
			r.responseWriter.WriteHeader(r.returnStatus)
		} else {
			r.responseWriter.WriteHeader(http.StatusInternalServerError)
		}
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
