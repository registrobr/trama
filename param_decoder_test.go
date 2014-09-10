package trama

import (
	"strconv"
	"testing"
)

func TestDecoder(t *testing.T) {
	type testStr struct {
		Name      string `param:"nome"`
		Number    int    `param:"número"`
		NotTagged int
		BigNumber int64 `param:"numerão"`
	}

	data := []struct {
		description string
		structure   testStr
		result      testStr
		uriParams   map[string]string
		errorFunc   func(error, *testing.T)
	}{
		{
			description: "It should inject the values of the variables",
			uriParams: map[string]string{
				"nome":    "Dezessete",
				"número":  "17",
				"numerão": "137438953472",
			},
			result: testStr{
				Name:      "Dezessete",
				Number:    17,
				BigNumber: 137438953472,
			},
		},
		{
			description: "It should leave Number blank",
			uriParams: map[string]string{
				"nome":    "Dezessete",
				"numerão": "137438953472",
			},
			result: testStr{
				Name:      "Dezessete",
				Number:    0,
				BigNumber: 137438953472,
			},
		},
		{
			description: "It should generate a ParseInt error",
			uriParams: map[string]string{
				"nome":    "Dezessete",
				"número":  "Dezessete",
				"numerão": "big dum número",
			},
			result: testStr{Name: "Dezessete"},
			errorFunc: func(err error, t *testing.T) {
				if _, ok := err.(*strconv.NumError); !ok {
					t.Errorf("Unexpected error type. Expecting *strconv.NumError; found %T. Error: “%s”", err, err)
				}
			},
		},
	}

	for i, item := range data {
		errorFuncCalled := false

		errorFunc := func(err error) {
			errorFuncCalled = true

			if item.errorFunc != nil {
				item.errorFunc(err, t)
			} else {
				t.Errorf("Item %d, “%s”, error func called when it shouldn't: %s", i, item.description, err)
			}
		}

		newParamDecoder(&item.structure, item.uriParams, errorFunc).decode()

		if item.result != item.structure {
			t.Errorf("Item %d, “%s”, wrong result. Expecting %+v; found %+v", i, item.description, item.result, item.structure)
		}

		if item.errorFunc != nil && !errorFuncCalled {
			t.Errorf("Item %d, “%s”, didn't call errorFunc when it should", i, item.description)
		}
	}
}
