package cmd

import (
	"fmt"
	"strconv"

	"github.com/DreamCats/byte-logid/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
	Long:  "管理 byte-logid 的本地配置，包括消息净化过滤规则。",
}

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "管理消息净化过滤规则",
	Long:  "管理 ~/.config/byte-logid/filters.json 中的消息净化过滤规则。",
}

var filterListCmd = &cobra.Command{
	Use:   "list",
	Short: "查看当前过滤规则",
	RunE: func(cmd *cobra.Command, args []string) error {
		appConfig, err := config.NewAppConfig()
		if err != nil {
			return err
		}
		if err := appConfig.EnsureFilters(); err != nil {
			return err
		}

		fc, err := config.Load(appConfig.FiltersPath())
		if err != nil {
			return err
		}

		fmt.Printf("消息净化过滤规则 (%d 条):\n\n", len(fc.MsgFilters))
		for i, pattern := range fc.MsgFilters {
			fmt.Printf("  [%d] %s\n", i, pattern)
		}
		fmt.Printf("\n配置文件: %s\n", appConfig.FiltersPath())
		return nil
	},
}

var filterAddCmd = &cobra.Command{
	Use:   "add <pattern>",
	Short: "添加过滤规则",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]

		appConfig, err := config.NewAppConfig()
		if err != nil {
			return err
		}
		if err := appConfig.EnsureFilters(); err != nil {
			return err
		}

		fc, err := config.Load(appConfig.FiltersPath())
		if err != nil {
			return err
		}

		fc.AddFilter(pattern)
		if err := fc.Save(appConfig.FiltersPath()); err != nil {
			return err
		}

		fmt.Printf("已添加过滤规则: %s\n", pattern)
		return nil
	},
}

var filterRemoveCmd = &cobra.Command{
	Use:   "remove <index>",
	Short: "删除过滤规则（按索引，从 0 开始）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("无效的索引: %s", args[0])
		}

		appConfig, err := config.NewAppConfig()
		if err != nil {
			return err
		}

		fc, err := config.Load(appConfig.FiltersPath())
		if err != nil {
			return err
		}

		removed, err := fc.RemoveFilter(index)
		if err != nil {
			return err
		}

		if err := fc.Save(appConfig.FiltersPath()); err != nil {
			return err
		}

		fmt.Printf("已删除过滤规则: %s\n", removed)
		return nil
	},
}

var filterResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "重置为默认过滤规则",
	RunE: func(cmd *cobra.Command, args []string) error {
		appConfig, err := config.NewAppConfig()
		if err != nil {
			return err
		}

		fc := &config.FilterConfig{}
		fc.Reset()
		if err := fc.Save(appConfig.FiltersPath()); err != nil {
			return err
		}

		fmt.Printf("已重置为默认过滤规则 (%d 条)\n", len(fc.MsgFilters))
		return nil
	},
}

func init() {
	filterCmd.AddCommand(filterListCmd)
	filterCmd.AddCommand(filterAddCmd)
	filterCmd.AddCommand(filterRemoveCmd)
	filterCmd.AddCommand(filterResetCmd)
	configCmd.AddCommand(filterCmd)
	rootCmd.AddCommand(configCmd)
}
