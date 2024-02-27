package assets_test

import (
	"testing"
	"testing/fstest"

	"github.com/leapkit/core/assets"
)

func TestManagerFingerprint(t *testing.T) {
	m := assets.NewManager(fstest.MapFS{
		"public/main.js": {Data: []byte("AAA")},
	})

	a := m.PathFor("public/main.js")
	b := m.PathFor("public/main.js")

	if a != b {
		t.Errorf("Expected %s to equal %s", a, b)
	}
}
