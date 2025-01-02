package semver

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	num     string = "0123456789"
	allowed string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-" + num
)

const (
	ReleaseTypeNone = iota
	ReleaseTypePatch
	ReleaseTypeMinor
	ReleaseTypeMajor
)

// NOTE(joel): This is not the official regex from the semver spec. It has been
// modified to allow for loose handling where versions like 2.1 are detected.
const semVerRegex string = `v?(0|[1-9]\d*)(?:\.(0|[1-9]\d*))?(?:\.(0|[1-9]\d*))?` +
	`(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
	`(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`

var versionRegex *regexp.Regexp

func init() {
	versionRegex = regexp.MustCompile("^" + semVerRegex + "$")
}

// NOTE(joel): Version represents a single semantic version.
type Version struct {
	major, minor, patch uint64
	pre                 string
	metadata            string
	Original            string
}

////////////////////////////////////////////////////////////////////////////////

// Parse parses a given version string and returns an instance of Version or
// an error if unable to parse the version.
func Parse(v string) (*Version, error) {
	m := versionRegex.FindStringSubmatch(v)
	if m == nil {
		return nil, errors.New("invalid version string")
	}

	sv := &Version{
		metadata: m[5],
		pre:      m[4],
		Original: v,
	}

	var err error
	sv.major, err = strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing version segment: %s", err)
	}

	if m[2] != "" {
		sv.minor, err = strconv.ParseUint(m[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing version segment: %s", err)
		}
	} else {
		sv.minor = 0
	}

	if m[3] != "" {
		sv.patch, err = strconv.ParseUint(m[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing version segment: %s", err)
		}
	} else {
		sv.patch = 0
	}

	if sv.pre != "" {
		if err = validatePrerelease(sv.pre); err != nil {
			return nil, err
		}
	}

	if sv.metadata != "" {
		if err = validateMetadata(sv.metadata); err != nil {
			return nil, err
		}
	}

	return sv, nil
}

// ToString converts a Version object to a string.
// Note, if the original version contained a leading v this version will not.
// See the Original() method to retrieve the original value. Semantic Versions
// don't contain a leading v per the spec. Instead it's optional on
// implementation.
func (v *Version) ToString() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%d.%d.%d", v.major, v.minor, v.patch)
	if v.pre != "" {
		fmt.Fprintf(&buf, "-%s", v.pre)
	}
	if v.metadata != "" {
		fmt.Fprintf(&buf, "+%s", v.metadata)
	}

	return buf.String()
}

////////////////////////////////////////////////////////////////////////////////

// IncPatch produces the next patch version
func (v *Version) IncPatch() {
	// NOTE(joel): According to http://semver.org/#spec-item-9, pre-release
	// versions have a lower precedence than the associated normal version.
	// According to http://semver.org/#spec-item-10, build metadata SHOULD be
	// ignored when determining version precedence.
	if v.pre != "" {
		v.metadata = ""
		v.pre = ""
	} else {
		v.metadata = ""
		v.pre = ""
		v.patch = v.patch + 1
	}
	v.Original = fmt.Sprintf("%s%s", originalVPrefix(v), v.ToString())
}

////////////////////////////////////////////////////////////////////////////////

// IncMinor produces the next minor version
func (v *Version) IncMinor() {
	v.metadata = ""
	v.pre = ""
	v.patch = 0
	v.minor = v.minor + 1
	v.Original = fmt.Sprintf("%s%s", originalVPrefix(v), v.ToString())
}

////////////////////////////////////////////////////////////////////////////////

// IncMajor produces the next major version
func (v *Version) IncMajor() {
	v.metadata = ""
	v.pre = ""
	v.patch = 0
	v.minor = 0
	v.major = v.major + 1
	v.Original = fmt.Sprintf("%s%s", originalVPrefix(v), v.ToString())
}

////////////////////////////////////////////////////////////////////////////////

// validatePrerelease validates a pre-release string.
// From the spec: "Identifiers MUST comprise only ASCII alphanumerics and
// hyphen [0-9A-Za-z-]. Identifiers MUST NOT be empty. Numeric identifiers
// MUST NOT include leading zeroes.".
func validatePrerelease(p string) error {
	eparts := strings.Split(p, ".")
	for _, p := range eparts {
		if p == "" {
			return errors.New("invalid metadata string")
		} else if containsOnly(p, num) {
			if len(p) > 1 && p[0] == '0' {
				return errors.New("version segment starts with 0")
			}
		} else if !containsOnly(p, allowed) {
			return errors.New("invalid pre-release string")
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// validateMetadata validates a metadata string.
// From the spec: "Build metadata MAY be denoted by appending a plus sign and
// a series of dot separated identifiers immediately following the patch or
// pre-release version. Identifiers MUST comprise only ASCII alphanumerics and
// hyphen [0-9A-Za-z-]. Identifiers MUST NOT be empty.".
func validateMetadata(m string) error {
	eparts := strings.Split(m, ".")
	for _, p := range eparts {
		if p == "" {
			return errors.New("invalid metadata string")
		} else if !containsOnly(p, allowed) {
			return errors.New("invalid metadata string")
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Like strings.ContainsAny but does an only instead of any.
func containsOnly(s string, comp string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune(comp, r)
	}) == -1
}

////////////////////////////////////////////////////////////////////////////////

// originalVPrefix returns the original 'v' prefix if any.
func originalVPrefix(v *Version) string {
	if v.Original != "" && (v.Original[:1] == "v" || v.Original[:1] == "V") {
		return v.Original[:1]
	}
	return ""
}

////////////////////////////////////////////////////////////////////////////////

// Bump increments the version based on the release type.
func Bump(v Version, release int) (*Version, error) {
	switch release {
	case ReleaseTypeMajor:
		v.IncMajor()
	case ReleaseTypeMinor:
		v.IncMinor()
	case ReleaseTypePatch:
		v.IncPatch()
	}

	return &v, nil
}
