package cmd

import (
	"fmt"

	"github.com/AkaraChen/tagger/internal/config"
	"github.com/AkaraChen/tagger/internal/semver"
	"github.com/AkaraChen/tagger/internal/ui"
	"github.com/AkaraChen/tagger/internal/version"
	sv "github.com/Masterminds/semver/v3"
)

// tagContext 包含 tag 操作的上下文
type tagContext struct {
	versionMgr     *semver.VersionManager
	selector       *version.Selector
	versions       []*sv.Version
	currentVersion *sv.Version
	opts           TagOptions
	cfg            *config.Config
}

func (ctx *tagContext) currentVersionStr() string {
	return ctx.versionMgr.FormatWithTemplate(ctx.currentVersion, ctx.cfg.GetTagTemplate())
}

func (ctx *tagContext) displayVersionStatus() {
	fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("Current Version: %s", ctx.currentVersionStr())))

	nextPatch := ctx.versionMgr.BumpPatch(ctx.currentVersion)
	preReleaseTypes := ctx.cfg.GetPreReleaseTypes()
	latestPreReleases := ctx.versionMgr.GetLatestPreReleasesForTypes(ctx.versions, nextPatch, preReleaseTypes)

	if len(latestPreReleases) > 0 {
		fmt.Println(ui.InfoStyle.Render("Latest pre-releases:"))
		for _, preType := range preReleaseTypes {
			if v, ok := latestPreReleases[preType]; ok {
				info := ctx.versionMgr.GetPreReleaseInfo(v)
				formatted := ctx.versionMgr.FormatPreReleaseWithTemplate(v, ctx.cfg.GetPreReleaseTagTemplate(), preType, info.Number)
				fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  %s: %s", preType, formatted)))
			} else {
				fmt.Println(ui.HelpStyle.Render(fmt.Sprintf("  %s: (none)", preType)))
			}
		}
		fmt.Println()
	}
}

func (ctx *tagContext) handleExistingPreRelease() (string, error) {
	if ctx.opts.PreReleaseType != "" {
		return ctx.handleNonInteractivePreRelease()
	}

	options, err := ctx.selector.GetPreReleaseOptions(ctx.currentVersion, ctx.versions)
	if err != nil {
		return "", fmt.Errorf("failed to get pre-release options: %w", err)
	}

	// Convert internal/version options to ui options
	uiOptions := make([]ui.PreReleaseActionOption, len(options))
	for i, opt := range options {
		uiOptions[i] = ui.PreReleaseActionOption{
			Action:     ui.PreReleaseAction(opt.Action),
			NewVersion: opt.NewVersion,
			Desc:       opt.Desc,
		}
	}

	action, err := ui.SelectPreReleaseAction(ctx.currentVersionStr(), uiOptions)
	if err != nil {
		return "", err
	}

	for _, opt := range options {
		if opt.Action == version.PreReleaseAction(action) {
			return opt.NewVersion, nil
		}
	}

	return "", fmt.Errorf("invalid action selected")
}

func (ctx *tagContext) handleStableVersion() (string, error) {
	if ctx.opts.PreReleaseType != "" {
		return ctx.handleNonInteractivePreRelease()
	}

	options := ctx.selector.GetStableBumpOptions(ctx.currentVersion)

	// Convert internal/version options to ui options
	uiOptions := make([]ui.BumpTypeOption, len(options))
	for i, opt := range options {
		uiOptions[i] = ui.BumpTypeOption{
			Type:       opt.Type,
			NewVersion: opt.NewVersion,
			Desc:       opt.Desc,
		}
	}

	bumpType, err := ui.SelectBumpType(ctx.currentVersionStr(), uiOptions)
	if err != nil {
		return "", err
	}

	newVersion, err := ctx.selector.CalculateBumpedVersion(ctx.currentVersion, bumpType)
	if err != nil {
		return "", fmt.Errorf("failed to calculate new version: %w", err)
	}

	isPreRelease, err := ui.ConfirmPreRelease()
	if err != nil {
		return "", err
	}

	if !isPreRelease {
		return ctx.versionMgr.FormatWithTemplate(newVersion, ctx.cfg.GetTagTemplate()), nil
	}

	return ctx.selectPreReleaseType(newVersion)
}

func (ctx *tagContext) selectPreReleaseType(baseVersion *sv.Version) (string, error) {
	options := ctx.selector.GetPreReleaseTypeOptions(baseVersion, ctx.versions)

	// Convert internal/version options to ui options
	uiOptions := make([]ui.PreReleaseTypeOption, len(options))
	for i, opt := range options {
		uiOptions[i] = ui.PreReleaseTypeOption{
			Type:       opt.Type,
			NewVersion: opt.NewVersion,
			LatestInfo: opt.LatestInfo,
		}
	}

	selectedType, err := ui.SelectPreReleaseType(ctx.currentVersionStr(), uiOptions)
	if err != nil {
		return "", err
	}

	for _, opt := range options {
		if opt.Type == selectedType {
			return opt.NewVersion, nil
		}
	}

	return "", fmt.Errorf("invalid pre-release type selected")
}

func (ctx *tagContext) handleNonInteractivePreRelease() (string, error) {
	return ctx.selector.CalculateNonInteractiveVersion(
		ctx.currentVersion,
		ctx.versions,
		ctx.opts.PreReleaseType,
		ctx.opts.PreReleaseNum,
	)
}
