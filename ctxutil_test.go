package ctxutil

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeviceID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setup         func() context.Context
		operation     func(context.Context) context.Context
		expectedValue string
		description   string
	}{
		{
			name:          "get from empty context returns empty string",
			setup:         func() context.Context { return context.Background() },
			operation:     func(ctx context.Context) context.Context { return ctx },
			expectedValue: "",
			description:   "Getting DeviceID from empty context should return empty string",
		},
		{
			name: "set and get simple value",
			setup: func() context.Context {
				return context.Background()
			},
			operation: func(ctx context.Context) context.Context {
				return SetDeviceID(ctx, "device-123")
			},
			expectedValue: "device-123",
			description:   "DeviceID should be retrievable after being set",
		},
		{
			name: "override existing value",
			setup: func() context.Context {
				return SetDeviceID(context.Background(), "original-device")
			},
			operation: func(ctx context.Context) context.Context {
				return SetDeviceID(ctx, "new-device")
			},
			expectedValue: "new-device",
			description:   "DeviceID should be updated when set multiple times",
		},
		{
			name: "derived context inherits value",
			setup: func() context.Context {
				return SetDeviceID(context.Background(), "parent-device")
			},
			operation: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, "some-key", "some-value")
			},
			expectedValue: "parent-device",
			description:   "Child context should inherit DeviceID from parent",
		},
		{
			name: "special characters handled correctly",
			setup: func() context.Context {
				return context.Background()
			},
			operation: func(ctx context.Context) context.Context {
				return SetDeviceID(ctx, "device@123:$-&*()")
			},
			expectedValue: "device@123:$-&*()",
			description:   "DeviceID with special characters should be handled correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := tc.setup()
			ctx = tc.operation(ctx)

			assert.Equal(t, tc.expectedValue, GetDeviceID(ctx), tc.description)
		})
	}
}

func TestTraceID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setup         func() context.Context
		operation     func(context.Context) context.Context
		expectedValue string
		description   string
	}{
		{
			name:          "get from empty context returns empty string",
			setup:         func() context.Context { return context.Background() },
			operation:     func(ctx context.Context) context.Context { return ctx },
			expectedValue: "",
			description:   "Getting TraceID from empty context should return empty string",
		},
		{
			name: "set and get simple value",
			setup: func() context.Context {
				return context.Background()
			},
			operation: func(ctx context.Context) context.Context {
				return SetTraceID(ctx, "trace-123")
			},
			expectedValue: "trace-123",
			description:   "TraceID should be retrievable after being set",
		},
		{
			name: "override existing value",
			setup: func() context.Context {
				return SetTraceID(context.Background(), "original-trace")
			},
			operation: func(ctx context.Context) context.Context {
				return SetTraceID(ctx, "new-trace")
			},
			expectedValue: "new-trace",
			description:   "TraceID should be updated when set multiple times",
		},
		{
			name: "derived context inherits value",
			setup: func() context.Context {
				return SetTraceID(context.Background(), "parent-trace")
			},
			operation: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, "some-key", "some-value")
			},
			expectedValue: "parent-trace",
			description:   "Child context should inherit TraceID from parent",
		},
		{
			name: "special characters handled correctly",
			setup: func() context.Context {
				return context.Background()
			},
			operation: func(ctx context.Context) context.Context {
				return SetTraceID(ctx, "trace@123:$-&*()")
			},
			expectedValue: "trace@123:$-&*()",
			description:   "TraceID with special characters should be handled correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := tc.setup()
			ctx = tc.operation(ctx)

			assert.Equal(t, tc.expectedValue, GetTraceID(ctx), tc.description)
		})
	}
}

func TestExtendTimeout(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		setupCtx       func() context.Context
		timeout        time.Duration
		verifyBefore   func(*testing.T, context.Context, context.Context)
		verifyAfter    func(*testing.T, context.Context, context.CancelFunc)
		waitForTimeout bool
	}{
		{
			name: "values are preserved when extending timeout",
			setupCtx: func() context.Context {
				ctx := context.Background()
				ctx = SetDeviceID(ctx, "device-abc")
				ctx = SetTraceID(ctx, "trace-xyz")
				return ctx
			},
			timeout: 50 * time.Millisecond,
			verifyBefore: func(t *testing.T, origCtx, newCtx context.Context) {
				assert.Equal(t, GetDeviceID(origCtx), GetDeviceID(newCtx), "Device ID should be preserved")
				assert.Equal(t, GetTraceID(origCtx), GetTraceID(newCtx), "Trace ID should be preserved")

				deadline, hasDeadline := newCtx.Deadline()
				assert.True(t, hasDeadline, "New context should have a deadline")
				assert.True(t, time.Until(deadline) > 0, "Deadline should be in the future")
			},
			verifyAfter: func(t *testing.T, ctx context.Context, _ context.CancelFunc) {
				assert.NoError(t, ctx.Err(), "Context should not timeout before sleep")
			},
			waitForTimeout: false,
		},
		{
			name: "context correctly times out",
			setupCtx: func() context.Context {
				ctx := context.Background()
				ctx = SetDeviceID(ctx, "device-timeout-test")
				return ctx
			},
			timeout: 10 * time.Millisecond,
			verifyBefore: func(t *testing.T, _, newCtx context.Context) {
				assert.Equal(t, "device-timeout-test", GetDeviceID(newCtx), "Device ID should be preserved")
				assert.NoError(t, newCtx.Err(), "Context should not have error initially")
			},
			verifyAfter: func(t *testing.T, ctx context.Context, _ context.CancelFunc) {
				assert.ErrorIs(t, ctx.Err(), context.DeadlineExceeded, "Context should timeout after sleep")
			},
			waitForTimeout: true,
		},
		{
			name: "cancel works immediately",
			setupCtx: func() context.Context {
				return SetTraceID(context.Background(), "trace-cancel-test")
			},
			timeout: 1 * time.Hour, // long timeout
			verifyBefore: func(t *testing.T, _, newCtx context.Context) {
				assert.Equal(t, "trace-cancel-test", GetTraceID(newCtx), "Trace ID should be preserved")
			},
			verifyAfter: func(t *testing.T, ctx context.Context, cancel context.CancelFunc) {
				// explicitly cancel the context
				cancel()
				assert.ErrorIs(t, ctx.Err(), context.Canceled, "Context should be canceled")
			},
			waitForTimeout: false,
		},
		{
			name: "parent cancellation doesn't affect extended context",
			setupCtx: func() context.Context {
				// create a context that we'll cancel in the verifyAfter function
				// to avoid a goroutine leak from an uncalled cancel function
				parentCtx, parentCancel := context.WithCancel(context.Background())
				parentCtx = SetDeviceID(parentCtx, "device-parent-cancel")

				// store the cancel in the context (using a value) so we can access it later
				type cancelKey struct{}
				parentCtx = context.WithValue(parentCtx, cancelKey{}, parentCancel)

				return parentCtx
			},
			timeout: 20 * time.Millisecond,
			verifyBefore: func(t *testing.T, origCtx, newCtx context.Context) {
				assert.Equal(t, GetDeviceID(origCtx), GetDeviceID(newCtx), "Device ID should be preserved")
			},
			verifyAfter: func(t *testing.T, ctx context.Context, _ context.CancelFunc) {
				type cancelKey struct{}
				if parentCancel, ok := ctx.Value(cancelKey{}).(context.CancelFunc); ok {
					parentCancel() // cancel the parent context
				}
				synctest.Wait() // wait for cancellation to propagate
				assert.NoError(t, ctx.Err(), "Extended context should not be affected by parent cancellation")
			},
			waitForTimeout: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			synctest.Run(func() {
				origCtx := tc.setupCtx()

				newCtx, cancel := ExtendTimeout(origCtx, tc.timeout)
				defer cancel()

				tc.verifyBefore(t, origCtx, newCtx)

				if tc.waitForTimeout {
					time.Sleep(tc.timeout + time.Millisecond)
					synctest.Wait()
				}

				tc.verifyAfter(t, newCtx, cancel)
			})
		})
	}
}

func TestExtendTimeoutSharedValues(t *testing.T) {
	t.Parallel()

	synctest.Run(func() {
		// verifies that extended contexts share the same values store
		// with the parent context, so changes in one are visible in the other

		// setup parent context with values
		parentCtx := context.Background()
		parentCtx = SetDeviceID(parentCtx, "original-device")
		parentCtx = SetTraceID(parentCtx, "original-trace")

		// create an extended context
		extendedCtx, cancel := ExtendTimeout(parentCtx, 100*time.Millisecond)
		defer cancel()

		// parent values were carried over
		assert.Equal(t, "original-device", GetDeviceID(extendedCtx))
		assert.Equal(t, "original-trace", GetTraceID(extendedCtx))

		// modifying the extended context WILL affect the parent because they share the same pointer
		extendedCtx = SetDeviceID(extendedCtx, "new-device")
		assert.Equal(t, "new-device", GetDeviceID(extendedCtx), "Extended context's device ID should be updated")
		assert.Equal(t, "new-device", GetDeviceID(parentCtx), "Parent context's device ID should also be updated")

		// modifying the parent context WILL affect the extended context
		parentCtx = SetTraceID(parentCtx, "updated-trace")
		assert.Equal(t, "updated-trace", GetTraceID(parentCtx), "Parent context's trace ID should be updated")
		assert.Equal(t, "updated-trace", GetTraceID(extendedCtx), "Extended context's trace ID should also be updated")

		// canceling the extended context should not affect the parent context's cancellation state
		cancel()
		synctest.Wait()
		assert.ErrorIs(t, extendedCtx.Err(), context.Canceled, "Extended context should be canceled")
		assert.NoError(t, parentCtx.Err(), "Parent context should remain valid")
	})
}
