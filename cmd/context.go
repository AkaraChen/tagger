package cmd

import (
	"fmt"

	"github.com/AkaraChen/tagger/internal/semver"
	"github.com/AkaraChen/tagger/internal/ui"
	sv "github.com/Masterminds/semver/v3"
)

// tagContext 包含 tag 操作的上下文
type tagContext struct {
	versionMgr     *semver.VersionManager
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
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  alpha: v%s", latestPreReleases.Alpha.String())))
		} else {
			fmt.Println(ui.HelpStyle.Render("  alpha: (none)"))
		}
		if latestPreReleases.Beta != nil {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  beta:  v%s", latestPreReleases.Beta.String())))
		} else {
			fmt.Println(ui.HelpStyle.Render("  beta:  (none)"))
		}
		if latestPreReleases.RC != nil {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  rc:    v%s", latestPreReleases.RC.String())))
		} else {
			fmt.Println(ui.HelpStyle.Render("  rc:    (none)"))
		}
		fmt.Println()
	}
}

func (ctx *tagContext) handleExistingPreRelease() (string, error) {
	info := ctx.versionMgr.GetPreReleaseInfo(ctx.currentVersion)
	if info == nil {
		return "", fmt.Errorf("failed to parse pre-release info")
	}

	if ctx.opts.PreReleaseType != "" {
		return ctx.handleNonInteractivePreRelease()
	}

	var options []ui.PreReleaseActionOption

	if bumpedVersion, err := ctx.versionMgr.BumpPreRelease(ctx.currentVersion); err == nil {
		options = append(options, ui.PreReleaseActionOption{
			Action:     ui.PreReleaseActionBump,
			NewVersion: ctx.versionMgr.FormatVersion(bumpedVersion),
			Desc:       fmt.Sprintf("increment %s number", info.Type),
		})
	}

	if info.Type != semver.PreReleaseRC {
		if advancedVersion, err := ctx.versionMgr.NextPreReleaseStage(ctx.currentVersion); err == nil {
			advancedInfo := ctx.versionMgr.GetPreReleaseInfo(advancedVersion)
			options = append(options, ui.PreReleaseActionOption{
				Action:     ui.PreReleaseActionAdvance,
				NewVersion: ctx.versionMgr.FormatVersion(advancedVersion),
				Desc:       fmt.Sprintf("%s → %s", info.Type, advancedInfo.Type),
			})
		}
	}

	if stableVersion, err := ctx.versionMgr.ReleaseStable(ctx.currentVersion); err == nil {
		options = append(options, ui.PreReleaseActionOption{
			Action:     ui.PreReleaseActionStable,
			NewVersion: ctx.versionMgr.FormatVersion(stableVersion),
			Desc:       "release as stable",
		})
	}

	action, err := ui.SelectPreReleaseAction(ctx.currentVersionStr(), options)
	if err != nil {
		return "", err
	}

	for _, opt := range options {
		if opt.Action == action {
			return opt.NewVersion, nil
		}
	}

	return "", fmt.Errorf("invalid action selected")
}

func (ctx *tagContext) handleStableVersion() (string, error) {
	if ctx.opts.PreReleaseType != "" {
		return ctx.handleNonInteractivePreRelease()
	}

	bumpOptions := []ui.BumpTypeOption{
		{Type: "patch", NewVersion: ctx.versionMgr.FormatVersion(ctx.versionMgr.BumpPatch(ctx.currentVersion)), Desc: "补丁更新"},
		{Type: "minor", NewVersion: ctx.versionMgr.FormatVersion(ctx.versionMgr.BumpMinor(ctx.currentVersion)), Desc: "小版本更新"},
		{Type: "major", NewVersion: ctx.versionMgr.FormatVersion(ctx.versionMgr.BumpMajor(ctx.currentVersion)), Desc: "大版本更新"},
	}

	bumpType, err := ui.SelectBumpType(ctx.currentVersionStr(), bumpOptions)
	if err != nil {
		return "", err
	}

	newVersion, err := ctx.versionMgr.CalculateNewVersion(ctx.currentVersion, bumpType)
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
	latestPreReleases := ctx.versionMgr.GetLatestPreReleases(ctx.versions, baseVersion)

	var options []ui.PreReleaseTypeOption

	for _, preType := range []semver.PreReleaseType{semver.PreReleaseAlpha, semver.PreReleaseBeta, semver.PreReleaseRC} {
		num := ctx.versionMgr.FindNextPreReleaseNumber(ctx.versions, baseVersion, preType)
		version, _ := ctx.versionMgr.SetPreRelease(baseVersion, preType, num)

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

		options = append(options, ui.PreReleaseTypeOption{
			Type:       preType,
			NewVersion: ctx.versionMgr.FormatVersion(version),
			LatestInfo: latestInfo,
		})
	}

	selectedType, err := ui.SelectPreReleaseType(ctx.currentVersionStr(), options)
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
	var preType semver.PreReleaseType
	switch ctx.opts.PreReleaseType {
	case "alpha":
		preType = semver.PreReleaseAlpha
	case "beta":
		preType = semver.PreReleaseBeta
	case "rc":
		preType = semver.PreReleaseRC
	default:
		return "", fmt.Errorf("invalid pre-release type: %s (must be alpha, beta, or rc)", ctx.opts.PreReleaseType)
	}

	num := ctx.opts.PreReleaseNum
	if num <= 0 {
		num = 1
	}

	baseVersion := ctx.currentVersion
	if !ctx.versionMgr.IsPreRelease(ctx.currentVersion) {
		baseVersion = ctx.versionMgr.BumpPatch(ctx.currentVersion)
	}

	newVersion, err := ctx.versionMgr.SetPreRelease(baseVersion, preType, num)
	if err != nil {
		return "", fmt.Errorf("failed to create pre-release version: %w", err)
	}

	return ctx.versionMgr.FormatVersion(newVersion), nil
}
