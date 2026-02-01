package version

import (
	"fmt"

	"github.com/AkaraChen/tagger/internal/semver"
	sv "github.com/Masterminds/semver/v3"
)

// PreReleaseAction 表示对预发布版本可以执行的操作
type PreReleaseAction string

const (
	PreReleaseActionBump   PreReleaseAction = "bump"
	PreReleaseActionAdvance PreReleaseAction = "advance"
	PreReleaseActionStable PreReleaseAction = "stable"
)

// PreReleaseOption 表示一个可用的预发布版本操作选项
type PreReleaseOption struct {
	Action     PreReleaseAction
	NewVersion string
	Desc       string
}

// BumpOption 表示一个版本升级选项
type BumpOption struct {
	Type       string
	NewVersion string
	Desc       string
}

// PreReleaseTypeOption 表示一个预发布类型选项
type PreReleaseTypeOption struct {
	Type       semver.PreReleaseType
	NewVersion string
	LatestInfo string
}

// Selector 封装版本选择的核心业务逻辑
type Selector struct {
	vm *semver.VersionManager
}

// NewSelector 创建一个新的版本选择器
func NewSelector() *Selector {
	return &Selector{
		vm: semver.NewVersionManager(),
	}
}

// GetPreReleaseOptions 获取对当前预发布版本可以执行的操作选项
func (s *Selector) GetPreReleaseOptions(
	current *sv.Version,
	versions []*sv.Version,
) ([]PreReleaseOption, error) {
	info := s.vm.GetPreReleaseInfo(current)
	if info == nil {
		return nil, fmt.Errorf("version %s is not a valid pre-release", current.String())
	}

	var options []PreReleaseOption

	// 选项1: 增加预发布版本号 (alpha-1 -> alpha-2)
	if bumpedVersion, err := s.vm.BumpPreRelease(current); err == nil {
		options = append(options, PreReleaseOption{
			Action:     PreReleaseActionBump,
			NewVersion: s.vm.FormatVersion(bumpedVersion),
			Desc:       fmt.Sprintf("increment %s number", info.Type),
		})
	}

	// 选项2: 升级到下一个预发布阶段 (alpha -> beta -> rc)
	if info.Type != semver.PreReleaseRC {
		if advancedVersion, err := s.vm.NextPreReleaseStage(current); err == nil {
			advancedInfo := s.vm.GetPreReleaseInfo(advancedVersion)
			options = append(options, PreReleaseOption{
				Action:     PreReleaseActionAdvance,
				NewVersion: s.vm.FormatVersion(advancedVersion),
				Desc:       fmt.Sprintf("%s → %s", info.Type, advancedInfo.Type),
			})
		}
	}

	// 选项3: 发布为稳定版本
	if stableVersion, err := s.vm.ReleaseStable(current); err == nil {
		options = append(options, PreReleaseOption{
			Action:     PreReleaseActionStable,
			NewVersion: s.vm.FormatVersion(stableVersion),
			Desc:       "release as stable",
		})
	}

	return options, nil
}

// GetStableBumpOptions 获取稳定版本的升级选项
func (s *Selector) GetStableBumpOptions(current *sv.Version) []BumpOption {
	return []BumpOption{
		{
			Type:       "patch",
			NewVersion: s.vm.FormatVersion(s.vm.BumpPatch(current)),
			Desc:       "补丁更新",
		},
		{
			Type:       "minor",
			NewVersion: s.vm.FormatVersion(s.vm.BumpMinor(current)),
			Desc:       "小版本更新",
		},
		{
			Type:       "major",
			NewVersion: s.vm.FormatVersion(s.vm.BumpMajor(current)),
			Desc:       "大版本更新",
		},
	}
}

// CalculateBumpedVersion 根据 bump 类型计算新版本
func (s *Selector) CalculateBumpedVersion(current *sv.Version, bumpType string) (*sv.Version, error) {
	return s.vm.CalculateNewVersion(current, bumpType)
}

// GetPreReleaseTypeOptions 获取所有可用的预发布类型选项
func (s *Selector) GetPreReleaseTypeOptions(
	baseVersion *sv.Version,
	versions []*sv.Version,
) []PreReleaseTypeOption {
	latestPreReleases := s.vm.GetLatestPreReleases(versions, baseVersion)

	var options []PreReleaseTypeOption

	for _, preType := range []semver.PreReleaseType{
		semver.PreReleaseAlpha,
		semver.PreReleaseBeta,
		semver.PreReleaseRC,
	} {
		num := s.vm.FindNextPreReleaseNumber(versions, baseVersion, preType)
		version, _ := s.vm.SetPreRelease(baseVersion, preType, num)

		var latestInfo string
		switch preType {
		case semver.PreReleaseAlpha:
			if latestPreReleases.Alpha != nil {
				latestInfo = "v" + latestPreReleases.Alpha.String()
			}
		case semver.PreReleaseBeta:
			if latestPreReleases.Beta != nil {
				latestInfo = "v" + latestPreReleases.Beta.String()
			}
		case semver.PreReleaseRC:
			if latestPreReleases.RC != nil {
				latestInfo = "v" + latestPreReleases.RC.String()
			}
		}

		options = append(options, PreReleaseTypeOption{
			Type:       preType,
			NewVersion: s.vm.FormatVersion(version),
			LatestInfo: latestInfo,
		})
	}

	return options
}

// CalculateNonInteractiveVersion 计算非交互式模式下的新版本
func (s *Selector) CalculateNonInteractiveVersion(
	current *sv.Version,
	versions []*sv.Version,
	preReleaseType string,
	preReleaseNum int,
) (string, error) {
	var preType semver.PreReleaseType
	switch preReleaseType {
	case "alpha":
		preType = semver.PreReleaseAlpha
	case "beta":
		preType = semver.PreReleaseBeta
	case "rc":
		preType = semver.PreReleaseRC
	default:
		return "", fmt.Errorf("invalid pre-release type: %s (must be alpha, beta, or rc)", preReleaseType)
	}

	if preReleaseNum <= 0 {
		preReleaseNum = 1
	}

	baseVersion := current
	if !s.vm.IsPreRelease(current) {
		baseVersion = s.vm.BumpPatch(current)
	}

	newVersion, err := s.vm.SetPreRelease(baseVersion, preType, preReleaseNum)
	if err != nil {
		return "", fmt.Errorf("failed to create pre-release version: %w", err)
	}

	return s.vm.FormatVersion(newVersion), nil
}

// GetVersionManager 返回内部的 VersionManager（用于需要直接访问的复杂场景）
func (s *Selector) GetVersionManager() *semver.VersionManager {
	return s.vm
}
