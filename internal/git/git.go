package git

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/joelvoss/release-lit/internal/semver"
)

const (
	releaseMessage = "chore(release): v%s"
	releaseAuthor  = "release-lit-bot"
	releaseEmail   = "bot@release-lit"
)

type GitOpts struct {
	RootDir string
}

// GetRootDir returns the root directory of the git repository.
// If the current directory is not a git repository, an error is returned.
func GetRoot(opts *GitOpts) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	if opts != nil && opts.RootDir != "" {
		cmd.Dir = opts.RootDir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf(
			"Error getting git root directory. Reason: '%s'\n",
			strings.TrimSpace(string(output)),
		)
		return "", errors.New(msg)
	}

	// NOTE(joel): Remove `/private` prefix on macOS
	if runtime.GOOS == "darwin" {
		output = []byte(strings.ReplaceAll(string(output), "/private", ""))
	}

	return strings.TrimSpace(string(output)), nil
}

////////////////////////////////////////////////////////////////////////////////

// GetTags gets all tags sorted by version in descending order, e.g. v1.0.0,
// v0.1.0, v0.0.1.
func GetTags(opts *GitOpts) ([]*semver.Version, error) {
	cmd := exec.Command("git", "tag", "--sort=-v:refname", "--merged")
	if opts != nil && opts.RootDir != "" {
		cmd.Dir = opts.RootDir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf(
			"Error getting git tags. Reason: '%s'\n",
			strings.TrimSpace(string(output)),
		)
		return nil, errors.New(msg)
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")

	parsedTags := make([]*semver.Version, 0)
	if len(tags) == 0 || (len(tags) == 1 && tags[0] == "") {
		return parsedTags, nil
	}

	for _, tag := range tags {
		parsed, err := semver.Parse(tag)
		if err == nil {
			parsedTags = append(parsedTags, parsed)
		} else {
			fmt.Printf("WARN: Ignoring tag '%s'. Reason: '%s'\n", tag, err)
		}
	}

	return parsedTags, nil
}

////////////////////////////////////////////////////////////////////////////////

// GetTagHead gets the sha1 of the commit that the tag points to.
func GetTagHead(tag string, opts *GitOpts) (string, error) {
	if tag == "" {
		return "", nil
	}

	cmd := exec.Command("git", "rev-list", "-n", "1", tag)
	if opts != nil && opts.RootDir != "" {
		cmd.Dir = opts.RootDir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf(
			"Error getting sha1 for tag '%s'. Reason: '%s'\n",
			tag, strings.TrimSpace(string(output)),
		)
		return "", errors.New(msg)
	}

	return strings.TrimSpace(string(output)), nil
}

////////////////////////////////////////////////////////////////////////////////

// GetCommits gets all commits since the given sha.
func GetCommits(sha string, opts *GitOpts) ([]*Commit, error) {
	commitStr := `{"long": "%H", "short": "%h"}`
	committerStr := `{"name": "%cN", "email": "%cE"}`
	format := fmt.Sprintf(
		`--pretty=format:{"sha": %s, "committer": %s, "date": "%%ci", "subject": "%%s", "body": "%%b"}==DEL==`,
		commitStr, committerStr,
	)

	gitArgs := []string{"log", format}
	if sha != "" {
		sha = sha + ".."
		gitArgs = append(gitArgs, sha)
	}
	cmd := exec.Command("git", gitArgs...)
	if opts != nil && opts.RootDir != "" {
		cmd.Dir = opts.RootDir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf(
			"Error getting commits. Reason: '%s'\n",
			strings.TrimSpace(string(output)),
		)
		return nil, errors.New(msg)
	}

	// NOTE(joel): Split by '},==DEL=={'
	splits := strings.Split(string(output), "==DEL==\n")

	// NOTE(joel): Create commit structs from the splits.
	// If there are no commits, return an empty slice.
	commits := make([]*Commit, 0)
	if len(splits) == 0 || (len(splits) == 1 && splits[0] == "") {
		return commits, nil
	}
	for _, s := range splits {
		s = strings.TrimSpace(s)
		s = strings.TrimSuffix(s, "==DEL==")
		s = strings.ReplaceAll(s, "\n", "\\n")

		c := &Commit{}
		if err := json.Unmarshal([]byte(s), &c); err != nil {
			fmt.Printf("WARN: Could not parse commit. Reason: '%s'\n", err)
			continue
		}
		// NOTE(joel): After unmarshalling, post-process the commit and set type,
		// scope, breaking and subject.
		if err := c.PostProcess(); err != nil {
			fmt.Printf("WARN: Could not post-process commit. Reason: '%s'\n", err)
			continue
		}
		commits = append(commits, c)
	}

	return commits, nil
}

////////////////////////////////////////////////////////////////////////////////

// CreateRelease creates a release commit and tags it with the given version.
func CreateRelease(v *semver.Version, opts *GitOpts) error {
	// Add all files to git
	cmd := exec.Command("git", "add", ".")
	if opts != nil && opts.RootDir != "" {
		cmd.Dir = opts.RootDir
	}
	if err := cmd.Run(); err != nil {
		return errors.New("error adding files to git")
	}

	// Create commit and add any changed files
	commitMsg := fmt.Sprintf(releaseMessage, v.ToString())
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", commitMsg)
	cmd.Env = []string{
		fmt.Sprintf("GIT_AUTHOR_NAME=%s", releaseAuthor),
		fmt.Sprintf("GIT_AUTHOR_EMAIL=%s", releaseEmail),
		fmt.Sprintf("GIT_COMMITTER_NAME=%s", releaseAuthor),
		fmt.Sprintf("GIT_COMMITTER_EMAIL=%s", releaseEmail),
	}
	if opts != nil && opts.RootDir != "" {
		cmd.Dir = opts.RootDir
	}
	if err := cmd.Run(); err != nil {
		return errors.New("error creating release commit")
	}

	// Tag the commit
	versionStr := fmt.Sprintf("v%s", v.ToString())
	cmd = exec.Command("git", "tag", "-a", versionStr, "-m", versionStr)
	cmd.Env = []string{
		fmt.Sprintf("GIT_AUTHOR_NAME=%s", releaseAuthor),
		fmt.Sprintf("GIT_AUTHOR_EMAIL=%s", releaseEmail),
		fmt.Sprintf("GIT_COMMITTER_NAME=%s", releaseAuthor),
		fmt.Sprintf("GIT_COMMITTER_EMAIL=%s", releaseEmail),
	}
	if opts != nil && opts.RootDir != "" {
		cmd.Dir = opts.RootDir
	}
	if err := cmd.Run(); err != nil {
		return errors.New("error creating release tag")
	}

	return nil
}
