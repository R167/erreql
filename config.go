package erreql

import (
	"regexp"
	"strings"
)

type Config struct {
	// Skip warning on switch statements in addition to equality checks.
	//
	// Default: false
	SkipSwitches bool

	// Regexps for packages that use sentinel error values to indicate success.
	// Inlined as fmt.Sprintf(`^(%s)$`, strings.Join(SkipPackages, "|"))
	//
	// Default: []string{"errors", "errors_test"}
	SkipPackages []string
}

type config struct {
	Config

	packageChecker *regexp.Regexp
}

func (c Config) compileConfig() config {
	if len(c.SkipPackages) == 0 {
		c.SkipPackages = []string{"errors", "errors_test"}
	}

	return config{
		Config:         c,
		packageChecker: regexp.MustCompile(`^(?:` + strings.Join(c.SkipPackages, "|") + `)$`),
	}
}

func (c config) skipPackage(pkg string) bool {
	return c.packageChecker.MatchString(pkg)
}
