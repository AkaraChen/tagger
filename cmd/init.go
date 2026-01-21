package cmd

import (
	"fmt"

	"github.com/AkaraChen/tagger/internal/config"
	"github.com/AkaraChen/tagger/internal/ui"
)

// RunInit 执行 init 命令，创建默认配置文件
func RunInit() error {
	fmt.Println(ui.InfoStyle.Render("Creating tagger configuration file..."))

	err := config.CreateDefault()
	if err != nil {
		return err
	}

	schemaURL := config.GetSchemaURL()
	fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Created %s", config.ConfigFileName)))
	fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  Schema: %s", schemaURL)))
	fmt.Println()
	fmt.Println(ui.HelpStyle.Render("You can now customize your configuration:"))
	fmt.Println(ui.HelpStyle.Render("  - gitHostingProvider: GitHub or Other"))
	fmt.Println(ui.HelpStyle.Render("  - github.openActionPage: true (Actions page) or false (homepage)"))

	return nil
}
