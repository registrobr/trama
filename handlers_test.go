package trama

import (
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestServeWeb(t *testing.T) {
	data := []struct {
		description     string
		content1        string
		data1           interface{}
		expectedResult1 string
		content2        string
		data2           interface{}
		expectedResult2 string
		interceptors    WebInterceptorChain
		testStatusCode  bool
	}{
		{
			description: "It should write the expected results",
			content1: `
				Um {{.Galo}} sozinho não tece uma manhã:
				ele precisará sempre de outros {{.Galos}}.
				De um que {{myFunc "apanhe"}} esse grito que ele
				e o lance a outro; de um outro {{.Galo}}
				que apanhe o grito de um {{.Galo}} antes
				e o lance a outro; e de outros {{.Galos}}
				que com muitos outros {{.Galos}} se cruzem
				os fios de sol de seus gritos de {{.Galo}},
				para que a manhã, desde uma teia tênue,
				se vá tecendo, entre todos os {{.Galos}}.`,
			data1: struct{ Galo, Galos string }{"galo", "galos"},
			expectedResult1: `
				Um galo sozinho não tece uma manhã:
				ele precisará sempre de outros galos.
				De um que !confidential! esse grito que ele
				e o lance a outro; de um outro galo
				que apanhe o grito de um galo antes
				e o lance a outro; e de outros galos
				que com muitos outros galos se cruzem
				os fios de sol de seus gritos de galo,
				para que a manhã, desde uma teia tênue,
				se vá tecendo, entre todos os galos.`,
			content2: `
				E se encorpando em tela, entre {{.Todos}},
				se erguendo tenda, onde entrem {{.Todos}},
				se entretendo para {{.Todos}}, no toldo
				(a manhã) que plana livre de armação.
				A manhã, toldo de um tecido tão aéreo
				que, tecido, se eleva por si: luz balão.`,
			data2: struct{ Todos string }{"todos"},
			expectedResult2: `
				E se encorpando em tela, entre todos,
				se erguendo tenda, onde entrem todos,
				se entretendo para todos, no toldo
				(a manhã) que plana livre de armação.
				A manhã, toldo de um tecido tão aéreo
				que, tecido, se eleva por si: luz balão.`,
			testStatusCode: true,
		},
		{
			description: "It should write the expected results after running the interceptors",
			content1: `
				Um {{.Galo}} sozinho não tece uma manhã:
				ele precisará sempre de outros {{.Galos}}.
				De um que apanhe esse grito que ele
				e o lance a outro; de um outro {{.Galo}}
				que apanhe o grito de um {{.Galo}} antes
				e o lance a outro; e de outros {{.Galos}}
				que com muitos outros {{.Galos}} se cruzem
				os fios de sol de seus gritos de {{.Galo}},
				para que a manhã, desde uma teia tênue,
				se vá tecendo, entre todos os {{.Galos}}.`,
			data1: struct{ Galo, Galos string }{"galo", "galos"},
			expectedResult1: `
				Um galo sozinho não tece uma manhã:
				ele precisará sempre de outros galos.
				De um que apanhe esse grito que ele
				e o lance a outro; de um outro galo
				que apanhe o grito de um galo antes
				e o lance a outro; e de outros galos
				que com muitos outros galos se cruzem
				os fios de sol de seus gritos de galo,
				para que a manhã, desde uma teia tênue,
				se vá tecendo, entre todos os galos.`,
			content2: `
				E se encorpando em tela, entre {{.Todos}},
				se erguendo tenda, onde entrem {{.Todos}},
				se entretendo para {{.Todos}}, no toldo
				(a manhã) que plana livre de armação.
				A manhã, toldo de um tecido tão aéreo
				que, tecido, se eleva por si: luz balão.`,
			data2: struct{ Todos string }{"todos"},
			expectedResult2: `
				E se encorpando em tela, entre todos,
				se erguendo tenda, onde entrem todos,
				se entretendo para todos, no toldo
				(a manhã) que plana livre de armação.
				A manhã, toldo de um tecido tão aéreo
				que, tecido, se eleva por si: luz balão.`,
			testStatusCode: true,
			interceptors: WebInterceptorChain{
				&struct{ NopWebInterceptor }{},
				&struct{ NopWebInterceptor }{},
				&struct{ NopWebInterceptor }{},
			},
		},
		{
			description:     "It should break at the interceptor's Before run",
			content1:        "Tecendo a manhã",
			expectedResult1: "",
			interceptors: WebInterceptorChain{
				&struct{ NopWebInterceptor }{},
				&brokenBeforeInterceptor{},
				&struct{ NopWebInterceptor }{},
			},
		},
		{
			description:     "It should break at the interceptor's After run",
			content1:        "Tecendo a manhã",
			expectedResult1: "",
			interceptors: WebInterceptorChain{
				&struct{ NopWebInterceptor }{},
				&brokenAfterInterceptor{},
				&struct{ NopWebInterceptor }{},
			},
			testStatusCode: true,
		},
	}

	for i, item := range data {
		mock := &mockWebHandler{
			template1Content: item.content1,
			template1Data:    item.data1,
			template2Content: item.content2,
			template2Data:    item.data2,
			interceptors:     item.interceptors,
		}

		defer mock.closeTemplates()
		templatesNames := mock.Templates()
		templ, err := template.New("mock").Funcs(mock.TemplatesFunc()).ParseFiles(templatesNames...)

		if err != nil {
			t.Fatalf("Item %d, “%s”, unexpected error: “%s”", i, item.description, err)
		}

		handler := adapter{
			webHandler: func() WebHandler { return mock },
			err: func(err error) {
				notBeforeError := err.Error() != brokenBeforeError.Error()
				notAfterError := err.Error() != brokenAfterError.Error()

				if notBeforeError && notAfterError {
					t.Errorf("Item %d, “%s”, unexpected error: “%s”", i, item.description, err)
				}
			},
			template: templ,
		}

		w := httptest.NewRecorder()
		r, err := http.NewRequest("GET", "/uri", nil)

		if err != nil {
			t.Error(err)
		}

		handler.serveHTTP(w, r)

		if item.testStatusCode && w.Code != http.StatusOK {
			t.Errorf("Item %d, “%s”, wrong status code. Expecting 200; found %d", i, item.description, w.Code)
		}

		if w.Body.String() != item.expectedResult1 {
			t.Errorf("Item %d, “%s”, unexpected result. Expecting “%s”;\nfound “%s”", i, item.description, item.expectedResult1, w.Body.String())
		}

		w = httptest.NewRecorder()
		r, err = http.NewRequest("POST", "/uri", nil)

		if err != nil {
			t.Error(err)
		}

		handler.serveHTTP(w, r)

		if item.testStatusCode && w.Code != http.StatusOK {
			t.Errorf("Item %d, “%s”, wrong status code. Expecting 200; found %d", i, item.description, w.Code)
		}

		if w.Body.String() != item.expectedResult2 {
			t.Errorf("Item %d, “%s”, unexpected result. Expecting “%s”;\nfound “%s”", i, item.description, item.expectedResult2, w.Body.String())
		}

		if item.testStatusCode && w.Header().Get("Set-Cookie") != "cookie1=value1" {
			t.Errorf("Item %d, “%s”, unexpected result. Expecting “cookie1=value1”;\nfound “%s”", i, item.description, w.Header().Get("Set-Cookie"))
		}

		w = httptest.NewRecorder()
		r, err = http.NewRequest("DELETE", "/uri", nil)

		if err != nil {
			t.Error(err)
		}

		handler.serveHTTP(w, r)

		if item.testStatusCode && w.Code != http.StatusNotImplemented {
			t.Errorf("Item %d, “%s”, wrong status code. Expecting %d; found %d", i, item.description, http.StatusNotImplemented, w.Code)
		}
	}
}

func TestServeAJAX(t *testing.T) {
	data := []struct {
		description           string
		interceptors          AJAXInterceptorChain
		httpMethod            string
		expectedStatusCode    int
		handlerShouldBeCalled bool
	}{
		{
			description:           "It should handle the GET request properly",
			httpMethod:            "GET",
			expectedStatusCode:    http.StatusOK,
			handlerShouldBeCalled: true,
		},
		{
			description:           "It should handle the PUT request properly",
			httpMethod:            "PUT",
			expectedStatusCode:    http.StatusOK,
			handlerShouldBeCalled: true,
		},
		{
			description:           "It should handle the POST request properly",
			httpMethod:            "POST",
			expectedStatusCode:    http.StatusOK,
			handlerShouldBeCalled: true,
		},
		{
			description:           "It should handle the PATCH request properly",
			httpMethod:            "PATCH",
			expectedStatusCode:    http.StatusOK,
			handlerShouldBeCalled: true,
		},
		{
			description:           "It should handle the DELETE request properly",
			httpMethod:            "DELETE",
			expectedStatusCode:    http.StatusOK,
			handlerShouldBeCalled: true,
		},
		{
			description:           "It should handle the HEAD request properly",
			httpMethod:            "HEAD",
			expectedStatusCode:    http.StatusOK,
			handlerShouldBeCalled: true,
		},
		{
			description:        "It should handle the HEAD request with interceptors properly",
			httpMethod:         "HEAD",
			expectedStatusCode: http.StatusOK,
			interceptors: AJAXInterceptorChain{
				&struct{ NopAJAXInterceptor }{},
				&struct{ NopAJAXInterceptor }{},
				&struct{ NopAJAXInterceptor }{},
			},
			handlerShouldBeCalled: true,
		},
		{
			description:        "It should break at the interceptor's Before run and not run the handler's method",
			httpMethod:         "HEAD",
			expectedStatusCode: http.StatusInternalServerError,
			interceptors: AJAXInterceptorChain{
				&struct{ NopAJAXInterceptor }{},
				&brokenBeforeAJAXInterceptor{},
				&struct{ NopAJAXInterceptor }{},
			},
			handlerShouldBeCalled: false,
		},
	}

	for i, item := range data {
		handleFuncCalled := false
		mock := &mockAJAXHandler{
			handleFunc: func(http.ResponseWriter, *http.Request) {
				handleFuncCalled = true
			},
			interceptors: item.interceptors,
		}
		handler := adapter{
			ajaxHandler: func() AJAXHandler { return mock },
			err: func(err error) {
				t.Errorf("Item %d, “%s”, unexpected error found: %s", i, item.description, err)
			},
			uriVars: map[string]string{"param1": "1", "param2": "2"},
		}

		w := httptest.NewRecorder()
		r, err := http.NewRequest(item.httpMethod, "", nil)

		if err != nil {
			t.Error(err)
		}

		handler.serveHTTP(w, r)

		if item.handlerShouldBeCalled {
			if !handleFuncCalled {
				t.Errorf("Item %d, “%s”, not calling handler", i, item.description)
			} else {
				if mock.methodCalled != item.httpMethod {
					t.Errorf("Item %d, “%s”, wrong method called. Expecting %s; found %s", i, item.description, item.httpMethod, mock.methodCalled)
				}
			}
		}

		if mock.Param1 != "1" {
			t.Errorf("Item %d, “%s”, wrong param1. Expecting “1”; found “%s”", i, item.description, mock.Param1)
		}

		if mock.Param2 != 2 {
			t.Errorf("Item %d, “%s”, wrong param1. Expecting “2”; found “%d”", i, item.description, mock.Param2)
		}

		if w.Code != item.expectedStatusCode {
			t.Errorf("Item %d, “%s”, wrong status code. Expecting %d; found %d", i, item.description, item.expectedStatusCode, w.Code)
		}
	}
}

type mockWebHandler struct {
	template1        *os.File
	template1Content string
	template1Data    interface{}
	template2        *os.File
	template2Content string
	template2Data    interface{}
	interceptors     WebInterceptorChain
}

func (m *mockWebHandler) closeTemplates() {
	m.template1.Close()
	m.template2.Close()
}

func (m *mockWebHandler) Get(res Response, req *http.Request) error {
	if m.template1 == nil {
		return errors.New("Template 1 not set")
	}

	res.SetTemplate(m.template1.Name(), m.template1Data)
	return nil
}

func (m *mockWebHandler) Post(res Response, req *http.Request) error {
	if m.template2 == nil {
		return errors.New("Template 2 not set")
	}

	res.SetCookie(&http.Cookie{
		Name:  "cookie1",
		Value: "value1",
	})
	res.SetTemplate(m.template2.Name(), m.template2Data)
	return nil
}

func (m *mockWebHandler) Templates() []string {
	var err error

	m.template1, err = ioutil.TempFile("", "mockWebHandler")
	if err != nil {
		println(err.Error())
		return nil
	}

	m.template2, err = ioutil.TempFile("", "mockWebHandler")
	if err != nil {
		println(err.Error())
		return nil
	}

	if _, err = io.WriteString(m.template1, m.template1Content); err != nil {
		println(err.Error())
		return nil
	}

	if _, err = io.WriteString(m.template2, m.template2Content); err != nil {
		println(err.Error())
		return nil
	}

	return []string{m.template1.Name(), m.template2.Name()}
}

func (m *mockWebHandler) Interceptors() WebInterceptorChain {
	return m.interceptors
}

func (m *mockWebHandler) TemplatesFunc() template.FuncMap {
	return template.FuncMap{
		"myFunc": func(value string) string {
			return "!confidential!"
		},
	}
}

type brokenBeforeInterceptor struct {
	NopWebInterceptor
}

var (
	brokenBeforeError = errors.New("Error from a broken Before implementation of a web interceptor")
	brokenAfterError  = errors.New("Error from a broken After implementation of a web interceptor")
)

func (b *brokenBeforeInterceptor) Before(Response, *http.Request) error {
	return brokenBeforeError
}

type brokenAfterInterceptor struct {
	NopWebInterceptor
}

func (b *brokenAfterInterceptor) After(Response, *http.Request) error {
	return brokenAfterError
}

type mockAJAXHandler struct {
	Param1       string `param:"param1"`
	Param2       int    `param:"param2"`
	handleFunc   func(http.ResponseWriter, *http.Request)
	interceptors AJAXInterceptorChain
	methodCalled string
}

func (m *mockAJAXHandler) Get(w http.ResponseWriter, r *http.Request) {
	m.methodCalled = "GET"
	m.handleFunc(w, r)
}

func (m *mockAJAXHandler) Post(w http.ResponseWriter, r *http.Request) {
	m.methodCalled = "POST"
	m.handleFunc(w, r)
}

func (m *mockAJAXHandler) Put(w http.ResponseWriter, r *http.Request) {
	m.methodCalled = "PUT"
	m.handleFunc(w, r)
}

func (m *mockAJAXHandler) Delete(w http.ResponseWriter, r *http.Request) {
	m.methodCalled = "DELETE"
	m.handleFunc(w, r)
}

func (m *mockAJAXHandler) Patch(w http.ResponseWriter, r *http.Request) {
	m.methodCalled = "PATCH"
	m.handleFunc(w, r)
}

func (m *mockAJAXHandler) Head(w http.ResponseWriter, r *http.Request) {
	m.methodCalled = "HEAD"
	m.handleFunc(w, r)
}

func (m *mockAJAXHandler) Interceptors() AJAXInterceptorChain {
	return m.interceptors
}

type brokenBeforeAJAXInterceptor struct {
	NopAJAXInterceptor
}

func (b *brokenBeforeAJAXInterceptor) Before(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

type brokenAfterAJAXInterceptor struct {
	NopAJAXInterceptor
}

func (b *brokenAfterAJAXInterceptor) After(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}
