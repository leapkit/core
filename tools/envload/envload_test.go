package envload

import (
	"strings"
	"testing"
)

func TestParseVars(t *testing.T) {

	t.Run("simple one", func(t *testing.T) {
		r := strings.NewReader("KEY=value\n")
		vars := parseVars(r)

		if vars["KEY"] != "value" {
			t.Errorf("Expected value to be 'value', got %s", vars["KEY"])
		}
	})

	t.Run("multiple", func(t *testing.T) {
		vars := parseVars(strings.NewReader(`
			KEY=value
			KEY2=value
		`))

		if vars["KEY"] != "value" {
			t.Errorf("Expected value to be 'value', got %s", vars["KEY"])
		}

		if vars["KEY2"] != "value" {
			t.Errorf("Expected value to be 'value', got %s", vars["KEY"])
		}
	})

	t.Run("quotes", func(t *testing.T) {
		vars := parseVars(strings.NewReader(`
			KEY="value"
		`))

		if vars["KEY"] != "value" {
			t.Errorf("Expected value to be 'value', got %s", vars["KEY"])
		}
	})

	t.Run("multiple equals sign", func(t *testing.T) {
		vars := parseVars(strings.NewReader(`
			KEY="value=with=equals"
		`))

		if vars["KEY"] != "value=with=equals" {
			t.Errorf("Expected value to be 'value=with=equals', got %s", vars["KEY"])
		}
	})

}
