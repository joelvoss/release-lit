package golang

import (
	"os"
	"path"
	"regexp"

	"github.com/joelvoss/release-lit/internal/semver"
)

var npmRegexp *regexp.Regexp

func init() {
	npmRegexp = regexp.MustCompile(`(?i)VERSION=".*"`)
}

func UpdateVersion(v *semver.Version, cwd string) error {
	f := path.Join(cwd, "./Taskfile.sh")

	if _, err := os.Stat(f); err != nil {
		return err
	}

	content, err := os.ReadFile(f)
	if err != nil {
		return err
	}

	repl := []byte(`VERSION="` + v.ToString() + `"`)
	content = npmRegexp.ReplaceAll(content, repl)

	if err := os.WriteFile(f, content, 0644); err != nil {
		return err
	}

	return nil
}
