package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.0.0", "1.0.0"},
		{"v1.0.0", "1.0.0"},
		{"1.0.0-alpha", "1.0.0-alpha"},
		{"1.0.0+build", "1.0.0+build"},
		{"1.0.0-alpha+build", "1.0.0-alpha+build"},
		{"2.1", "2.1.0"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()

			v, err := Parse(test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, v.ToString())
		})
	}
}

func TestIncPatch(t *testing.T) {
	v, _ := Parse("1.0.0")
	v.IncPatch()
	assert.Equal(t, "1.0.1", v.ToString())
}

func TestIncMinor(t *testing.T) {
	v, _ := Parse("1.0.0")
	v.IncMinor()
	assert.Equal(t, "1.1.0", v.ToString())
}

func TestIncMajor(t *testing.T) {
	v, _ := Parse("1.0.0")
	v.IncMajor()
	assert.Equal(t, "2.0.0", v.ToString())
}

func TestBump(t *testing.T) {
	v, _ := Parse("1.0.0")

	v1, _ := Bump(*v, ReleaseTypePatch)
	assert.Equal(t, "1.0.1", v1.ToString())
	assert.Equal(t, "1.0.0", v.ToString())

	v2, _ := Bump(*v, ReleaseTypeMinor)
	assert.Equal(t, "1.1.0", v2.ToString())

	v3, _ := Bump(*v, ReleaseTypeMajor)
	assert.Equal(t, "2.0.0", v3.ToString())
}
