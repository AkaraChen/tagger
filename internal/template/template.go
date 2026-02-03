package template

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Default templates matching current behavior
const (
	DefaultReleaseTemplate    = "v{major}.{minor}.{patch}"
	DefaultPreReleaseTemplate = "v{major}.{minor}.{patch}-{preType}-{preNum}"
)

// Placeholder patterns for template conversion
var placeholderPatterns = map[string]string{
	"{major}":   `(?P<major>\d+)`,
	"{minor}":   `(?P<minor>\d+)`,
	"{patch}":   `(?P<patch>\d+)`,
	"{preType}": `(?P<preType>[a-zA-Z]+)`,
	"{preNum}":  `(?P<preNum>\d+)`,
}

// Required placeholders for each template type
var (
	requiredReleasePlaceholders    = []string{"{major}", "{minor}", "{patch}"}
	requiredPreReleasePlaceholders = []string{"{major}", "{minor}", "{patch}", "{preType}", "{preNum}"}
)

// ParsedTag contains the parsed components from a tag string
type ParsedTag struct {
	Major   uint64
	Minor   uint64
	Patch   uint64
	PreType string // empty for release tags
	PreNum  int    // 0 for release tags
}

// IsPreRelease returns true if the parsed tag has pre-release information
func (p *ParsedTag) IsPreRelease() bool {
	return p.PreType != ""
}

// TagTemplate handles parsing and formatting of version tags
type TagTemplate struct {
	releaseTemplate    string
	preReleaseTemplate string
	releaseRegex       *regexp.Regexp
	preReleaseRegex    *regexp.Regexp
}

// NewTagTemplate creates a new TagTemplate with the given templates
// If empty strings are provided, defaults are used
func NewTagTemplate(release, preRelease string) (*TagTemplate, error) {
	if release == "" {
		release = DefaultReleaseTemplate
	}
	if preRelease == "" {
		preRelease = DefaultPreReleaseTemplate
	}

	// Validate templates
	if err := validateTemplate(release, requiredReleasePlaceholders); err != nil {
		return nil, fmt.Errorf("invalid release template: %w", err)
	}
	if err := validateTemplate(preRelease, requiredPreReleasePlaceholders); err != nil {
		return nil, fmt.Errorf("invalid pre-release template: %w", err)
	}

	// Build regex patterns
	releaseRegex, err := buildParseRegex(release)
	if err != nil {
		return nil, fmt.Errorf("failed to build release regex: %w", err)
	}

	preReleaseRegex, err := buildParseRegex(preRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to build pre-release regex: %w", err)
	}

	return &TagTemplate{
		releaseTemplate:    release,
		preReleaseTemplate: preRelease,
		releaseRegex:       releaseRegex,
		preReleaseRegex:    preReleaseRegex,
	}, nil
}

// NewDefaultTagTemplate creates a TagTemplate with default templates
func NewDefaultTagTemplate() *TagTemplate {
	t, _ := NewTagTemplate("", "")
	return t
}

// validateTemplate checks if the template contains all required placeholders
func validateTemplate(template string, required []string) error {
	for _, placeholder := range required {
		if !strings.Contains(template, placeholder) {
			return fmt.Errorf("missing required placeholder: %s", placeholder)
		}
	}
	return nil
}

// buildParseRegex converts a template string to a regex pattern
func buildParseRegex(template string) (*regexp.Regexp, error) {
	// Escape regex special characters
	pattern := regexp.QuoteMeta(template)

	// Replace placeholders with named capture groups
	for placeholder, captureGroup := range placeholderPatterns {
		escaped := regexp.QuoteMeta(placeholder)
		pattern = strings.ReplaceAll(pattern, escaped, captureGroup)
	}

	// Anchor the pattern
	pattern = "^" + pattern + "$"

	return regexp.Compile(pattern)
}

// Parse attempts to parse a tag string and extract version components
// It tries pre-release pattern first (more specific), then release pattern
func (t *TagTemplate) Parse(tag string) (*ParsedTag, error) {
	// Try pre-release pattern first (more specific)
	if parsed, err := t.parseWithRegex(t.preReleaseRegex, tag); err == nil {
		return parsed, nil
	}

	// Try release pattern
	if parsed, err := t.parseWithRegex(t.releaseRegex, tag); err == nil {
		return parsed, nil
	}

	return nil, fmt.Errorf("tag %q does not match configured template", tag)
}

// parseWithRegex extracts components using the given regex
func (t *TagTemplate) parseWithRegex(re *regexp.Regexp, tag string) (*ParsedTag, error) {
	matches := re.FindStringSubmatch(tag)
	if matches == nil {
		return nil, fmt.Errorf("no match")
	}

	result := &ParsedTag{}

	// Extract named groups
	for i, name := range re.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}

		value := matches[i]
		switch name {
		case "major":
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid major version: %w", err)
			}
			result.Major = v
		case "minor":
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid minor version: %w", err)
			}
			result.Minor = v
		case "patch":
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid patch version: %w", err)
			}
			result.Patch = v
		case "preType":
			result.PreType = value
		case "preNum":
			v, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid pre-release number: %w", err)
			}
			result.PreNum = v
		}
	}

	return result, nil
}

// Format generates a tag string from version components
func (t *TagTemplate) Format(major, minor, patch uint64) string {
	result := t.releaseTemplate
	result = strings.ReplaceAll(result, "{major}", strconv.FormatUint(major, 10))
	result = strings.ReplaceAll(result, "{minor}", strconv.FormatUint(minor, 10))
	result = strings.ReplaceAll(result, "{patch}", strconv.FormatUint(patch, 10))
	return result
}

// FormatPreRelease generates a pre-release tag string from version components
func (t *TagTemplate) FormatPreRelease(major, minor, patch uint64, preType string, preNum int) string {
	result := t.preReleaseTemplate
	result = strings.ReplaceAll(result, "{major}", strconv.FormatUint(major, 10))
	result = strings.ReplaceAll(result, "{minor}", strconv.FormatUint(minor, 10))
	result = strings.ReplaceAll(result, "{patch}", strconv.FormatUint(patch, 10))
	result = strings.ReplaceAll(result, "{preType}", preType)
	result = strings.ReplaceAll(result, "{preNum}", strconv.Itoa(preNum))
	return result
}

// ReleaseTemplate returns the release template string
func (t *TagTemplate) ReleaseTemplate() string {
	return t.releaseTemplate
}

// PreReleaseTemplate returns the pre-release template string
func (t *TagTemplate) PreReleaseTemplate() string {
	return t.preReleaseTemplate
}

// CanParse checks if a tag matches either template pattern
func (t *TagTemplate) CanParse(tag string) bool {
	_, err := t.Parse(tag)
	return err == nil
}
