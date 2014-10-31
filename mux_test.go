package trama

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
)

func TestRegisterPage(t *testing.T) {
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
		description                   string
		templateContentHigherPriority string
		templateContentLowerPriority  string
		expectedContent               string
	}{
		{
			description: "I should render the template including the header and footer",
			templateContentLowerPriority: `
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
			description: "I should render the template including the header, footer, and the aditional template",

			templateContentHigherPriority: `
			[[template "` + name1 + `"]]

			Viver seu tempo: para o que ir viver
			num deserto literal ou de alpendres;
			em ermos, que não distraiam de viver
			a agulha de um só instante, plenamente.

			[[template "` + name2 + `"]]
			`,

			templateContentLowerPriority: `ele corre vazio, o tal tempo ao vivo;`,

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

	mux := New(func(err error) { t.Error("Unexpected error:", err) })
	mux.SetTemplateDelims("[[", "]]")
	mux.GlobalTemplates = NewTemplateGroupSet(nil)
	groupName := "pt"
	mux.GlobalTemplates.Insert(TemplateGroup{
		Name:  groupName,
		Files: []string{globalTemplates[0].Name(), globalTemplates[1].Name()},
	})

	for i, item := range data {
		handler := &mockWebHandler{
			templateGroup:       groupName,
			templateGetContent:  item.templateContentHigherPriority,
			templatePostContent: item.templateContentLowerPriority,
		}

		defer handler.closeTemplates()

		uri := fmt.Sprintf("/uri/%d", i)
		mux.RegisterPage(uri, func() WebHandler { return handler })
		err := mux.ParseTemplates()

		if err != nil {
			t.Errorf("Item %d, “%s”, unexpected error: “%s”", i, item.description, err)
		}

		h, err := mux.router.match(uri)

		if err != nil {
			t.Errorf("Item %d, “%s”, unexpected error: “%s”", i, item.description, err)

		} else if h.webHandler == nil {
			t.Errorf("Item %d, “%s”, nil web handler constructor", i, item.description)

		} else if h.webHandler() != handler {
			t.Errorf("Item %d, “%s”, mismatch handlers. Expecting %p; found %p", i, item.description, handler, h.webHandler())

		} else {
			buffer := new(bytes.Buffer)
			var err error

			if item.templateContentHigherPriority != "" {
				_, filename := path.Split(handler.templateGet.Name())
				err = h.templates.elements[groupName].executeTemplate(buffer, filename, nil)
			} else {
				_, filename := path.Split(handler.templatePost.Name())
				err = h.templates.elements[groupName].executeTemplate(buffer, filename, nil)
			}

			if err != nil {
				t.Errorf("Item %d, “%s”, unexpected error: “%s”", i, item.description, err)

			} else if buffer.String() != item.expectedContent {
				t.Errorf("Item %d, “%s”, unexpected result. Expecting %s; found %s", i, item.description, item.expectedContent, buffer.String())
			}
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
	mock := &mockWebHandler{templateGetRedirectURL: "/redirect"}
	defer mock.closeTemplates()

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
				"/example": func() WebHandler { return mock },
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

func (h *crazyWebHandler) Templates() TemplateGroupSet {
	return NewTemplateGroupSet(nil)
}
