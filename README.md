# erreql
An anaylsis to check for usages of checking error types using `==` equality instead of `errors.Is` introduced in [go1.13](https://go.dev/doc/go1.13#error_wrapping).

When passing errors in an application, it's useful to capture stack traces of where an error was encountered by wrapping the original error. However, this breaks equality checks for statically defined errors. For example:

```go
package main

import (
  "errors"
  "runtime/debug"
)

var ErrNotFound = errors.New("not found")

func database() error {
  return ErrNotFound
}

func helper() error {
  if err := database(); err != nil {
    return WithTrace(err)
  }
  return nil
}

func controllerB() error {
  return helper()
}

func main() {
  err := controllberB()
  if err == nil {
    // untyped nil is fine to compare against
  } else if err == ErrNotFound {
    // Oh no! b/c helper() wraps the database error, we never hit this path
  } else if errors.Is(err, ErrNotFound) {
    // This is the correct way to check the underlying type of an error
  }
}

type TraceError struct {
  err   error
  stack []byte
}

func (e TraceError) Error() string { return e.Error() }
func (e TraceError) Trace() []byte { return e.stack }
func (e TraceError) Unwrap() error { return e.err }

func WithTrace(err error) error { return TraceError{err, debug.Stack()} }
```

However, this package makes an exception for special "sentinel error value" types which generally indicate an end of normal operations when calling the function directly. For example, `io.EOF` is returned to indicate the end of an input stream

```go
// ReadLength reads upto n bytes from r
func ReadLength(n int, r io.Reader) ([]byte, error) {
	b := make([]byte, 0, n)
	for {
		c, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+c]
		if err != nil {
      // allow sentinel values to use `==` checks
			if err == io.EOF {
				return b, nil
			}
			return b, err
		}

		if len(b) == cap(b) {
			// At capacity
			return b, nil
		}
	}
}
```

Sentinel values are currently defined as
- Comparison of an identifier which implements `error`
- Identifier name does NOT match `^err.|Err|Exception` e.g.
  - `db.ErrNotFound` - use errors.Is
  - `internalError` - use errors.Is
  - `cursor.EndOfData` - sentinel value

In general, errors in golang should follow the naming pattern `errName/ErrName` so this case should be fairly reasonable in most scenarios, however there are edge casese to be mindful of (and note this is best effort)

```go
var ignoredErrors = []error{
  db.ErrNotFound,
  context.DeadlineExceeded,
}

func maybeSwallow(err error) error {
  for _, skip := range ignoredErrors {
    if err == skip {
      // accepted by the linter. skip is treated as sentinel
      return nil
    }
    if skipErr := skip; err == skipErr {
      // LINT ERROR: Expected errors.Is but got ==
      return nil
    } else if errors.Is(err, skipErr) {
      // linter is happy again
      return nil
    }
  }
  return err
}
```
