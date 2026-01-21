package cmd

import (
	"fmt"

	"github.com/AkaraChen/tagger/internal/config"
	"github.com/AkaraChen/tagger/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "创建配置文件",
	Long:  `创建 tagger.config.json 配置文件，用于自定义 tagger 的行为`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(ui.InfoStyle.Render("Creating tagger configuration file..."))

		err := config.CreateDefault()
		if err != nil {
			return err
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Created %s", config.ConfigFileName)))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  Schema: %s", config.SchemaURL)))
		fmt.Println()
		fmt.Println(ui.HelpStyle.Render("You can now customize your configuration:"))
		fmt.Println(ui.HelpStyle.Render("  - gitHostingProvider: GitHub or Other"))
		fmt.Println(ui.HelpStyle.Render("  - github.openActionPage: true (Actions page) or false (homepage)"))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
