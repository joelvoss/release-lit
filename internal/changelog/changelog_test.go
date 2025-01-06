package changelog

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/joelvoss/release-lit/internal/git"
	"github.com/joelvoss/release-lit/internal/semver"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepareChangelog(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "git-repo-*")
	if err != nil {
		t.Fatalf("Error creating temp directory: %v", err)
	}

	cleanUpFunc := func() {
		os.RemoveAll(tempDir)
	}

	changelogStr := "# Changelog\n\n## 0.0.0 - 2006-01-02\n\nSome old content\n"
	err = os.WriteFile(tempDir+"/CHANGELOG.md", []byte(changelogStr), 0644)
	if err != nil {
		t.Fatalf("Error preparing changelog: %v", err)
	}

	return tempDir, cleanUpFunc
}

func assertFileContent(t *testing.T, path string, expected string) {
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	assert.Equal(t, content, []byte(expected))
}

////////////////////////////////////////////////////////////////////////////////

func TestGenerate(t *testing.T) {
	tempDir, cleanUp := prepareChangelog(t)
	defer cleanUp()

	// NOTE(joel): Mock time
	now = func() time.Time {
		return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	}

	commits := []*git.Commit{
		{
			Sha: git.Sha{
				Short: "1234567",
				Long:  "1234567891234567891234567891234567891234",
			},
			Message:  "some breaking change",
			Body:     "BREAKING CHANGE: some breaking change\nWith a linebreak",
			Breaking: true,
			Type:     "feat",
			Scope:    "",
		},
		{
			Sha: git.Sha{
				Short: "2234567",
				Long:  "2234567891234567891234567891234567891234",
			},
			Message:  "some feature",
			Breaking: false,
			Type:     "feat",
			Scope:    "",
		},
		{
			Sha: git.Sha{
				Short: "4234567",
				Long:  "4234567891234567891234567891234567891234",
			},
			Message:  "some fix",
			Breaking: false,
			Type:     "fix",
			Scope:    "",
		},
		{
			Sha: git.Sha{
				Short: "3234567",
				Long:  "3234567891234567891234567891234567891234",
			},
			Message:  "some feature",
			Breaking: false,
			Type:     "feat",
			Scope:    "scope",
		},
		{
			Sha: git.Sha{
				Short: "5234567",
				Long:  "5234567891234567891234567891234567891234",
			},
			Message:  "some chore",
			Breaking: false,
			Type:     "chore",
			Scope:    "",
		},
		{
			Sha: git.Sha{
				Short: "6234567",
				Long:  "6234567891234567891234567891234567891234",
			},
			Message:  "some change of documentation",
			Breaking: false,
			Type:     "docs",
			Scope:    "",
		},
	}
	newVersion, _ := semver.Parse("1.0.0")

	filepath := path.Join(tempDir, "./CHANGELOG.md")
	err := Generate(commits, newVersion, filepath)

	require.NoError(t, err)
	require.FileExists(t, filepath)
	assertFileContent(t, filepath, `# Changelog

## 1.0.0 - 2006-01-02

### BREAKING CHANGES
- feat: some breaking change (1234567)

### Features
- some feature (2234567)
- **scope:** some feature (3234567)

### Bug Fixes
- some fix (4234567)

### Miscellaneous
- chore: some chore (5234567)
- docs: some change of documentation (6234567)

## 0.0.0 - 2006-01-02

Some old content
`)
}

func TestGenerate2(t *testing.T) {
	tempDir, cleanUp := prepareChangelog(t)
	defer cleanUp()

	// NOTE(joel): Mock time
	now = func() time.Time {
		return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	}

	commits := make([]*git.Commit, 0)
	newVersion, _ := semver.Parse("1.0.0")

	filepath := path.Join(tempDir, "./CHANGELOG.md")
	err := Generate(commits, newVersion, filepath)

	require.NoError(t, err)
	require.FileExists(t, filepath)
	assertFileContent(t, filepath, `# Changelog

## 1.0.0 - 2006-01-02

- No changes

## 0.0.0 - 2006-01-02

Some old content
`)
}
