package render

import (
	"context"
)

// FromCtx returns the render engine from the context.
func FromCtx(ctx context.Context) *Page {
	return ctx.Value("renderer").(*Page)
}
