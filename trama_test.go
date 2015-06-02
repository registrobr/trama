package trama_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"trama"
)

type interceptorA struct{}

func (i interceptorA) Before(trama.Response, *http.Request) error {
	fmt.Println("E como eu palmilhasse vagamente")
	return nil
}

func (i interceptorA) After(trama.Response, *http.Request, error) {
	fmt.Println("no céu de chumbo, e suas formas pretas")
}

type interceptorB struct{}

func (i interceptorB) Before(trama.Response, *http.Request) error {
	fmt.Println("uma estrada de Minas, pedregosa,")
	return nil
}

func (i interceptorB) After(trama.Response, *http.Request, error) {
	fmt.Println("que era pausado e seco; e aves pairassem")
}

type interceptorC struct{}

func (i interceptorC) Before(trama.Response, *http.Request) error {
	fmt.Println("e no fecho da tarde um sino rouco")
	return nil
}

func (i interceptorC) After(trama.Response, *http.Request, error) {
	fmt.Println("se misturasse ao som de meus sapatos")
}

type handler struct {
	trama.NopHandler
}

func (h *handler) Get(resp trama.Response, req *http.Request) error {
	fmt.Println("")
	resp.Redirect("/sai-daqui", http.StatusSeeOther)
	return nil
}

func (h *handler) Interceptors() trama.InterceptorChain {
	return trama.NewInterceptorChain(
		interceptorA{},
		interceptorB{},
		interceptorC{},
	)
}

func Example() {
	mux := trama.NewMux()
	mux.Register("/vai-prali", func() trama.Handler { return &handler{} })
	server := httptest.NewServer(mux)
	defer server.Close()

	http.Get(server.URL + "/vai-prali")
	// Output:
	// E como eu palmilhasse vagamente
	// uma estrada de Minas, pedregosa,
	// e no fecho da tarde um sino rouco
	//
	// se misturasse ao som de meus sapatos
	// que era pausado e seco; e aves pairassem
	// no céu de chumbo, e suas formas pretas
}
