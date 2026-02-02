package version

import (
	"fmt"

	"github.com/AkaraChen/tagger/internal/semver"
	sv "github.com/Masterminds/semver/v3"
)

// PreReleaseAction 表示对预发布版本可以执行的操作
type PreReleaseAction string

const (
	PreReleaseActionBump    PreReleaseAction = "bump"
	PreReleaseActionAdvance PreReleaseAction = "advance"
	PreReleaseActionStable  PreReleaseAction = "stable"
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
	Type       string
	NewVersion string
	LatestInfo string
}

// SelectorConfig 选择器配置
type SelectorConfig struct {
	TagTemplate           string
	PreReleaseTagTemplate string
	PreReleaseTypes       []string
}

// DefaultSelectorConfig 默认选择器配置
func DefaultSelectorConfig() SelectorConfig {
	return SelectorConfig{
		TagTemplate:           "v{major}.{minor}.{patch}",
		PreReleaseTagTemplate: "v{major}.{minor}.{patch}-{preReleaseType}{preReleaseNum}",
		PreReleaseTypes:       []string{"alpha", "beta", "rc"},
	}
}

// Selector 封装版本选择的核心业务逻辑
type Selector struct {
	vm     *semver.VersionManager
	config SelectorConfig
}

// NewSelector 创建一个新的版本选择器
func NewSelector() *Selector {
	return &Selector{
		vm:     semver.NewVersionManager(),
		config: DefaultSelectorConfig(),
	}
}

// NewSelectorWithConfig 创建一个带自定义配置的版本选择器
func NewSelectorWithConfig(cfg SelectorConfig) *Selector {
	// 填充默认值
	if cfg.TagTemplate == "" {
		cfg.TagTemplate = "v{major}.{minor}.{patch}"
	}
	if cfg.PreReleaseTagTemplate == "" {
		cfg.PreReleaseTagTemplate = "v{major}.{minor}.{patch}-{preReleaseType}{preReleaseNum}"
	}
	if len(cfg.PreReleaseTypes) == 0 {
		cfg.PreReleaseTypes = []string{"alpha", "beta", "rc"}
	}

	return &Selector{
		vm:     semver.NewVersionManager(),
		config: cfg,
	}
}

// formatVersion 使用模板格式化稳定版本号
func (s *Selector) formatVersion(v *sv.Version) string {
	return s.vm.FormatWithTemplate(v, s.config.TagTemplate)
}

// formatPreRelease 使用模板格式化预发布版本号
func (s *Selector) formatPreRelease(v *sv.Version, preType string, preNum int) string {
	return s.vm.FormatPreReleaseWithTemplate(v, s.config.PreReleaseTagTemplate, preType, preNum)
}

// isLastPreReleaseType 检查是否是最后一个预发布类型
func (s *Selector) isLastPreReleaseType(preType string) bool {
	types := s.config.PreReleaseTypes
	return len(types) > 0 && types[len(types)-1] == preType
}

// getNextPreReleaseType 获取下一个预发布类型
func (s *Selector) getNextPreReleaseType(currentType string) (string, bool) {
	types := s.config.PreReleaseTypes
	for i, t := range types {
		if t == currentType && i < len(types)-1 {
			return types[i+1], true
		}
	}
	return "", false
}

// GetPreReleaseOptions 获取对当前预发布版本可以执行的操作选项
func (s *Selector) GetPreReleaseOptions(
	current *sv.Version,
	versions []*sv.Version,
) ([]PreReleaseOption, error) {
	info := s.vm.GetPreReleaseInfoWithTypes(current, s.config.PreReleaseTypes)
	if info == nil {
		return nil, fmt.Errorf("version %s is not a valid pre-release", current.String())
	}

	currentType := string(info.Type)
	var options []PreReleaseOption

	// 选项1: 增加预发布版本号 (alpha1 -> alpha2)
	newNum := info.Number + 1
	bumpedVersion := s.formatPreRelease(current, currentType, newNum)
	options = append(options, PreReleaseOption{
		Action:     PreReleaseActionBump,
		NewVersion: bumpedVersion,
		Desc:       fmt.Sprintf("increment %s number", currentType),
	})

	// 选项2: 升级到下一个预发布阶段
	if nextType, ok := s.getNextPreReleaseType(currentType); ok {
		advancedVersion := s.formatPreRelease(current, nextType, 1)
		options = append(options, PreReleaseOption{
			Action:     PreReleaseActionAdvance,
			NewVersion: advancedVersion,
			Desc:       fmt.Sprintf("%s → %s", currentType, nextType),
		})
	}

	// 选项3: 发布为稳定版本
	stableVersion := s.formatVersion(current)
	options = append(options, PreReleaseOption{
		Action:     PreReleaseActionStable,
		NewVersion: stableVersion,
		Desc:       "release as stable",
	})

	return options, nil
}

// GetStableBumpOptions 获取稳定版本的升级选项
func (s *Selector) GetStableBumpOptions(current *sv.Version) []BumpOption {
	return []BumpOption{
		{
			Type:       "patch",
			NewVersion: s.formatVersion(s.vm.BumpPatch(current)),
			Desc:       "补丁更新",
		},
		{
			Type:       "minor",
			NewVersion: s.formatVersion(s.vm.BumpMinor(current)),
			Desc:       "小版本更新",
		},
		{
			Type:       "major",
			NewVersion: s.formatVersion(s.vm.BumpMajor(current)),
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
	latestPreReleases := s.vm.GetLatestPreReleasesForTypes(versions, baseVersion, s.config.PreReleaseTypes)

	var options []PreReleaseTypeOption

	for _, preType := range s.config.PreReleaseTypes {
		num := s.vm.FindNextPreReleaseNumberForType(versions, baseVersion, preType)
		newVersion := s.formatPreRelease(baseVersion, preType, num)

		var latestInfo string
		if latest, ok := latestPreReleases[preType]; ok {
			latestInfo = s.formatPreRelease(latest, preType, s.vm.GetPreReleaseInfo(latest).Number)
		}

		options = append(options, PreReleaseTypeOption{
			Type:       preType,
			NewVersion: newVersion,
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
	// 验证预发布类型
	validType := false
	for _, t := range s.config.PreReleaseTypes {
		if t == preReleaseType {
			validType = true
			break
		}
	}
	if !validType {
		return "", fmt.Errorf("invalid pre-release type: %s (must be one of %v)", preReleaseType, s.config.PreReleaseTypes)
	}

	if preReleaseNum <= 0 {
		preReleaseNum = 1
	}

	baseVersion := current
	if !s.vm.IsPreRelease(current) {
		baseVersion = s.vm.BumpPatch(current)
	}

	return s.formatPreRelease(baseVersion, preReleaseType, preReleaseNum), nil
}

// GetVersionManager 返回内部的 VersionManager
func (s *Selector) GetVersionManager() *semver.VersionManager {
	return s.vm
}

// GetConfig 返回选择器配置
func (s *Selector) GetConfig() SelectorConfig {
	return s.config
}
