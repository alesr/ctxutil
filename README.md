# ctxutil

A tiny Go package that makes working with `context.Context` values less painful.

## Why?

Working with context values in Go traditionally requires a lot of boilerplate and type assertions:

```go
// 🤢 without ctxutil
type deviceIDKey struct{}
ctx = context.WithValue(ctx, deviceIDKey{}, "device-123")
deviceID, _ := ctx.Value(deviceIDKey{}).(string) // type assertion every time

// 😎 With ctxutil
ctx = ctxutil.SetDeviceID(ctx, "device-123")
deviceID := ctxutil.GetDeviceID(ctx) // strongly typed, no assertions
```

Additionally, extending a context timeout while preserving values is annoying:

```go
// 🤢 without ctxutil values don't carry over
newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// Now you need to manually copy all values...

// 😎 with ctxutil
newCtx, cancel := ctxutil.ExtendTimeout(ctx, 5*time.Second) // values are preserved
// works for both adding a timeout to a context without one
// and for replacing/extending an existing timeout
```

## Install

```
go get github.com/alesr/ctxutil
```

## Usage

```go
import (
    "context"
    "github.com/alesr/ctxutil"
)

func main() {
    // store values in context
    ctx := context.Background()
    ctx = ctxutil.SetDeviceID(ctx, "device-123")
    ctx = ctxutil.SetTraceID(ctx, "trace-abc")

    // retrieve values
    deviceID := ctxutil.GetDeviceID(ctx) // "device-123"
    traceID := ctxutil.GetTraceID(ctx)   // "trace-abc"

    // add or extend context timeout while preserving values
    // Works whether ctx already has a timeout or not
    newCtx, cancel := ctxutil.ExtendTimeout(ctx, 5*time.Second)
    defer cancel()

    // values are still accessible
    deviceID = ctxutil.GetDeviceID(newCtx) // "device-123"
}
```

## License

MIT
