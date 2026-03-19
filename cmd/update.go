package cmd

import (
	"github.com/DreamCats/logid/internal/updater"
	"github.com/spf13/cobra"
)

var (
	updateCheck bool
	updateForce bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "更新 logid 到最新版本",
	Long: `更新 logid 到最新版本。

示例:
  logid update           # 更新到最新版本
  logid update --check   # 仅检查是否有新版本
  logid update --force   # 强制更新`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return updater.Update(version, updateCheck, updateForce)
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "仅检查是否有新版本，不执行更新")
	updateCmd.Flags().BoolVar(&updateForce, "force", false, "强制更新，即使当前已是最新版本")
	rootCmd.AddCommand(updateCmd)
}
