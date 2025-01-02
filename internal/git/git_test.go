package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/joelvoss/release-lit/internal/semver"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createGitRepo(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "git-repo-*")
	if err != nil {
		t.Fatalf("Error creating temp directory: %v", err)
	}

	cleanUpFunc := func() {
		os.RemoveAll(tempDir)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error initializing git repository: %v", err)
	}

	// Update git config to set user name and email
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error setting user name: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test.user@example.com")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error setting user email: %v", err)
	}

	return tempDir, cleanUpFunc
}

type TestCommit struct {
	Msg string
	Tag string
}

func createCommits(t *testing.T, repoPath string, commits []TestCommit) []string {
	var commitShas []string

	for i, commit := range commits {
		cmd := exec.Command("git", "commit", "--allow-empty", "-m", commit.Msg)
		cmd.Env = []string{
			"GIT_AUTHOR_DATE=2024-01-01T12:00:00-01:00",
			"GIT_COMMITTER_DATE=2024-01-01T12:00:00-01:00",
		}
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Error creating commit %d: %v", i, err)
		}

		cmd = exec.Command("git", "rev-parse", "HEAD")
		cmd.Dir = repoPath
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Error getting commit SHA %d: %v", i, err)
		}
		commitShas = append(commitShas, strings.TrimSpace(string(output)))

		if commit.Tag != "" {
			cmd = exec.Command("git", "tag", commit.Tag)
			cmd.Dir = repoPath
			if err := cmd.Run(); err != nil {
				t.Fatalf("Error creating tag %d: %v", i, err)
			}
		}
	}

	return commitShas
}

////////////////////////////////////////////////////////////////////////////////

func TestGetRoot(t *testing.T) {
	cwd, cleanUp := createGitRepo(t)
	defer cleanUp()

	fmt.Println(cwd)

	root, err := GetRoot(&GitOpts{RootDir: cwd})
	assert.NoError(t, err)
	assert.Equal(t, cwd, root)
}

////////////////////////////////////////////////////////////////////////////////

func TestGetTags(t *testing.T) {
	cwd, cleanUp := createGitRepo(t)
	defer cleanUp()

	commits := []TestCommit{
		{Msg: "Commit #1", Tag: "v0.0.1"},
		{Msg: "Commit #2", Tag: "v0.1.0"},
		{Msg: "Commit #3", Tag: "v1.0.0"},
	}
	createCommits(t, cwd, commits)

	tags, err := GetTags(&GitOpts{RootDir: cwd})
	assert.NoError(t, err)
	assert.Len(t, tags, 3)
	assert.Equal(t, "1.0.0", tags[0].ToString())
	assert.Equal(t, "0.1.0", tags[1].ToString())
	assert.Equal(t, "0.0.1", tags[2].ToString())
	assert.Equal(t, "v1.0.0", tags[0].Original)
	assert.Equal(t, "v0.1.0", tags[1].Original)
	assert.Equal(t, "v0.0.1", tags[2].Original)
}

////////////////////////////////////////////////////////////////////////////////

func TestGetTagHead(t *testing.T) {
	cwd, cleanUp := createGitRepo(t)
	defer cleanUp()

	commits := []TestCommit{
		{Msg: "Commit #1", Tag: "v0.0.1"},
		{Msg: "Commit #2", Tag: "v0.1.0"},
		{Msg: "Commit #3", Tag: "v1.0.0"},
	}
	expectedShas := createCommits(t, cwd, commits)

	sha, err := GetTagHead("v0.1.0", &GitOpts{RootDir: cwd})
	assert.NoError(t, err)
	assert.NotEmpty(t, sha)
	assert.Equal(t, expectedShas[1], sha)
}

////////////////////////////////////////////////////////////////////////////////

func TestGetCommits(t *testing.T) {
	cwd, cleanUp := createGitRepo(t)
	defer cleanUp()

	tCommits := []TestCommit{
		{Msg: "feat: commit #1", Tag: "v0.0.1"},
		{Msg: "feat: commit #2", Tag: "v0.1.0"},
		{Msg: "feat: commit #3", Tag: "v1.0.0"},
		{Msg: "feat: commit #4", Tag: "v1.1.0"},
	}
	expectedShas := createCommits(t, cwd, tCommits)

	commits, err := GetCommits(expectedShas[1], &GitOpts{RootDir: cwd})

	assert.NoError(t, err)
	assert.Len(t, commits, 2)
	assert.Equal(t, commits, []*Commit{
		{
			Sha: Sha{
				Short: expectedShas[3][:7],
				Long:  expectedShas[3],
			},
			Committer: Committer{
				Name:  "Test User",
				Email: "test.user@example.com",
			},
			Date:     "2024-01-01 12:00:00 -0100",
			Subject:  tCommits[3].Msg,
			Type:     "feat",
			Scope:    "",
			Breaking: false,
			Message:  "commit #4",
		},
		{
			Sha: Sha{
				Short: expectedShas[2][:7],
				Long:  expectedShas[2],
			},
			Committer: Committer{
				Name:  "Test User",
				Email: "test.user@example.com",
			},
			Date:     "2024-01-01 12:00:00 -0100",
			Subject:  tCommits[2].Msg,
			Type:     "feat",
			Scope:    "",
			Breaking: false,
			Message:  "commit #3",
		},
	})
}

////////////////////////////////////////////////////////////////////////////////

func TestGetCommitsNoCommits(t *testing.T) {
	cwd, cleanUp := createGitRepo(t)
	defer cleanUp()

	commits, err := GetCommits("", &GitOpts{RootDir: cwd})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Error getting commits")
	}
	assert.Nil(t, commits)
}

////////////////////////////////////////////////////////////////////////////////

func TestCreateRelease(t *testing.T) {
	cwd, cleanUp := createGitRepo(t)
	defer cleanUp()

	v, _ := semver.Parse("v2.0.0")
	err := CreateRelease(v, &GitOpts{RootDir: cwd})

	require.NoError(t, err)

	// Assert release tag
	cmd := exec.Command("git", "tag")
	cmd.Dir = cwd
	output, err := cmd.Output()
	require.NoError(t, err)
	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	assert.Contains(t, tags, "v2.0.0")

	// Assert release commit
	cmd = exec.Command("git", "show", "v2.0.0")
	cmd.Dir = cwd
	output, err = cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(output), "chore(release): v2.0.0")
}
