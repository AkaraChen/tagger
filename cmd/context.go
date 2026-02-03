package cmd

import (
	"fmt"

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
}

func (ctx *tagContext) currentVersionStr() string {
	return ctx.versionMgr.FormatVersion(ctx.currentVersion)
}

func (ctx *tagContext) displayVersionStatus() {
	fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("Current Version: %s", ctx.currentVersionStr())))

	nextPatch := ctx.versionMgr.BumpPatch(ctx.currentVersion)
	latestPreReleases := ctx.versionMgr.GetLatestPreReleases(ctx.versions, nextPatch)

	hasPreReleases := latestPreReleases.Alpha != nil || latestPreReleases.Beta != nil || latestPreReleases.RC != nil
	if hasPreReleases {
		fmt.Println(ui.InfoStyle.Render("Latest pre-releases:"))
		if latestPreReleases.Alpha != nil {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  alpha: %s", ctx.versionMgr.FormatVersion(latestPreReleases.Alpha))))
		} else {
			fmt.Println(ui.HelpStyle.Render("  alpha: (none)"))
		}
		if latestPreReleases.Beta != nil {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  beta:  %s", ctx.versionMgr.FormatVersion(latestPreReleases.Beta))))
		} else {
			fmt.Println(ui.HelpStyle.Render("  beta:  (none)"))
		}
		if latestPreReleases.RC != nil {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  rc:    %s", ctx.versionMgr.FormatVersion(latestPreReleases.RC))))
		} else {
			fmt.Println(ui.HelpStyle.Render("  rc:    (none)"))
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
		return ctx.versionMgr.FormatVersion(newVersion), nil
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
