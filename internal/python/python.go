package python

import (
	"os"
	"path"
	"regexp"

	"github.com/joelvoss/release-lit/internal/semver"
)

var npmRegexp *regexp.Regexp

func init() {
	npmRegexp = regexp.MustCompile(`(?i)version\s*=\s*".*"`)
}

func UpdateVersion(v *semver.Version, cwd string) error {
	f := path.Join(cwd, "./pyproject.toml")

	if _, err := os.Stat(f); err != nil {
		return err
	}

	content, err := os.ReadFile(f)
	if err != nil {
		return err
	}

	repl := []byte(`version = "` + v.ToString() + `"`)
	content = npmRegexp.ReplaceAll(content, repl)

	if err := os.WriteFile(f, content, 0644); err != nil {
		return err
	}

	return nil
}
