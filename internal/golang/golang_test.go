package golang

import (
	"os"
	"path"
	"testing"

	"github.com/joelvoss/release-lit/internal/semver"

	"github.com/stretchr/testify/assert"
)

func prepareDir(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "release-me-repo-*")
	if err != nil {
		t.Fatalf("Error creating temp directory: %v", err)
	}

	cleanUpFunc := func() {
		os.RemoveAll(tempDir)
	}

	taskfileSh := []byte(`#!/bin/bash

set -e

NAME="release-lit"
VERSION="0.1.0"`)
	err = os.WriteFile(tempDir+"/Taskfile.sh", taskfileSh, 0644)
	if err != nil {
		t.Fatalf("Error preparing version file: %v", err)
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

func TestUpdateVersion(t *testing.T) {
	tempDir, cleanUp := prepareDir(t)
	defer cleanUp()

	v, _ := semver.Parse("1.1.0")
	err := UpdateVersion(v, tempDir)

	assert.NoError(t, err)
	assertFileContent(t, path.Join(tempDir, "Taskfile.sh"), `#!/bin/bash

set -e

NAME="release-lit"
VERSION="1.1.0"`)
}
