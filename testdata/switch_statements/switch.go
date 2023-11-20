package switchstatements

import (
	"context"
	"io/fs"
)

func _() {
	var err error
	switch err {
	case nil:
	case fs.ErrClosed: // want `switch does not handle wrapped errors`
	default:
	}

	switch e := context.Canceled; e {
	case nil:
	case err: // want `switch does not handle wrapped errors`
	default:
	}
}
