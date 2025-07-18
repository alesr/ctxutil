package ctxutil

import (
	"context"
	"time"
)

type contextKey struct{}

type contextValues struct {
	deviceID string
	traceID  string
	// moar fields as needed
}

// SetDeviceID sets the device ID in the context.
func SetDeviceID(ctx context.Context, deviceID string) context.Context {
	return setString(ctx, func(v *contextValues, s string) { v.deviceID = s }, deviceID)
}

// GetDeviceID gets the device ID from the context.
func GetDeviceID(ctx context.Context) string {
	return getString(ctx, func(v *contextValues) string { return v.deviceID })
}

// SetTraceID sets the trace ID in the context.
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return setString(ctx, func(v *contextValues, s string) { v.traceID = s }, traceID)
}

// GetTraceID gets the trace ID from the context.
func GetTraceID(ctx context.Context) string {
	return getString(ctx, func(v *contextValues) string { return v.traceID })
}

// ExtendTimeout creates a fresh context with the given timeout
// and carries over known values from the original context.
func ExtendTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	newCtx, cancel := context.WithTimeout(context.Background(), timeout)
	vals := getValues(ctx)
	if vals != nil {
		newCtx = withValues(newCtx, vals)
	}
	return newCtx, cancel
}
