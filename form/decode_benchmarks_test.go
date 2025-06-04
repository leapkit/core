package form_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"go.leapkit.dev/core/form"
)

func BenchmarkSimpleDecodeStruct(b *testing.B) {
	type Simple struct {
		FieldA string `form:"field_a"`
		FieldB int    `form:"field_b"`
		FieldC float64
		FieldD time.Time `form:"fieldD"`
	}

	values := url.Values{
		"field_a": []string{"John Smith"},
		"field_b": []string{"30"},
		"FieldC":  []string{"36.5"},
		"FieldD":  []string{"2025-06-26"},
	}

	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		var s Simple

		if err := form.Decode(req, &s); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkComplexDecodeStruct(b *testing.B) {
	type Complex struct {
		FieldA string `form:"field_a"`
		FieldB int    `form:"field_b"`
		FieldC float64
		FieldD string `form:"fieldD"`
		FieldE *struct {
			FieldF []string `form:"Field-F"`
			FieldG *[]int
			FieldH []*int `form:"hField"`
		}
		FieldI []struct {
			FieldJ string
		}
		FieldK []*struct {
			FieldL *int
		}
		FieldM any
	}

	values := url.Values{
		"field_a":           []string{"John Smith"},
		"field_b":           []string{"30"},
		"FieldC":            []string{"36.5"},
		"fieldD":            []string{"2025-06-26"},
		"FieldE.Field-F[0]": []string{"Hello"},
		"FieldE.Field-F[1]": []string{"World"},
		"FieldE.Field-F[2]": []string{"Foo"},
		"FieldE.Field-F[3]": []string{"Bar"},
		"FieldE.FieldG[0]":  []string{"1"},
		"FieldE.FieldG[1]":  []string{"2"},
		"FieldE.hField[0]":  []string{"100"},
		"FieldE.hField[1]":  []string{"200"},
		"FieldI[0].FieldJ":  []string{"Value A"},
		"FieldI[1].FieldJ":  []string{"Value B"},
		"FieldK[0].FieldL":  []string{"10"},
		"FieldK[1].FieldL":  []string{"20"},
		"FieldM":            []string{"any-value"},
	}

	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		var c Complex
		if err := form.Decode(req, &c); err != nil {
			b.Error(err)
		}
	}
}
