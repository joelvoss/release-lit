package git

import (
	"errors"
	"testing"

	"github.com/joelvoss/release-lit/internal/semver"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

////////////////////////////////////////////////////////////////////////////////

type ExpectedCommit struct {
	Type     string
	Scope    string
	Breaking bool
	Message  string
}

func TestPostProcess(t *testing.T) {
	tests := []struct {
		name     string
		commit   Commit
		expected ExpectedCommit
		error    error
	}{
		{
			name: "Conventional commit message + scope",
			commit: Commit{
				Subject: "feat(scope): add new feature",
			},
			expected: ExpectedCommit{
				Type:     "feat",
				Scope:    "scope",
				Breaking: false,
				Message:  "add new feature",
			},
			error: nil,
		},
		{
			name: "Conventional commit message",
			commit: Commit{
				Subject: "fix: fix bug",
			},
			expected: ExpectedCommit{
				Type:     "fix",
				Scope:    "",
				Breaking: false,
				Message:  "fix bug",
			},
			error: nil,
		},
		{
			name: "Conventional commit message + body + breaking change",
			commit: Commit{
				Subject: "feat!: add breaking change",
				Body:    "BREAKING CHANGE: this is a breaking change",
			},
			expected: ExpectedCommit{
				Type:     "feat",
				Scope:    "",
				Breaking: true,
				Message:  "add breaking change",
			},
		},
		{
			name: "Conventional commit message + breaking change",
			commit: Commit{
				Subject: "feat!: add breaking change",
			},
			expected: ExpectedCommit{
				Type:     "feat",
				Scope:    "",
				Breaking: true,
				Message:  "add breaking change",
			},
			error: nil,
		},
		{
			name: "Conventional commit message + scope + breaking change (body only)",
			commit: Commit{
				Subject: "feat(scope): add breaking change",
				Body:    "BREAKING CHANGE: this is a breaking change",
			},
			expected: ExpectedCommit{
				Type:     "feat",
				Scope:    "scope",
				Breaking: true,
				Message:  "add breaking change",
			},
			error: nil,
		},
		{
			name: "Invalid conventional commit message",
			commit: Commit{
				Subject: "not a convetional commit message",
			},
			error: errors.New("error finding string submatch for commit ''"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.commit.PostProcess()
			if test.error == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, test.error.Error())
			}

			assert.Equal(t, test.expected.Type, test.commit.Type)
			assert.Equal(t, test.expected.Scope, test.commit.Scope)
			assert.Equal(t, test.expected.Breaking, test.commit.Breaking)
			assert.Equal(t, test.expected.Message, test.commit.Message)
		})
	}
}

////////////////////////////////////////////////////////////////////////////////

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		commit   Commit
		expected string
	}{
		{
			name: "Message + body",
			commit: Commit{
				Subject: "feat: add new feature",
				Body:    "This is the body of the commit",
			},
			expected: "feat: add new feature\n\nThis is the body of the commit",
		},
		{
			name: "Message",
			commit: Commit{
				Subject: "fix: fix bug",
			},
			expected: "fix: fix bug",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := test.commit.ToString()
			assert.Equal(t, test.expected, result)
		})
	}
}

////////////////////////////////////////////////////////////////////////////////

func TestGetNextReleaseType(t *testing.T) {
	tests := []struct {
		name     string
		commits  []*Commit
		expected int
	}{
		{
			name:     "No commits",
			commits:  []*Commit{},
			expected: semver.ReleaseTypeNone,
		},
		{
			name: "Chore commits",
			commits: []*Commit{
				{Type: "chore"},
			},
			expected: semver.ReleaseTypeNone,
		},
		{
			name: "Patch release",
			commits: []*Commit{
				{Type: "fix"},
			},
			expected: semver.ReleaseTypePatch,
		},
		{
			name: "Minor release",
			commits: []*Commit{
				{Type: "feat"},
			},
			expected: semver.ReleaseTypeMinor,
		},
		{
			name: "Minor release #2",
			commits: []*Commit{
				{Type: "feat"},
				{Type: "fix"},
			},
			expected: semver.ReleaseTypeMinor,
		},
		{
			name: "Major release",
			commits: []*Commit{
				{Breaking: true},
			},
			expected: semver.ReleaseTypeMajor,
		},
		{
			name: "Major release #2",
			commits: []*Commit{
				{Type: "fix"},
				{Breaking: true},
			},
			expected: semver.ReleaseTypeMajor,
		},
		{
			name: "Major release #3",
			commits: []*Commit{
				{Type: "fix"},
				{Type: "feat"},
				{Breaking: true},
			},
			expected: semver.ReleaseTypeMajor,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := GetNextReleaseType(test.commits)
			assert.Equal(t, test.expected, got)
		})
	}
}
