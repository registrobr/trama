package trama

import (
	"net/http"
	"testing"
)

func TestAppendRoute(t *testing.T) {
	var data = []struct {
		description string
		uri         string
		handler     adapter
		err         error
	}{
		{
			description: "It should append the web handler route",
			uri:         "/test/web",
			handler: adapter{
				webHandler: func() WebHandler { return &DefaultWebHandler{} },
			},
		},
		{
			description: "It should append the AJAX handler route",
			uri:         "/test/ajax",
			handler: adapter{
				ajaxHandler: func() AJAXHandler { return &DefaultAJAXHandler{} },
			},
		},
		{
			description: "It should append the handler at a parent route",
			uri:         "/test",
		},
		{
			description: "It shouldn't append to an already registered route",
			uri:         "/test",
			err:         ErrRouteAlreadyExists,
		},
		{
			description: "It shouldn't append to an already registered route",
			uri:         "/test/",
			err:         ErrRouteAlreadyExists,
		},
		{
			description: "It shouldn't append to an already registered route",
			uri:         "/test/web",
			err:         ErrRouteAlreadyExists,
		},
		{
			description: "It should append a route with wildcard",
			uri:         "/um/dois/{três}/quatro",
		},
		{
			description: "It shouldn't append an already registered route with the same wildcard",
			uri:         "/um/dois/{três}/quatro",
			err:         ErrRouteAlreadyExists,
		},
		{
			description: "It shouldn't append an already registered route with a wildcard with another name",
			uri:         "/um/dois/{cinco}/quatro",
			err:         ErrWildcardConflict,
		},
		{
			description: "It shouldn't append a route with a constant sibling of a wildcard",
			uri:         "/um/dois/seis/quatro",
			err:         ErrWildcardConflict,
		},
		{
			description: "It should append a long route",
			uri:         "gestos/das/folhas/do/fogo",
		},
		{
			description: "It shouldn't append a route with a wildcard sibling of a constant",
			uri:         "gestos/das/folhas/{da}/flama",
			err:         ErrWildcardConflict,
		},
		{
			description: "It shouldn't append a route with a wildcard sibling of a constant",
			uri:         "gestos/das/folhas/{da}",
			err:         ErrWildcardConflict,
		},
		{
			description: "It should append a route with a constant sibling of another constant",
			uri:         "gestos/das/folhas/de/fogo",
		},
		{
			description: "It should append the web handler route in root",
			uri:         "/",
			handler: adapter{
				webHandler: func() WebHandler { return &DefaultWebHandler{} },
			},
		},
	}

	rt := newRouter()

	for i, item := range data {
		err := rt.appendRoute(item.uri, &item.handler)

		if err != item.err {
			t.Errorf(
				"Item %d, “%s”, unexpected error. Expecting “%s”; found “%s”",
				i,
				item.description,
				item.err,
				err,
			)
		}
	}
}

func TestFindRoute(t *testing.T) {
	var data = []struct {
		description  string
		route        string
		matchURI     string
		uriVars      map[string]string
		handler      *adapter
		shouldntFind bool
	}{
		{
			description: "It should find a simple route",
			route:       "/test/web",
			matchURI:    "/test/web",
			handler:     &adapter{},
		},
		{
			description: "It should find a route with a wildcard",
			route:       "/find/{param}",
			matchURI:    "/find/web",
			uriVars:     map[string]string{"param": "web"},
			handler:     &adapter{},
		},
		{
			description: "It should find a route with multiple wildcards",
			route:       "/route/{x}/{y}",
			matchURI:    "/route/xx/yy",
			uriVars: map[string]string{
				"x": "xx",
				"y": "yy",
			},
			handler: &adapter{},
		},
		{
			description: "It should find a route with multiple wildcards",
			route:       "gestos/{das}/folhas/do/{fogo}",
			matchURI:    "gestos/estas/folhas/do/outono",
			uriVars: map[string]string{
				"das":  "estas",
				"fogo": "outono",
			},
			handler: &adapter{},
		},
		{
			description:  "It shouldn't find a route",
			route:        "gestuais/do/corpo/do/fogo",
			matchURI:     "gestuais/das/mãos",
			handler:      &adapter{},
			shouldntFind: true,
		},
		{
			description:  "It shouldn't find a route with a non-static parent handler",
			route:        "gestuais/do/corpo",
			matchURI:     "gestuais/do/corpo/do/fogo",
			handler:      &adapter{},
			shouldntFind: true,
		},
		{
			description: "It should find a route with a static parent handler",
			route:       "gestuais/do/corpo",
			matchURI:    "gestuais/do/corpo/do/fogo",
			handler: &adapter{
				staticHandler: http.NotFoundHandler(),
			},
		},
	}

	for i, item := range data {
		rt := newRouter()
		err := rt.appendRoute(item.route, item.handler)

		if err != nil {
			t.Errorf("Item %d, “%s”, couldn't append a route: %s", i, item.description, err)
		}

		handler, err := rt.match(item.matchURI)

		if err == ErrRouteNotFound {
			if !item.shouldntFind {
				t.Errorf("Item %d, “%s”, couldn't find a route: %s", i, item.description, err)
			}
		} else if err != nil {
			t.Errorf("Item %d, “%s”, unexpected error: %s", i, item.description, err)
		} else {
			if item.shouldntFind {
				t.Fatalf("Item %d, “%s”, found a route when it shoudn't", i, item.description)
			}

			if !equalTemplateGroupSet(handler.templates, item.handler.templates) {
				t.Errorf("Item %d, “%s”, wrong handler found! Expecting %+v; found %+v", i, item.description, item.handler.templates, handler.templates)
			}

			if item.uriVars != nil {
				for k, v := range item.uriVars {
					found, ok := handler.uriVars[k]

					if !ok {
						t.Errorf("Item %d, “%s”: couldn't find URI parameter %s", i, item.description, k)
					} else if found != v {
						t.Errorf("Item %d, “%s”: wrong URI parameter! Expecting %s; found %s", i, item.description, v, found)
					}

					t.Log(handler.uriVars)
				}
			}
		}
	}
}

func equalTemplateGroupSet(a, b TemplateGroupSet) bool {
	if len(a.elements) != len(b.elements) {
		return false
	}

	for k, v := range a.elements {
		if b.elements[k] != v {
			return false
		}
	}

	return true
}
