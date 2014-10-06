package trama

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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
	mux.GlobalTemplates = []string{globalTemplates[0].Name(), globalTemplates[1].Name()}

	defer func() {
		if x := recover(); x != nil {
			t.Error("Recovering from error:", x)
		}
	}()

	for i, item := range data {
		handler := &mockWebHandler{
			templateGetContent:  item.templateContentHigherPriority,
			templatePostContent: item.templateContentLowerPriority,
		}

		defer handler.closeTemplates()

		uri := fmt.Sprintf("/uri/%d", i)
		mux.RegisterPage(uri, func() WebHandler { return handler })

		h, err := mux.router.match(uri)

		if err != nil {
			t.Errorf("Item %d, “%s”, unexpected error: “%s”", i, item.description, err)

		} else if h.webHandler == nil {
			t.Errorf("Item %d, “%s”, nil web handler constructor", i, item.description)

		} else if h.webHandler() != handler {
			t.Errorf("Item %d, “%s”, mismatch handlers. Expecting %s; found %s", i, item.description, handler, h.webHandler())

		} else {
			buffer := new(bytes.Buffer)
			var err error

			if item.templateContentHigherPriority != "" {
				_, filename := path.Split(handler.templateGet.Name())
				err = h.template.ExecuteTemplate(buffer, filename, nil)
			} else {
				_, filename := path.Split(handler.templatePost.Name())
				err = h.template.ExecuteTemplate(buffer, filename, nil)
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
