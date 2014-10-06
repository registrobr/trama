package trama

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterService(t *testing.T) {
	data := []struct {
		description   string
		uri           string
		expectedPanic bool
	}{
		{description: "it should register a service correctly", uri: "/example", expectedPanic: false},
		{description: "it should deny duplicated URI", uri: "/example", expectedPanic: true},
	}

	trama := New(func(err error) {
		t.Fatal(err)
	})

	for i, item := range data {
		defer func() {
			if r := recover(); r != nil {
				if !item.expectedPanic {
					t.Errorf("Item %d, “%s”, wrong result. Unexpected panic: %+v", i, item.description, r)
				}
			}
		}()

		trama.RegisterService(item.uri, func() AJAXHandler {
			return &mockAJAXHandler{}
		})

		if item.expectedPanic {
			t.Errorf("Item %d, “%s”, wrong result. Expected panic!", i, item.description)
		}
	}
}

func TestServeHTTP(t *testing.T) {
	data := []struct {
		description    string
		uri            string
		routes         map[string]webHandlerConstructor
		recoverDefined bool
		expectedStatus int
	}{
		{
			description: "it should call a handler correctly",
			uri:         "/example",
			routes: map[string]webHandlerConstructor{
				"/example": func() WebHandler {
					return &mockWebHandler{
						templateGetRedirectURL: "/redirect",
					}
				},
			},
			recoverDefined: true,
			expectedStatus: http.StatusFound,
		},
		{
			description:    "it should detect when the URI doesn't exist",
			uri:            "/idontexist",
			routes:         nil,
			recoverDefined: true,
			expectedStatus: http.StatusNotFound,
		},
		{
			description: "it should panic in the handler and call recover function",
			uri:         "/example",
			routes: map[string]webHandlerConstructor{
				"/example": func() WebHandler {
					return &crazyWebHandler{}
				},
			},
			recoverDefined: true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			description: "it should panic in the handler and log the recover",
			uri:         "/example",
			routes: map[string]webHandlerConstructor{
				"/example": func() WebHandler {
					return &crazyWebHandler{}
				},
			},
			recoverDefined: false,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for i, item := range data {
		trama := New(func(err error) {
			if item.recoverDefined {
				t.Fatal(err)
			}
		})

		for uri, handler := range item.routes {
			trama.RegisterPage(uri, handler)
		}

		if item.recoverDefined {
			trama.Recover = func(r interface{}) {}
		} else {
			trama.Recover = nil
		}

		r, err := http.NewRequest("GET", item.uri, nil)
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()

		trama.ServeHTTP(w, r)

		if w.Code != item.expectedStatus {
			t.Errorf("Item %d, “%s”, unexpected result. Expecting “%d”;\nfound “%d”",
				i, item.description, item.expectedStatus, w.Code)
		}
	}
}

type crazyWebHandler struct {
}

func (h *crazyWebHandler) Get(Response, *http.Request) {
	panic(fmt.Errorf("I'm a crazy handler!"))
}

func (h *crazyWebHandler) Post(Response, *http.Request) {
	panic(fmt.Errorf("I'm a crazy handler!"))
}

func (h *crazyWebHandler) Interceptors() WebInterceptorChain {
	return NewWebInterceptorChain()
}

func (h *crazyWebHandler) Templates() []string {
	return []string{}
}

func (h *crazyWebHandler) TemplatesFunc() template.FuncMap {
	return nil
}
