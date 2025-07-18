package ctxutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		setupCtx       func() context.Context
		expectedDevice string
		expectedTrace  string
		description    string
	}{
		{
			name:           "returns empty struct for background context",
			setupCtx:       func() context.Context { return context.Background() },
			expectedDevice: "",
			expectedTrace:  "",
			description:    "Background context should return empty values",
		},
		{
			name: "returns correct values when properly set",
			setupCtx: func() context.Context {
				ctx := context.Background()
				vals := &contextValues{
					deviceID: "device-123",
					traceID:  "trace-456",
				}
				return context.WithValue(ctx, contextKey{}, vals)
			},
			expectedDevice: "device-123",
			expectedTrace:  "trace-456",
			description:    "Should retrieve exact values that were set",
		},
		{
			name: "handles incorrect type gracefully",
			setupCtx: func() context.Context {
				ctx := context.Background()
				return context.WithValue(ctx, contextKey{}, "wrong-type-value")
			},
			expectedDevice: "",
			expectedTrace:  "",
			description:    "Should handle incorrect value type gracefully",
		},
		{
			name: "works with derived contexts",
			setupCtx: func() context.Context {
				// Start with values
				ctx := context.Background()
				vals := &contextValues{
					deviceID: "original-device",
					traceID:  "original-trace",
				}
				ctx = context.WithValue(ctx, contextKey{}, vals)

				// Add some other values to create a derived context
				ctx = context.WithValue(ctx, "other-key", 12345)
				return context.WithValue(ctx, "yet-another-key", true)
			},
			expectedDevice: "original-device",
			expectedTrace:  "original-trace",
			description:    "Should retrieve values through context inheritance chain",
		},
		{
			name: "nil value handled properly",
			setupCtx: func() context.Context {
				ctx := context.Background()
				return context.WithValue(ctx, contextKey{}, nil)
			},
			expectedDevice: "",
			expectedTrace:  "",
			description:    "Should handle nil value gracefully",
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := tc.setupCtx()
			vals := getValues(ctx)

			require.NotNil(t, vals, "getValues should never return nil")
			assert.Equal(t, tc.expectedDevice, vals.deviceID, tc.description+" (deviceID)")
			assert.Equal(t, tc.expectedTrace, vals.traceID, tc.description+" (traceID)")
		})
	}
}

func TestWithValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		setupCtx    func() context.Context
		setupVals   func() *contextValues
		verifyFunc  func(*testing.T, context.Context, *contextValues)
		description string
	}{
		{
			name: "adds values to empty context",
			setupCtx: func() context.Context {
				return context.Background()
			},
			setupVals: func() *contextValues {
				return &contextValues{
					deviceID: "device-123",
					traceID:  "trace-456",
				}
			},
			verifyFunc: func(t *testing.T, ctx context.Context, vals *contextValues) {
				retrieved, ok := ctx.Value(contextKey{}).(*contextValues)
				require.True(t, ok, "Should retrieve correct type")
				assert.Equal(t, vals, retrieved, "Retrieved values should match what was set")
			},
			description: "Values should be retrievable from the context",
		},
		{
			name: "replaces existing values",
			setupCtx: func() context.Context {
				original := &contextValues{
					deviceID: "original-device",
					traceID:  "original-trace",
				}
				return context.WithValue(context.Background(), contextKey{}, original)
			},
			setupVals: func() *contextValues {
				return &contextValues{
					deviceID: "new-device",
					traceID:  "new-trace",
				}
			},
			verifyFunc: func(t *testing.T, ctx context.Context, vals *contextValues) {
				retrieved, ok := ctx.Value(contextKey{}).(*contextValues)
				require.True(t, ok, "Should retrieve correct type")
				assert.Equal(t, vals, retrieved, "Retrieved values should replace original values")
			},
			description: "New values should replace existing values",
		},
		{
			name: "preserves other context values",
			setupCtx: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, "other-key", "other-value")
				ctx = context.WithValue(ctx, 42, true) // Different key type
				return ctx
			},
			setupVals: func() *contextValues {
				return &contextValues{deviceID: "test-device"}
			},
			verifyFunc: func(t *testing.T, ctx context.Context, vals *contextValues) {
				// Check our values were set
				retrieved, ok := ctx.Value(contextKey{}).(*contextValues)
				require.True(t, ok, "Should retrieve correct type")
				assert.Equal(t, vals, retrieved, "Retrieved values should match what was set")

				// Check other values still exist
				assert.Equal(t, "other-value", ctx.Value("other-key"), "Other string key should be preserved")
				assert.Equal(t, true, ctx.Value(42), "Other int key should be preserved")
			},
			description: "Other context values should be preserved",
		},
		{
			name: "handles nil values",
			setupCtx: func() context.Context {
				return context.Background()
			},
			setupVals: func() *contextValues {
				return nil
			},
			verifyFunc: func(t *testing.T, ctx context.Context, _ *contextValues) {
				// When nil is provided, nil should be stored
				val := ctx.Value(contextKey{})
				assert.Nil(t, val, "Nil value should be stored as nil")
			},
			description: "Nil values should be handled properly",
		},
		{
			name: "values are copy-independent",
			setupCtx: func() context.Context {
				return context.Background()
			},
			setupVals: func() *contextValues {
				return &contextValues{
					deviceID: "mutable-device",
					traceID:  "mutable-trace",
				}
			},
			verifyFunc: func(t *testing.T, ctx context.Context, vals *contextValues) {
				// Get the initial values
				retrieved, ok := ctx.Value(contextKey{}).(*contextValues)
				require.True(t, ok, "Should retrieve correct type")
				assert.Equal(t, vals, retrieved, "Initial values should match")

				// Modify the original values
				vals.deviceID = "modified-device"
				vals.traceID = "modified-trace"

				// The context should still have the pointer to the same struct
				// which means it will reflect the changes
				retrieved2 := ctx.Value(contextKey{}).(*contextValues)
				assert.Equal(t, vals, retrieved2, "Context stores reference, so should see changes")
			},
			description: "Context stores a reference to the values struct",
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			vals := tc.setupVals()
			origCtx := tc.setupCtx()

			// Create new context with values
			newCtx := withValues(origCtx, vals)

			// Original context should be unchanged
			if vals != nil {
				origVal := origCtx.Value(contextKey{})
				if origVal != nil {
					assert.NotEqual(t, vals, origVal, "Original context should be unchanged")
				}
			}

			// Run test-specific verification
			tc.verifyFunc(t, newCtx, vals)
		})
	}
}

func TestSetString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		initialCtx  func() context.Context
		setValue    string
		setterField func(*contextValues, string)
		getterField func(*contextValues) string
		verify      func(*testing.T, context.Context, string)
		description string
	}{
		{
			name: "initializes values struct in empty context",
			initialCtx: func() context.Context {
				return context.Background()
			},
			setValue:    "new-device-id",
			setterField: func(v *contextValues, s string) { v.deviceID = s },
			getterField: func(v *contextValues) string { return v.deviceID },
			verify: func(t *testing.T, ctx context.Context, value string) {
				vals := getValues(ctx)
				assert.Equal(t, value, vals.deviceID)
				assert.Empty(t, vals.traceID, "Other fields should remain empty")
			},
			description: "When setting on empty context, should initialize values struct",
		},
		{
			name: "updates existing value without affecting others",
			initialCtx: func() context.Context {
				// Setup context with both values set
				vals := &contextValues{
					deviceID: "original-device",
					traceID:  "original-trace",
				}
				return context.WithValue(context.Background(), contextKey{}, vals)
			},
			setValue:    "updated-device",
			setterField: func(v *contextValues, s string) { v.deviceID = s },
			getterField: func(v *contextValues) string { return v.deviceID },
			verify: func(t *testing.T, ctx context.Context, value string) {
				vals := getValues(ctx)
				assert.Equal(t, value, vals.deviceID, "Target field should be updated")
				assert.Equal(t, "original-trace", vals.traceID, "Other fields should be preserved")
			},
			description: "Updating one field should preserve other fields",
		},
		{
			name: "creates new context instance",
			initialCtx: func() context.Context {
				ctx := context.Background()
				deadline := time.Now().Add(10 * time.Millisecond)
				timeoutCtx, cancel := context.WithDeadline(ctx, deadline)
				defer cancel() // Properly cancel to avoid context leak
				vals := &contextValues{traceID: "trace-on-timeout"}
				return context.WithValue(timeoutCtx, contextKey{}, vals)
			},
			setValue:    "device-with-timeout",
			setterField: func(v *contextValues, s string) { v.deviceID = s },
			getterField: func(v *contextValues) string { return v.deviceID },
			verify: func(t *testing.T, ctx context.Context, value string) {
				// Verify the new context has the same deadline
				deadline, hasDeadline := ctx.Deadline()
				assert.True(t, hasDeadline, "New context should preserve deadline")
				assert.True(t, deadline.After(time.Now()), "Deadline should be in the future")

				// Verify value was set
				vals := getValues(ctx)
				assert.Equal(t, value, vals.deviceID)
				assert.Equal(t, "trace-on-timeout", vals.traceID)
			},
			description: "Should preserve context properties like deadlines",
		},
		{
			name: "empty string clears previous value",
			initialCtx: func() context.Context {
				vals := &contextValues{traceID: "existing-trace-value"}
				return context.WithValue(context.Background(), contextKey{}, vals)
			},
			setValue:    "", // empty string
			setterField: func(v *contextValues, s string) { v.traceID = s },
			getterField: func(v *contextValues) string { return v.traceID },
			verify: func(t *testing.T, ctx context.Context, _ string) {
				vals := getValues(ctx)
				assert.Empty(t, vals.traceID, "Value should be cleared to empty string")
			},
			description: "Setting empty string should clear previous value",
		},
		{
			name: "handles special characters correctly",
			initialCtx: func() context.Context {
				return context.Background()
			},
			setValue:    "特殊文字@#$%^&*()",
			setterField: func(v *contextValues, s string) { v.deviceID = s },
			getterField: func(v *contextValues) string { return v.deviceID },
			verify: func(t *testing.T, ctx context.Context, value string) {
				vals := getValues(ctx)
				assert.Equal(t, value, vals.deviceID, "Special characters should be preserved exactly")
			},
			description: "Should handle special and non-ASCII characters correctly",
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := tc.initialCtx()
			newCtx := setString(ctx, tc.setterField, tc.setValue)

			// Should get a new context instance
			assert.NotEqual(t, ctx, newCtx, "setString should return a new context instance")

			// Run test-specific verification
			tc.verify(t, newCtx, tc.setValue)
		})
	}
}

func TestGetString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupCtx      func() context.Context
		getterField   func(*contextValues) string
		expectedValue string
		description   string
	}{
		{
			name: "returns empty string for missing context values",
			setupCtx: func() context.Context {
				return context.Background()
			},
			getterField:   func(v *contextValues) string { return v.deviceID },
			expectedValue: "",
			description:   "Should return empty string when context has no values",
		},
		{
			name: "returns empty string for wrong value type",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), contextKey{}, "not-a-contextValues")
			},
			getterField:   func(v *contextValues) string { return v.deviceID },
			expectedValue: "",
			description:   "Should return empty string when context value is wrong type",
		},
		{
			name: "returns value when available",
			setupCtx: func() context.Context {
				vals := &contextValues{deviceID: "valid-device-id"}
				return context.WithValue(context.Background(), contextKey{}, vals)
			},
			getterField:   func(v *contextValues) string { return v.deviceID },
			expectedValue: "valid-device-id",
			description:   "Should return correct value when available",
		},
		{
			name: "returns empty string when field is empty",
			setupCtx: func() context.Context {
				// Set other field but not the one we're querying
				vals := &contextValues{traceID: "some-trace"}
				return context.WithValue(context.Background(), contextKey{}, vals)
			},
			getterField:   func(v *contextValues) string { return v.deviceID },
			expectedValue: "",
			description:   "Should return empty string when specific field is empty",
		},
		{
			name: "correctly retrieves value through context chain",
			setupCtx: func() context.Context {
				// Start with a context with values
				vals := &contextValues{traceID: "inherited-trace"}
				ctx := context.WithValue(context.Background(), contextKey{}, vals)

				// Create child contexts with different values
				ctx = context.WithValue(ctx, "level1", true)
				ctx = context.WithValue(ctx, "level2", 42)
				return context.WithValue(ctx, "level3", "value")
			},
			getterField:   func(v *contextValues) string { return v.traceID },
			expectedValue: "inherited-trace",
			description:   "Should retrieve value through multi-level context chain",
		},
		{
			name: "handles nil value in context",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), contextKey{}, nil)
			},
			getterField:   func(v *contextValues) string { return v.deviceID },
			expectedValue: "",
			description:   "Should handle nil value in context gracefully",
		},
		{
			name: "preserves exact string content including special chars",
			setupCtx: func() context.Context {
				vals := &contextValues{deviceID: "特殊文字@#$%^&*()"}
				return context.WithValue(context.Background(), contextKey{}, vals)
			},
			getterField:   func(v *contextValues) string { return v.deviceID },
			expectedValue: "特殊文字@#$%^&*()",
			description:   "Should preserve special characters exactly",
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := tc.setupCtx()
			value := getString(ctx, tc.getterField)

			assert.Equal(t, tc.expectedValue, value, tc.description)
		})
	}
}

// TestGettersAndSettersTogether verifies that getString and setString work together correctly
func TestGettersAndSettersTogether(t *testing.T) {
	t.Parallel()

	// Start with empty context
	ctx := context.Background()

	// Define our field accessors
	deviceGetter := func(v *contextValues) string { return v.deviceID }
	deviceSetter := func(v *contextValues, s string) { v.deviceID = s }
	traceGetter := func(v *contextValues) string { return v.traceID }
	traceSetter := func(v *contextValues, s string) { v.traceID = s }

	// Verify empty initially
	assert.Empty(t, getString(ctx, deviceGetter), "Initial deviceID should be empty")
	assert.Empty(t, getString(ctx, traceGetter), "Initial traceID should be empty")

	// Set one value
	ctx = setString(ctx, deviceSetter, "device-first")
	assert.Equal(t, "device-first", getString(ctx, deviceGetter), "DeviceID should be set")
	assert.Empty(t, getString(ctx, traceGetter), "TraceID should still be empty")

	// Set the other value
	ctx = setString(ctx, traceSetter, "trace-second")
	assert.Equal(t, "device-first", getString(ctx, deviceGetter), "DeviceID should be unchanged")
	assert.Equal(t, "trace-second", getString(ctx, traceGetter), "TraceID should be set")

	// Update first value
	ctx = setString(ctx, deviceSetter, "device-updated")
	assert.Equal(t, "device-updated", getString(ctx, deviceGetter), "DeviceID should be updated")
	assert.Equal(t, "trace-second", getString(ctx, traceGetter), "TraceID should be unchanged")

	// Clear second value
	ctx = setString(ctx, traceSetter, "")
	assert.Equal(t, "device-updated", getString(ctx, deviceGetter), "DeviceID should be unchanged")
	assert.Empty(t, getString(ctx, traceGetter), "TraceID should be cleared")
}
