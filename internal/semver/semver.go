package semver

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

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
