package render

import (
	"github.com/leapkit/core/hctx"
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
)

// AllHelpers contains all of the default helpers for
// These will be available to all templates.
var AllHelpers = hctx.Merge(
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
