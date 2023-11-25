package render

import (
	"github.com/leapkit/core/internal/hctx"
	"github.com/leapkit/core/internal/helpers/content"
	"github.com/leapkit/core/internal/helpers/debug"
	"github.com/leapkit/core/internal/helpers/encoders"
	"github.com/leapkit/core/internal/helpers/env"
	"github.com/leapkit/core/internal/helpers/escapes"
	"github.com/leapkit/core/internal/helpers/forms"
	"github.com/leapkit/core/internal/helpers/inflections"
	"github.com/leapkit/core/internal/helpers/iterators"
	"github.com/leapkit/core/internal/helpers/meta"
	"github.com/leapkit/core/internal/helpers/paths"
	"github.com/leapkit/core/internal/helpers/text"
	"github.com/leapkit/core/internal/plush"
)

// HelperContext is an alias for plush.HelperContext
type HelperContext plush.HelperContext

// HelperMap is an alias for hctx.Map
type HelperMap hctx.Map

// AllHelpers contains all of the default helpers for
// These will be available to all templates.
var AllHelpers = MergeHelpers(
	content.New(),
	debug.New(),
	encoders.New(),
	env.New(),
	escapes.New(),
	inflections.New(),
	iterators.New(),
	meta.New(),
	paths.New(),
	text.New(),
	forms.New(),
)

// MergeHelpers merges multiple maps of helpers into one
// map. If a helper exists in multiple maps, the last one
// wins.
func MergeHelpers(helpers ...map[string]any) HelperMap {
	mx := map[string]interface{}{}
	for _, m := range helpers {
		for k, v := range m {
			mx[k] = v
		}
	}

	return mx
}
