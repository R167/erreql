package erreql_test

import (
	"testing"

	"github.com/R167/erreql"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, erreql.Analyzer, "./...")
}
