package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DreamCats/logid/internal/auth"
	"github.com/DreamCats/logid/internal/config"
	"github.com/DreamCats/logid/internal/filter"
	"github.com/DreamCats/logid/internal/query"
	"github.com/spf13/cobra"
)

var (
	flagRegion  string
	flagPSM     []string
	flagKeyword []string
	flagToken   string
	flagMaxLen  int
)

var rootCmd = &cobra.Command{
	Use:   "logid [LOGID] [flags]",
	Short: "字节跳动 logid 查询工具",
	Long: `logid 是一个通过 Log ID（Trace ID）查询字节跳动内部日志服务的 CLI 工具。

支持多区域查询、PSM 过滤、关键词筛选，输出 JSON 格式。
认证通过 byte-auth 自动获取，也可通过 --token 手动传入。

示例:
  logid 550e8400-e29b-41d4-a716-446655440000 --region us
  logid abc123 -r i18n -p my.service
  logid abc123 -r us -k ForLive -k ErrorCode
  logid abc123 -r us --token <jwt-token>
  logid abc123 -r us --max-len 0          # 不截断，查看完整内容

区域说明:
  us   - 美区
  i18n - 国际化区域（新加坡）
  eu   - 欧洲区
  cn   - 中国区（待上线）`,
	Args:               cobra.MaximumNArgs(1),
	DisableFlagParsing: false,
	RunE:               runQuery,
}

func init() {
	rootCmd.Flags().StringVarP(&flagRegion, "region", "r", "", "查询区域 (us/i18n/eu/cn)")
	rootCmd.Flags().StringSliceVarP(&flagPSM, "psm", "p", nil, "按 PSM 服务名过滤（可多次指定）")
	rootCmd.Flags().StringSliceVarP(&flagKeyword, "keyword", "k", nil, "按关键词过滤日志条目（可多次指定，OR 关系）")
	rootCmd.Flags().StringVarP(&flagToken, "token", "t", "", "手动指定 JWT Token（跳过 byte-auth 调用）")
	rootCmd.Flags().IntVar(&flagMaxLen, "max-len", 1000, "消息最大长度，超出截断（0 表示不截断）")
}

// Execute 执行根命令。
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// runQuery 执行日志查询的主逻辑。
func runQuery(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	logid := args[0]

	// 校验 region 参数
	if flagRegion == "" {
		return fmt.Errorf("必须指定查询区域，使用 --region 或 -r 参数\n支持的区域: us, i18n, eu, cn")
	}

	region, err := config.ParseRegion(flagRegion)
	if err != nil {
		return err
	}

	regionConfig := config.GetRegionConfig(region)
	if regionConfig == nil {
		return fmt.Errorf("获取区域 %s 配置失败", region)
	}
	if !regionConfig.Configured {
		return fmt.Errorf("区域 %s 尚未配置日志服务，请联系相关团队获取配置信息", region)
	}

	// 获取 JWT Token
	token, err := auth.GetToken(flagToken, string(region))
	if err != nil {
		return err
	}

	// 初始化配置，确保 filters.json 存在
	appConfig, err := config.NewAppConfig()
	if err != nil {
		return fmt.Errorf("初始化配置失败: %w", err)
	}
	if err := appConfig.EnsureFilters(); err != nil {
		return fmt.Errorf("初始化过滤配置失败: %w", err)
	}

	// 加载过滤规则并创建净化器
	fc, err := config.Load(appConfig.FiltersPath())
	if err != nil {
		return fmt.Errorf("加载过滤配置失败: %w", err)
	}

	sanitizer, err := filter.NewMessageSanitizer(fc.MsgFilters)
	if err != nil {
		return err
	}

	// 创建关键词过滤器
	keywordFilter := filter.NewKeywordFilter(flagKeyword)

	// 创建查询客户端并执行查询
	client := query.NewClient(regionConfig, sanitizer, keywordFilter, flagMaxLen)
	result, err := client.Query(logid, token, flagPSM)
	if err != nil {
		return err
	}

	// JSON 格式化输出
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	fmt.Println(string(output))
	return nil
}
