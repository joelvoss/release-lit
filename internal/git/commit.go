package git

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/joelvoss/release-lit/internal/semver"
)

type CommitType int

const (
	CommitTypeBreaking CommitType = iota
	CommitTypeFeat
	CommitTypeFix
	CommitTypeMisc
)

var commitRegexp *regexp.Regexp

func init() {
	commitRegexp = regexp.MustCompile(`(?ms)^(?<type>\w*)(?:\((?<scope>[\w$.\-*/ ]*)\))?(?<breaking>\!)?:(?<message>.*)`)
}

////////////////////////////////////////////////////////////////////////////////

type Sha struct {
	Long  string `json:"long"`
	Short string `json:"short"`
}

type Committer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Commit struct {
	Sha       Sha       `json:"sha"`
	Committer Committer `json:"committer"`
	Date      string    `json:"date"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Breaking  bool      `json:"breaking,omitempty"`
	Scope     string    `json:"scope,omitempty"`
	Type      string    `json:"type,omitempty"`
	Message   string    `json:"message,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////

// Analyze parses the commit subject and sets the type, scope, breaking and
// subject.
func (c *Commit) PostProcess() error {
	matches := commitRegexp.FindStringSubmatch(c.Subject)
	if matches == nil {
		return fmt.Errorf("error finding string submatch for commit '%s'", c.Sha.Long)
	}

	c.Type = strings.TrimSpace(matches[1])
	c.Scope = strings.TrimSpace(matches[2])
	c.Breaking = matches[3] != "" || strings.Contains(c.Body, "BREAKING CHANGE")
	c.Message = strings.TrimSpace(matches[4])

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// ToString returns the commit as a string.
func (c *Commit) ToString() string {
	raw := c.Subject
	if c.Body != "" {
		raw += "\n\n" + c.Body
	}
	return raw
}

////////////////////////////////////////////////////////////////////////////////

// GroupByType groups commits by type.
func GroupByType(commits []*Commit) map[CommitType][]*Commit {
	grouped := make(map[CommitType][]*Commit)
	for _, c := range commits {
		if c.Breaking {
			grouped[CommitTypeBreaking] = append(grouped[CommitTypeBreaking], c)
			continue
		}
		if c.Type == "feat" {
			grouped[CommitTypeFeat] = append(grouped[CommitTypeFeat], c)
			continue
		}
		if c.Type == "fix" {
			grouped[CommitTypeFix] = append(grouped[CommitTypeFix], c)
			continue
		}
		grouped[CommitTypeMisc] = append(grouped[CommitTypeMisc], c)
	}
	return grouped
}

////////////////////////////////////////////////////////////////////////////////

// GetNextRelease determines the release type based on the commits.
func GetNextReleaseType(commits []*Commit) int {
	releaseType := semver.ReleaseTypeNone

	for _, c := range commits {
		if c.Breaking {
			releaseType = semver.ReleaseTypeMajor
		}
		if c.Type == "feat" && releaseType <= semver.ReleaseTypeMinor {
			releaseType = semver.ReleaseTypeMinor
		}
		if c.Type == "fix" && releaseType <= semver.ReleaseTypePatch {
			releaseType = semver.ReleaseTypePatch
		}
	}

	return releaseType
}
