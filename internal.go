package ctxutil

import "context"

// setString sets a string value for a field in the context.
func setString(ctx context.Context, setter func(*contextValues, string), value string) context.Context {
	vals := getValues(ctx)
	if vals == nil {
		vals = &contextValues{}
	}
	setter(vals, value)
	return withValues(ctx, vals)
}

// getString gets a string value for a field from the context.
func getString(ctx context.Context, getter func(*contextValues) string) string {
	vals := getValues(ctx)
	if vals == nil {
		return ""
	}
	return getter(vals)
}

// getValues retrieves the contextValues from the context.
func getValues(ctx context.Context) *contextValues {
	val, ok := ctx.Value(contextKey{}).(*contextValues)
	if !ok {
		return &contextValues{}
	}
	return val
}

// withValues creates a fresh context with the given values.
func withValues(ctx context.Context, vals *contextValues) context.Context {
	return context.WithValue(ctx, contextKey{}, vals)
}
