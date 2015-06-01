package trama

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
)

func TestMuxServeHTTP(t *testing.T) {
	globalTemplates, err := writeGlobalTemplates()

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	defer func() {
		globalTemplates[0].Close()
		globalTemplates[1].Close()
	}()

	_, name1 := path.Split(globalTemplates[0].Name())
	_, name2 := path.Split(globalTemplates[1].Name())

	data := []struct {
		description     string
		uri             string
		templateContent string
		expectedContent string
	}{
		{
			description: "It should render the template including the header and footer",
			uri:         "/olha/aqui",

			templateContent: `
			[[template "` + name1 + `"]]

			Viver seu tempo: para o que ir viver
			num deserto literal ou de alpendres;
			em ermos, que não distraiam de viver
			a agulha de um só instante, plenamente.

			[[template "` + name2 + `"]]
			`,

			expectedContent: `
			Habitar o tempo

			Viver seu tempo: para o que ir viver
			num deserto literal ou de alpendres;
			em ermos, que não distraiam de viver
			a agulha de um só instante, plenamente.

			João Cabral de Melo Neto
			`,
		},
		{
			description: "It should render this template instead of the previous one",
			uri:         "/olha/aqui-também",

			templateContent: `
			[[template "` + name1 + `"]]

			Viver seu tempo: para o que ir viver
			num deserto literal ou de alpendres;
			em ermos, que não distraiam de viver
			a agulha de um só instante, plenamente.

			[[template "` + name2 + `"]]
			`,

			expectedContent: `
			Habitar o tempo

			Viver seu tempo: para o que ir viver
			num deserto literal ou de alpendres;
			em ermos, que não distraiam de viver
			a agulha de um só instante, plenamente.

			João Cabral de Melo Neto
			`,
		},
	}

	mux := NewMux()
	mux.SetLogger(func(err error) { t.Error("Unexpected error:", err) })
	mux.SetTemplateDelims("[[", "]]")
	mux.GlobalTemplates = NewTemplateGroupSet(nil)
	groupName := "pt"
	mux.GlobalTemplates.Insert(TemplateGroup{
		Name:  groupName,
		Files: []string{globalTemplates[0].Name(), globalTemplates[1].Name()},
	})

	for i, item := range data {
		handler := &mockHandler{
			templateGroup:      groupName,
			templateGetContent: item.templateContent,
		}

		defer handler.closeTemplates()

		mux.Register(item.uri, func() Handler { return handler })
		err := mux.ParseTemplates()

		if err != nil {
			t.Errorf("Item %d, “%s”, unexpected error: “%s”", i, item.description, err)
		}

		r, err := http.NewRequest("GET", item.uri, nil)

		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Item %d, “%s”, unexpected status code. Expecting %d; found %d", i, item.description, http.StatusOK, w.Code)
		}

		if w.Body.String() != item.expectedContent {
			t.Errorf("Item %d, “%s”, unexpected result. Expecting %s; found %s", i, item.description, item.expectedContent, w.Body)
		}
	}
}

func writeGlobalTemplates() ([]*os.File, error) {
	var err error
	var global1, global2 *os.File

	global1, err = ioutil.TempFile("", "global1")
	if err != nil {
		return nil, err
	}

	global2, err = ioutil.TempFile("", "global2")
	if err != nil {
		return nil, err
	}

	if _, err = io.WriteString(global1, "Habitar o tempo"); err != nil {
		return nil, err
	}

	if _, err = io.WriteString(global2, "João Cabral de Melo Neto"); err != nil {
		return nil, err
	}

	return []*os.File{global1, global2}, nil
}

func TestServeHTTP(t *testing.T) {
	mock := &mockHandler{templateGetRedirectURL: "/redirect"}
	defer mock.closeTemplates()

	data := []struct {
		description    string
		uriToRegister  string
		handler        Handler
		requestURI     string
		recoverDefined bool
		expectedStatus int
	}{
		{
			description:    "it should call a handler correctly",
			uriToRegister:  "/example",
			requestURI:     "/example",
			handler:        mock,
			recoverDefined: true,
			expectedStatus: http.StatusFound,
		},
		{
			description:    "it should detect when the URI doesn't exist",
			uriToRegister:  "/example",
			requestURI:     "/cadê-eu",
			handler:        mock,
			recoverDefined: true,
			expectedStatus: http.StatusNotFound,
		},
		{
			description:    "it should panic in the handler and call recover function",
			uriToRegister:  "/example",
			requestURI:     "/example",
			handler:        &crazyHandler{},
			recoverDefined: true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			description:    "it should panic in the handler and log the recover",
			uriToRegister:  "/example",
			requestURI:     "/example",
			handler:        &crazyHandler{},
			recoverDefined: false,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for i, item := range data {
		mux := NewMux()
		mux.SetLogger(func(err error) {
			if item.recoverDefined {
				t.Fatal(err)
			}
		})

		mux.Register(item.uriToRegister, func() Handler { return item.handler })

		if item.recoverDefined {
			mux.Recover = func(r interface{}) {}
		} else {
			mux.Recover = nil
		}

		r, err := http.NewRequest("GET", item.requestURI, nil)

		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != item.expectedStatus {
			t.Errorf("Item %d, “%s”, unexpected result. Expecting “%d”;\nfound “%d”",
				i, item.description, item.expectedStatus, w.Code)
		}
	}
}

type crazyHandler struct{}

func (h *crazyHandler) Get(Response, *http.Request) error {
	panic(fmt.Errorf("I'm a crazy handler!"))
}

func (h *crazyHandler) Post(Response, *http.Request) error {
	panic(fmt.Errorf("I'm a crazy handler!"))
}

func (h *crazyHandler) Interceptors() InterceptorChain {
	return NewInterceptorChain()
}

func (h *crazyHandler) Templates() TemplateGroupSet {
	return NewTemplateGroupSet(nil)
}
