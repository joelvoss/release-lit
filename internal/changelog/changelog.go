package changelog

import (
	"bytes"
	_ "embed"
	"os"
	"text/template"
	"time"

	"github.com/joelvoss/release-lit/internal/git"
	"github.com/joelvoss/release-lit/internal/semver"
)

// NOTE(joel): Declare a package global variable to hold time.Now so that it
// can be mocked/overwritten in tests.
var now = time.Now

//go:embed changelog.tpl
var changelogTemplate string

type ChangelogTpl struct {
	Version string
	Date    string
	Commits map[git.CommitType][]*git.Commit
}

////////////////////////////////////////////////////////////////////////////////

// Generate creates a changelog based on the commits and new version.
func Generate(commits []*git.Commit, newVersion *semver.Version, filepath string) error {
	groupedCommits := git.GroupByType(commits)

	var b bytes.Buffer

	tpl, err := template.New("changelog").Parse(changelogTemplate)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, ChangelogTpl{
		Version: newVersion.ToString(),
		Date:    now().Format("2006-01-02"),
		Commits: groupedCommits,
	}); err != nil {
		return err
	}

	// NOTE(joel): Read the current changelog file (if any)
	if _, err := os.Stat(filepath); err == nil {
		oldChangelog, err := os.ReadFile(filepath)
		if err != nil {
			return err
		}
		// NOTE(joel): Append the old changelog to the new one (and remove the
		// header if it exists)
		idx := bytes.Index(oldChangelog, []byte("# Changelog\n"))
		if idx == -1 {
			idx = 0
		} else {
			idx += len("# Changelog\n")
		}
		_, err = b.Write(oldChangelog[idx:])
		if err != nil {
			return err
		}
	}

	// NOTE(joel): Write the new changelog file (or overwrite the current one)
	if err := os.WriteFile(filepath, b.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}
