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

func TestServe(t *testing.T) {
	data := []struct {
		description string

		contentGet        string
		dataGet           interface{}
		expectedResultGet string
		redirectURL       string

		contentPost        string
		dataPost           interface{}
		expectedResultPost string
		expectedCookies    string

		interceptors       InterceptorChain
		expectedStatusCode int
	}{
		{
			description: "It should write the expected results",
			contentGet: `
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
			dataGet: struct{ Galo, Galos string }{"galo", "galos"},
			expectedResultGet: `
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
			contentPost: `
				E se encorpando em tela, entre {{.Todos}},
				se erguendo tenda, onde entrem {{.Todos}},
				se entretendo para {{.Todos}}, no toldo
				(a manhã) que plana livre de armação.
				A manhã, toldo de um tecido tão aéreo
				que, tecido, se eleva por si: luz balão.`,
			dataPost: struct{ Todos string }{"todos"},
			expectedResultPost: `
				E se encorpando em tela, entre todos,
				se erguendo tenda, onde entrem todos,
				se entretendo para todos, no toldo
				(a manhã) que plana livre de armação.
				A manhã, toldo de um tecido tão aéreo
				que, tecido, se eleva por si: luz balão.`,
			expectedCookies:    "cookie1=value1",
			expectedStatusCode: http.StatusOK,
		},
		{
			description: "It should write the expected results after running the interceptors",
			contentGet: `
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
			dataGet: struct{ Galo, Galos string }{"galo", "galos"},
			expectedResultGet: `
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
			contentPost: `
				E se encorpando em tela, entre {{.Todos}},
				se erguendo tenda, onde entrem {{.Todos}},
				se entretendo para {{.Todos}}, no toldo
				(a manhã) que plana livre de armação.
				A manhã, toldo de um tecido tão aéreo
				que, tecido, se eleva por si: luz balão.`,
			dataPost: struct{ Todos string }{"todos"},
			expectedResultPost: `
				E se encorpando em tela, entre todos,
				se erguendo tenda, onde entrem todos,
				se entretendo para todos, no toldo
				(a manhã) que plana livre de armação.
				A manhã, toldo de um tecido tão aéreo
				que, tecido, se eleva por si: luz balão.`,
			expectedCookies:    "cookie1=value1",
			expectedStatusCode: http.StatusOK,
			interceptors: InterceptorChain{
				&struct{ NopInterceptor }{},
				&struct{ NopInterceptor }{},
				&struct{ NopInterceptor }{},
			},
		},
		{
			description:       "It should break at the interceptor's Before run",
			contentGet:        "Tecendo a manhã",
			expectedResultGet: "",
			interceptors: InterceptorChain{
				&struct{ NopInterceptor }{},
				&brokenBeforeInterceptor{},
				&struct{ NopInterceptor }{},
			},
		},
		{
			description:       "It should redirect when necessary",
			redirectURL:       "/test",
			expectedResultGet: "<a href=\"/test\">Found</a>.\n\n",
		},
	}

	for i, item := range data {
		mock := &mockHandler{
			templateGetContent:     item.contentGet,
			templateGetData:        item.dataGet,
			templateGetRedirectURL: item.redirectURL,
			templatePostContent:    item.contentPost,
			templatePostData:       item.dataPost,
			interceptors:           item.interceptors,
		}

		defer mock.closeTemplates()
		templates := mock.Templates()
		err := templates.parse("", "")

		if err != nil {
			t.Fatalf("Item %d, “%s”, unexpected error: “%s”", i, item.description, err)
		}

		handler := adapter{
			handler: func() Handler { return mock },
			log: func(err error) {
				notBeforeError := err.Error() != errorBrokenBefore.Error()
				notAfterError := err.Error() != errorBrokenAfter.Error()

				if notBeforeError && notAfterError {
					t.Errorf("Item %d, “%s”, unexpected error: “%s”", i, item.description, err)
				}
			},
			templates: templates,
		}

		w := httptest.NewRecorder()
		r, err := http.NewRequest("GET", "/uri", nil)

		if err != nil {
			t.Error(err)
		}

		handler.ServeHTTP(w, r)

		if item.expectedStatusCode != 0 && w.Code != item.expectedStatusCode {
			t.Errorf("Item %d, “%s”, wrong status code. Expecting %d; found %d", i, item.description, item.expectedStatusCode, w.Code)

		} else if item.redirectURL != "" && w.Code != http.StatusFound {
			t.Errorf("Item %d, “%s”, wrong status code. Expecting 302; found %d", i, item.description, w.Code)
		}

		if w.Body.String() != item.expectedResultGet {
			t.Errorf("Item %d, “%s”, unexpected result. Expecting “%s”;\nfound “%s”", i, item.description, item.expectedResultGet, w.Body.String())
		}

		if item.contentPost != "" || item.dataPost != nil {
			w = httptest.NewRecorder()
			r, err = http.NewRequest("POST", "/uri", nil)

			if err != nil {
				t.Error(err)
			}

			handler.ServeHTTP(w, r)

			if item.expectedStatusCode != 0 && w.Code != item.expectedStatusCode {
				t.Errorf("Item %d, “%s”, wrong status code. Expecting %d; found %d", i, item.description, item.expectedStatusCode, w.Code)
			}

			if w.Body.String() != item.expectedResultPost {
				t.Errorf("Item %d, “%s”, unexpected result. Expecting “%s”;\nfound “%s”", i, item.description, item.expectedResultPost, w.Body.String())
			}

			if w.Header().Get("Set-Cookie") != item.expectedCookies {
				t.Errorf("Item %d, “%s”, unexpected result. Expecting “%s”;\nfound “%s”", i, item.description, item.expectedCookies, w.Header().Get("Set-Cookie"))
			}
		}

		w = httptest.NewRecorder()
		r, err = http.NewRequest("DELETE", "/uri", nil)

		if err != nil {
			t.Error(err)
		}

		handler.ServeHTTP(w, r)

		if item.expectedStatusCode != 0 && w.Code != http.StatusNotImplemented {
			t.Errorf("Item %d, “%s”, wrong status code. Expecting %d; found %d", i, item.description, http.StatusNotImplemented, w.Code)
		}
	}
}

type mockHandler struct {
	templateGroup string

	templateGet            *os.File
	templateGetContent     string
	templateGetData        interface{}
	templateGetRedirectURL string

	templatePost        *os.File
	templatePostContent string
	templatePostData    interface{}

	interceptors InterceptorChain
}

func (m *mockHandler) closeTemplates() {
	m.templateGet.Close()
	m.templatePost.Close()
}

func (m *mockHandler) Get(res Response, req *http.Request) error {
	if m.templateGetRedirectURL != "" {
		res.Redirect(m.templateGetRedirectURL, http.StatusFound)

	} else {
		res.ExecuteTemplate(m.templateGet.Name(), m.templateGetData)
	}

	return nil
}

func (m *mockHandler) Post(res Response, req *http.Request) error {
	res.SetCookie(&http.Cookie{Name: "cookie1", Value: "value1"})
	res.ExecuteTemplate(m.templatePost.Name(), m.templatePostData)
	return nil
}

func (m *mockHandler) Templates() TemplateGroupSet {
	var err error

	m.templateGet, err = ioutil.TempFile("", "mockHandler")
	if err != nil {
		return NewTemplateGroupSet(nil)
	}

	m.templatePost, err = ioutil.TempFile("", "mockHandler")
	if err != nil {
		return NewTemplateGroupSet(nil)
	}

	if _, err = io.WriteString(m.templateGet, m.templateGetContent); err != nil {
		return NewTemplateGroupSet(nil)
	}

	if _, err = io.WriteString(m.templatePost, m.templatePostContent); err != nil {
		return NewTemplateGroupSet(nil)
	}

	set := NewTemplateGroupSet(template.FuncMap{
		"myFunc": func(value string) string { return "!confidential!" },
	})
	set.Insert(TemplateGroup{
		Name:  m.templateGroup,
		Files: []string{m.templateGet.Name(), m.templatePost.Name()},
	})

	return set
}

func (m *mockHandler) Interceptors() InterceptorChain {
	chain := NewInterceptorChain(&setGroupInterceptor{groupName: m.templateGroup})
	return append(chain, m.interceptors...)
}

type setGroupInterceptor struct {
	NopInterceptor
	groupName string
}

func (s *setGroupInterceptor) Before(r Response, _ *http.Request) error {
	r.SetTemplateGroup(s.groupName)
	return nil
}

type brokenBeforeInterceptor struct {
	NopInterceptor
}

var (
	errorBrokenBefore = errors.New("Error from a broken Before implementation of a web interceptor")
	errorBrokenAfter  = errors.New("Error from a broken After implementation of a web interceptor")
)

func (b *brokenBeforeInterceptor) Before(r Response, _ *http.Request) error {
	return errorBrokenBefore
}
