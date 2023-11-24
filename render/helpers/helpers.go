package helpers

import (
	"github.com/leapkit/core/internal/plush/hctx"
	"github.com/leapkit/core/render/helpers/content"
	"github.com/leapkit/core/render/helpers/debug"
	"github.com/leapkit/core/render/helpers/encoders"
	"github.com/leapkit/core/render/helpers/env"
	"github.com/leapkit/core/render/helpers/escapes"
	"github.com/leapkit/core/render/helpers/inflections"
	"github.com/leapkit/core/render/helpers/iterators"
	"github.com/leapkit/core/render/helpers/meta"
	"github.com/leapkit/core/render/helpers/paths"
	"github.com/leapkit/core/render/helpers/text"
)

var base = []hctx.Map{
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
}

var ALL = func() hctx.Map {
	return hctx.Merge(base...)
}
