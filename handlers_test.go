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
		templ, err := template.ParseFiles(templatesNames...)

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
			t.Errorf("Item %d, “%s”, unexpected result. Expecting “%s”;\nfound “%s”", i, item.description, item.expectedResult1, item.content1)
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
			t.Errorf("Item %d, “%s”, unexpected result. Expecting “%s”;\nfound “%s”", i, item.description, item.expectedResult2, item.content2)
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

	res.SetTemplate(m.template2.Name(), m.template2Data)
	return nil
}

func (m *mockWebHandler) Templates() []string {
	f, err := ioutil.TempFile("", "mockWebHandler")

	if err != nil {
		return nil
	}

	m.template1 = f

	m.template2, err = ioutil.TempFile("", "mockWebHandler")

	if err != nil {
		println(err.Error())
		return nil
	}

	_, err = io.WriteString(m.template1, m.template1Content)

	if err != nil {
		println(err.Error())
		return nil
	}

	_, err = io.WriteString(m.template2, m.template2Content)

	if err != nil {
		println(err.Error())
		return nil
	}

	return []string{m.template1.Name(), m.template2.Name()}
}

func (m *mockWebHandler) Interceptors() WebInterceptorChain {
	return m.interceptors
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
