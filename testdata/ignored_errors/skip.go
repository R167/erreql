package ignored_errors

import (
	"context"
	"database/sql"
	"errors"
)

var ignoredErrors = []error{
	sql.ErrNoRows,
	context.DeadlineExceeded,
}

func maybeSwallow(err error) error {
	for _, skip := range ignoredErrors {
		if err == skip {
			// accepted by the linter. skip is treated as sentinel
			return nil
		}
		if skipErr := skip; err == skipErr { // want `use errors\.Is or errors\.As instead of ==`
			// LINT ERROR: Expected errors.Is but got ==
			return nil
		} else if errors.Is(err, skipErr) {
			// linter is happy again
			return nil
		}
	}
	return err
}

var _ = maybeSwallow
