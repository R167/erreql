package test1

import (
	"errors"
	"io/fs"
)

type myError int

func (*myError) Error() string { return "" }

var errType = errors.New("some error")

var constantType error

type specificError struct{}

func (specificError) Error() string { return "" }

func getError() error {
	return nil
}

func _() {
	var (
		err         error
		myErr       *myError
		specificErr specificError
		i           interface{}
	)

	_ = errors.Is(err, nil) // excessive, but fine
	_ = err == nil          // == allowed for nil
	_ = myErr == nil        // == allowed for nil
	_ = err != nil          // != allowed for nil
	_ = myErr != nil        // != allowed for nil

	_ = err == errType   // want `use errors\.Is or errors\.As instead of ==`
	_ = err != errType   // want `use errors\.Is or errors\.As instead of !=`
	_ = myErr == errType // want `use errors\.Is or errors\.As instead of ==`
	_ = myErr != errType // want `use errors\.Is or errors\.As instead of !=`

	_ = err == specificErr // want `use errors\.Is or errors\.As instead of ==`
	_ = err != specificErr // want `use errors\.Is or errors\.As instead of !=`

	_ = err == i // Fine, not an error type

	_ = err == constantType // treated as a sentinel error value

	if err := getError(); err != nil { // Fine as usual
		// first off, you're probably doing something wrong here...
		_ = err == getError() // want `use errors\.Is or errors\.As instead of ==`
	}

	_ = err == fs.SkipAll // sentinel error value from the standard library
}
