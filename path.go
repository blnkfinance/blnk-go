package blnkgo

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// endpointPath fills the %s verbs in format with path segments, such as
// resource IDs, after validating and escaping each one. Every caller-supplied
// value that becomes part of a request path must go through it, so an ID like
// "../backup" or "%2e%2e/secret" cannot change which endpoint the client's API
// key is sent to.
func endpointPath(format string, segments ...string) (string, error) {
	args := make([]interface{}, len(segments))
	for i, value := range segments {
		segment, err := pathSegment(value)
		if err != nil {
			return "", err
		}
		args[i] = segment
	}
	return fmt.Sprintf(format, args...), nil
}

// pathSegment returns value escaped for use as a single path segment. It
// rejects values that are empty, are a dot segment, or contain characters
// that would add a segment, start a query or fragment, or carry an encoding
// a server might decode into one of those.
func pathSegment(value string) (string, error) {
	switch value {
	case "":
		return "", errors.New("invalid path segment: value is empty")
	case ".", "..":
		return "", fmt.Errorf("invalid path segment %q", value)
	}
	if strings.ContainsAny(value, "/\\?#%") {
		return "", fmt.Errorf("invalid path segment %q: must not contain / \\ ? # or %%", value)
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return "", fmt.Errorf("invalid path segment %q: must not contain control characters", value)
		}
	}
	return url.PathEscape(value), nil
}

// hasDotSegment reports whether a decoded URL path contains a "." or ".."
// segment.
func hasDotSegment(p string) bool {
	for _, segment := range strings.Split(p, "/") {
		if segment == "." || segment == ".." {
			return true
		}
	}
	return false
}
