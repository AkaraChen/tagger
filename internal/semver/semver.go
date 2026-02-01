package semver

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// PreReleaseType represents the type of pre-release
type PreReleaseType string

const (
	PreReleaseAlpha PreReleaseType = "alpha"
	PreReleaseBeta  PreReleaseType = "beta"
	PreReleaseRC    PreReleaseType = "rc"
)

// PreReleaseInfo contains parsed pre-release information
type PreReleaseInfo struct {
	Type   PreReleaseType
	Number int
}

// LatestPreReleases contains the latest version for each pre-release type
type LatestPreReleases struct {
	Alpha *semver.Version
	Beta  *semver.Version
	RC    *semver.Version
}

// VersionManager 管理语义化版本
type VersionManager struct{}

// NewVersionManager 创建一个新的 VersionManager
func NewVersionManager() *VersionManager {
	return &VersionManager{}
}

// ParseTags 解析 tags，返回符合 semver 格式的版本列表
func (vm *VersionManager) ParseTags(tags []string) ([]*semver.Version, error) {
	var versions []*semver.Version

	for _, tag := range tags {
		// 移除 v 前缀（如果有）
		versionStr := strings.TrimPrefix(tag, "v")

		// 尝试解析
		v, err := semver.NewVersion(versionStr)
		if err != nil {
			// 跳过不符合 semver 格式的 tag
			continue
		}

		versions = append(versions, v)
	}

	return versions, nil
}

// GetLatestVersion 获取最新版本，如果没有版本则返回 v0.0.0
func (vm *VersionManager) GetLatestVersion(versions []*semver.Version) *semver.Version {
	if len(versions) == 0 {
		// 默认从 v0.0.0 开始
		v, _ := semver.NewVersion("0.0.0")
		return v
	}

	latest := versions[0]
	for _, v := range versions[1:] {
		if v.GreaterThan(latest) {
			latest = v
		}
	}

	return latest
}

// BumpMajor 递增主版本号
func (vm *VersionManager) BumpMajor(v *semver.Version) *semver.Version {
	newVersion := v.IncMajor()
	return &newVersion
}

// BumpMinor 递增次版本号
func (vm *VersionManager) BumpMinor(v *semver.Version) *semver.Version {
	newVersion := v.IncMinor()
	return &newVersion
}

// BumpPatch 递增补丁版本号
func (vm *VersionManager) BumpPatch(v *semver.Version) *semver.Version {
	newVersion := v.IncPatch()
	return &newVersion
}

// FormatVersion 格式化版本号为 vX.Y.Z 格式
func (vm *VersionManager) FormatVersion(v *semver.Version) string {
	return fmt.Sprintf("v%s", v.String())
}

// CalculateNewVersion 根据更新类型计算新版本号
func (vm *VersionManager) CalculateNewVersion(current *semver.Version, bumpType string) (*semver.Version, error) {
	switch strings.ToLower(bumpType) {
	case "major":
		return vm.BumpMajor(current), nil
	case "minor":
		return vm.BumpMinor(current), nil
	case "patch":
		return vm.BumpPatch(current), nil
	default:
		return nil, fmt.Errorf("invalid bump type: %s (must be major, minor, or patch)", bumpType)
	}
}

// IsPreRelease checks if a version has a pre-release suffix
func (vm *VersionManager) IsPreRelease(v *semver.Version) bool {
	return v.Prerelease() != ""
}

// GetPreReleaseInfo parses the pre-release suffix and returns type and number
// Expected format: alpha-1, beta-2, rc-1
func (vm *VersionManager) GetPreReleaseInfo(v *semver.Version) *PreReleaseInfo {
	prerelease := v.Prerelease()
	if prerelease == "" {
		return nil
	}

	// Match patterns like "alpha-1", "beta-2", "rc-1"
	re := regexp.MustCompile(`^(alpha|beta|rc)-(\d+)$`)
	matches := re.FindStringSubmatch(prerelease)
	if matches == nil {
		return nil
	}

	num, _ := strconv.Atoi(matches[2])
	return &PreReleaseInfo{
		Type:   PreReleaseType(matches[1]),
		Number: num,
	}
}

// SetPreRelease creates a new version with the specified pre-release type and number
func (vm *VersionManager) SetPreRelease(v *semver.Version, preType PreReleaseType, num int) (*semver.Version, error) {
	versionStr := fmt.Sprintf("%d.%d.%d-%s-%d", v.Major(), v.Minor(), v.Patch(), preType, num)
	return semver.NewVersion(versionStr)
}

// BumpPreRelease increments the pre-release number (alpha-1 → alpha-2)
func (vm *VersionManager) BumpPreRelease(v *semver.Version) (*semver.Version, error) {
	info := vm.GetPreReleaseInfo(v)
	if info == nil {
		return nil, fmt.Errorf("version %s is not a valid pre-release", v.String())
	}

	return vm.SetPreRelease(v, info.Type, info.Number+1)
}

// NextPreReleaseStage advances the pre-release stage (alpha → beta → rc)
func (vm *VersionManager) NextPreReleaseStage(v *semver.Version) (*semver.Version, error) {
	info := vm.GetPreReleaseInfo(v)
	if info == nil {
		return nil, fmt.Errorf("version %s is not a valid pre-release", v.String())
	}

	var nextType PreReleaseType
	switch info.Type {
	case PreReleaseAlpha:
		nextType = PreReleaseBeta
	case PreReleaseBeta:
		nextType = PreReleaseRC
	case PreReleaseRC:
		return nil, fmt.Errorf("rc is the final pre-release stage, use ReleaseStable instead")
	default:
		return nil, fmt.Errorf("unknown pre-release type: %s", info.Type)
	}

	return vm.SetPreRelease(v, nextType, 1)
}

// ReleaseStable removes the pre-release suffix to create a stable version
func (vm *VersionManager) ReleaseStable(v *semver.Version) (*semver.Version, error) {
	if !vm.IsPreRelease(v) {
		return nil, fmt.Errorf("version %s is not a pre-release", v.String())
	}

	versionStr := fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch())
	return semver.NewVersion(versionStr)
}

// GetLatestPreReleases returns the latest version for each pre-release type for a given base version
func (vm *VersionManager) GetLatestPreReleases(versions []*semver.Version, baseVersion *semver.Version) *LatestPreReleases {
	result := &LatestPreReleases{}

	for _, v := range versions {
		// Check if this version matches the base version (same major.minor.patch)
		if v.Major() != baseVersion.Major() || v.Minor() != baseVersion.Minor() || v.Patch() != baseVersion.Patch() {
			continue
		}

		info := vm.GetPreReleaseInfo(v)
		if info == nil {
			continue
		}

		switch info.Type {
		case PreReleaseAlpha:
			if result.Alpha == nil || v.GreaterThan(result.Alpha) {
				result.Alpha = v
			}
		case PreReleaseBeta:
			if result.Beta == nil || v.GreaterThan(result.Beta) {
				result.Beta = v
			}
		case PreReleaseRC:
			if result.RC == nil || v.GreaterThan(result.RC) {
				result.RC = v
			}
		}
	}

	return result
}

// FindNextPreReleaseNumber finds the next available number for a pre-release type
func (vm *VersionManager) FindNextPreReleaseNumber(versions []*semver.Version, baseVersion *semver.Version, preType PreReleaseType) int {
	maxNum := 0

	for _, v := range versions {
		// Check if this version matches the base version (same major.minor.patch)
		if v.Major() != baseVersion.Major() || v.Minor() != baseVersion.Minor() || v.Patch() != baseVersion.Patch() {
			continue
		}

		info := vm.GetPreReleaseInfo(v)
		if info == nil || info.Type != preType {
			continue
		}

		if info.Number > maxNum {
			maxNum = info.Number
		}
	}

	return maxNum + 1
}

// GetLatestStableVersion returns the latest stable (non-pre-release) version
func (vm *VersionManager) GetLatestStableVersion(versions []*semver.Version) *semver.Version {
	var stableVersions []*semver.Version
	for _, v := range versions {
		if !vm.IsPreRelease(v) {
			stableVersions = append(stableVersions, v)
		}
	}

	if len(stableVersions) == 0 {
		v, _ := semver.NewVersion("0.0.0")
		return v
	}

	return vm.GetLatestVersion(stableVersions)
}

// GetAllPreReleasesForBase returns all pre-release versions for a given base version, sorted
func (vm *VersionManager) GetAllPreReleasesForBase(versions []*semver.Version, baseVersion *semver.Version) []*semver.Version {
	var result []*semver.Version

	for _, v := range versions {
		if v.Major() != baseVersion.Major() || v.Minor() != baseVersion.Minor() || v.Patch() != baseVersion.Patch() {
			continue
		}

		if vm.IsPreRelease(v) {
			result = append(result, v)
		}
	}

	// Sort by version
	sort.Slice(result, func(i, j int) bool {
		return result[i].LessThan(result[j])
	})

	return result
}
