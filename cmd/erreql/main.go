package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/R167/erreql"
)

func main() {
	singlechecker.Main(erreql.Build(erreql.Config{}))
}
