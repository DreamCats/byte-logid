# byte-logid v2.0 设计方案

> 版本: v2.0.0
> 作者: maifeng@bytedance.com
> 创建日期: 2026-03-20
> 基于: logid v0.1.2 (Rust 版本)

## 1. 项目结构

```
src/
├── main.rs                    # CLI 入口，命令定义与分发
├── lib.rs                     # 库入口，模块导出
├── error.rs                   # 统一错误类型
│
├── auth/                      # 认证模块
│   ├── mod.rs                 # 模块导出
│   └── provider.rs            # AuthProvider：byte-auth 委托 + --token 直传
│
├── config/                    # 配置模块
│   ├── mod.rs                 # 模块导出
│   ├── region.rs              # Region 枚举与区域配置
│   ├── filter.rs              # FilterConfig：消息净化规则管理
│   └── app_config.rs          # AppConfig：应用配置管理（~/.config/byte-logid/）
│
├── log_query/                 # 日志查询模块
│   ├── mod.rs                 # 模块导出
│   ├── types.rs               # 请求/响应数据结构
│   └── client.rs              # LogQueryClient：日志查询客户端
│
├── filter/                    # 过滤模块（新增）
│   ├── mod.rs                 # 模块导出
│   ├── sanitizer.rs           # MessageSanitizer：消息净化（正则剔除）
│   └── keyword.rs             # KeywordFilter：关键词筛选
│
├── output/                    # 输出模块
│   ├── mod.rs                 # 模块导出
│   ├── format.rs              # OutputConfig 输出配置
│   ├── formatter.rs           # OutputFormatter 格式化器
│   └── utils.rs               # 工具函数
│
└── commands/                  # 子命令模块
    ├── mod.rs                 # 模块导出
    ├── update.rs              # update 子命令
    └── config_cmd.rs          # config 子命令（新增）
```

### 与 v0.1.x 的主要变更

| 模块 | 变更 |
|------|------|
| `auth/` | 删除 `manager.rs`、`multi_region.rs`，新增 `provider.rs`（委托 byte-auth） |
| `config/` | 删除 `env.rs`、`jwt.rs`，新增 `app_config.rs`（用户级配置管理） |
| `filter/` | 新增模块，从 `log_query/client.rs` 中抽离过滤逻辑 |
| `commands/` | 新增 `config_cmd.rs` |
| `main.rs` | 去掉 `Query` 子命令，改为位置参数 |

## 2. 核心模块设计

### 2.1 CLI 命令定义（main.rs）

```rust
#[derive(Parser)]
#[command(name = "byte-logid")]
#[command(about = "字节跳动 logid 查询工具")]
#[command(version)]
struct Cli {
    /// 要查询的 Log ID（与子命令互斥）
    logid: Option<String>,

    /// 查询区域 (us/i18n/eu/cn)
    #[arg(short, long)]
    region: Option<String>,

    /// 按 PSM 服务名过滤（可多次指定）
    #[arg(short, long)]
    psm: Vec<String>,

    /// 按关键词过滤日志条目（可多次指定，OR 关系）
    #[arg(short, long)]
    keyword: Vec<String>,

    /// 手动指定 JWT Token
    #[arg(short, long)]
    token: Option<String>,

    #[command(subcommand)]
    command: Option<Commands>,
}

#[derive(Subcommand)]
enum Commands {
    /// 更新 byte-logid 到最新版本
    Update {
        #[arg(long)]
        check: bool,
        #[arg(long)]
        force: bool,
    },
    /// 配置管理
    Config {
        #[command(subcommand)]
        action: ConfigAction,
    },
}

#[derive(Subcommand)]
enum ConfigAction {
    /// 管理消息净化过滤规则
    Filter {
        #[command(subcommand)]
        action: FilterAction,
    },
}

#[derive(Subcommand)]
enum FilterAction {
    /// 查看当前过滤规则
    List,
    /// 添加过滤规则
    Add {
        /// 正则表达式
        pattern: String,
    },
    /// 删除过滤规则（按索引，从 0 开始）
    Remove {
        /// 规则索引
        index: usize,
    },
    /// 重置为默认过滤规则
    Reset,
}
```

#### 命令分发逻辑

```rust
fn main() {
    let cli = Cli::parse();

    match cli.command {
        Some(Commands::Update { .. }) => { /* 执行更新 */ }
        Some(Commands::Config { .. }) => { /* 执行配置管理 */ }
        None => {
            // 没有子命令，检查是否有 logid 位置参数
            if let Some(logid) = cli.logid {
                run_query(logid, cli.region, cli.psm, cli.keyword, cli.token);
            } else {
                // 无参数，打印帮助
                Cli::parse_from(["byte-logid", "--help"]);
            }
        }
    }
}
```

### 2.2 AuthProvider（auth/provider.rs）

负责获取 JWT Token，支持两种模式：

```rust
/// 认证提供者
///
/// 支持两种认证方式：
/// 1. 手动传入 token（--token 参数）
/// 2. 自动调用 byte-auth CLI 获取 token
pub struct AuthProvider;

impl AuthProvider {
    /// 获取 JWT Token
    ///
    /// 优先使用手动传入的 token，否则调用 byte-auth 获取
    ///
    /// # 参数
    /// - `manual_token`: 用户通过 --token 手动传入的 token
    /// - `region`: 目标区域
    ///
    /// # 错误
    /// - ByteAuthNotFound: byte-auth 未安装
    /// - ByteAuthFailed: byte-auth 执行失败
    pub fn get_token(
        manual_token: Option<&str>,
        region: &str,
    ) -> Result<String, LogidError>;

    /// 调用 byte-auth CLI 获取 token
    ///
    /// 执行: byte-auth token --region <region> --raw
    /// 从 stdout 读取 token（去除尾部换行）
    fn call_byte_auth(region: &str) -> Result<String, LogidError>;

    /// 检查 byte-auth 是否已安装
    fn is_byte_auth_installed() -> bool;
}
```

#### byte-auth 调用流程

```
AuthProvider::get_token(manual_token, region)
  │
  ├─ manual_token 存在？
  │   └─ 是 → 返回 manual_token
  │
  ├─ 检查 byte-auth 是否安装
  │   └─ 未安装 → 返回 ByteAuthNotFound 错误
  │
  └─ 执行 byte-auth token --region <region> --raw
      ├─ exit code == 0 → 返回 stdout 内容（trim）
      └─ exit code != 0 → 返回 ByteAuthFailed 错误（含 stderr）
```

### 2.3 AppConfig（config/app_config.rs）

管理 `~/.config/byte-logid/` 目录下的配置文件：

```rust
/// 应用配置管理器
///
/// 负责管理 ~/.config/byte-logid/ 下的配置文件，
/// 包括首次运行时的初始化和配置文件的读写。
pub struct AppConfig {
    /// 配置目录路径 (~/.config/byte-logid/)
    config_dir: PathBuf,
}

impl AppConfig {
    /// 创建配置管理器，确保配置目录存在
    pub fn new() -> Result<Self, LogidError>;

    /// 获取 filters.json 路径
    pub fn filters_path(&self) -> PathBuf;

    /// 确保 filters.json 存在，不存在则生成默认配置
    /// 文件权限设置为 0600
    pub fn ensure_filters(&self) -> Result<(), LogidError>;
}
```

### 2.4 FilterConfig（config/filter.rs）

改造后的过滤配置，从配置文件读取规则：

```rust
/// 消息净化过滤配置
///
/// 从 ~/.config/byte-logid/filters.json 读取过滤规则，
/// 不再在代码中硬编码默认规则。
pub struct FilterConfig {
    pub msg_filters: Vec<String>,
}

impl FilterConfig {
    /// 从配置文件加载过滤规则
    pub fn load(path: &Path) -> Result<Self, LogidError>;

    /// 保存过滤规则到配置文件
    pub fn save(&self, path: &Path) -> Result<(), LogidError>;

    /// 添加一条过滤规则
    pub fn add_filter(&mut self, pattern: String);

    /// 按索引删除过滤规则
    pub fn remove_filter(&mut self, index: usize) -> Result<String, LogidError>;

    /// 获取默认过滤规则（用于 reset 和首次初始化）
    pub fn default_filters() -> Vec<String>;
}
```

### 2.5 MessageSanitizer（filter/sanitizer.rs）

从 `LogQueryClient` 中抽离出来的消息净化逻辑：

```rust
/// 消息净化器
///
/// 使用预编译的正则表达式对日志消息进行净化，
/// 剔除合规日志、足迹数据等冗余内容。
pub struct MessageSanitizer {
    /// 预编译的正则过滤器
    filters: Vec<Regex>,
}

impl MessageSanitizer {
    /// 从过滤规则列表创建净化器
    pub fn new(patterns: &[String]) -> Result<Self, LogidError>;

    /// 净化消息内容
    pub fn sanitize(&self, message: &str) -> String;
}
```

### 2.6 KeywordFilter（filter/keyword.rs）

新增的关键词过滤器：

```rust
/// 关键词过滤器
///
/// 在日志条目级别进行筛选，只保留包含指定关键词的条目。
/// 多个关键词之间为 OR 关系（包含任意一个即保留）。
/// 匹配大小写不敏感，纯字符串匹配。
pub struct KeywordFilter {
    /// 关键词列表（已转小写）
    keywords: Vec<String>,
}

impl KeywordFilter {
    /// 创建关键词过滤器
    ///
    /// # 参数
    /// - `keywords`: 关键词列表，空列表表示不过滤
    pub fn new(keywords: Vec<String>) -> Self;

    /// 是否启用过滤（关键词列表非空）
    pub fn is_active(&self) -> bool;

    /// 检查消息是否匹配任一关键词
    pub fn matches(&self, message: &str) -> bool;

    /// 过滤日志消息列表，只保留匹配的条目
    pub fn filter_messages(
        &self,
        messages: Vec<ExtractedLogMessage>,
    ) -> Vec<ExtractedLogMessage>;
}
```

### 2.7 config 子命令（commands/config_cmd.rs）

```rust
/// 执行 config filter 子命令
///
/// 支持 list/add/remove/reset 操作
pub async fn config_filter_command(action: FilterAction) -> Result<()>;
```

#### 输出示例

```bash
$ byte-logid config filter list
消息净化过滤规则 (7 条):

  [0] _compliance_nlp_log
  [1] _compliance_whitelist_log
  [2] _compliance_source=footprint
  [3] (?s)"user_extra":\s*"\{.*?\}"
  [4] (?m)"LogID":\s*"[^"]*"
  [5] (?m)"Addr":\s*"[^"]*"
  [6] (?m)"Client":\s*"[^"]*"

配置文件: ~/.config/byte-logid/filters.json

$ byte-logid config filter add 'sensitive_data'
已添加过滤规则: sensitive_data

$ byte-logid config filter remove 0
已删除过滤规则: _compliance_nlp_log

$ byte-logid config filter reset
已重置为默认过滤规则 (7 条)
```

## 3. 数据处理流水线

完整的数据处理流程：

```
┌─────────────────────────────────────────────────────────┐
│ 输入: byte-logid <LOGID> -r us -p svc.a -k ForLive      │
└───────────────┬─────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────┐
│ 1. 认证 (AuthProvider)              │
│    --token 存在？→ 直接使用          │
│    否 → byte-auth token -r us --raw │
└───────────────┬─────────────────────┘
                │
                ▼
┌─────────────────────────────────────┐
│ 2. PSM 过滤（服务端）               │
│    请求体中包含 psm_list: ["svc.a"] │
│    由日志服务完成过滤                │
└───────────────┬─────────────────────┘
                │
                ▼
┌─────────────────────────────────────┐
│ 3. 消息净化 (MessageSanitizer)      │
│    按 filters.json 规则             │
│    剔除 _msg 中的冗余内容            │
└───────────────┬─────────────────────┘
                │
                ▼
┌─────────────────────────────────────┐
│ 4. 关键词过滤 (KeywordFilter)       │
│    只保留 _msg 包含 "ForLive" 的条目 │
│    大小写不敏感，OR 关系             │
└───────────────┬─────────────────────┘
                │
                ▼
┌─────────────────────────────────────┐
│ 5. 格式化输出 (OutputFormatter)     │
│    JSON 结构化输出                   │
└─────────────────────────────────────┘
```

## 4. 错误类型扩展

新增以下错误类型：

```rust
pub enum LogidError {
    // ... 保留现有错误类型 ...

    /// byte-auth 未安装
    #[error("byte-auth 未安装，请先安装: go install github.com/DreamCats/byte-auth@latest")]
    ByteAuthNotFound,

    /// byte-auth 执行失败
    #[error("byte-auth 执行失败: {0}")]
    ByteAuthFailed(String),

    /// 过滤规则索引越界
    #[error("过滤规则索引 {0} 越界，当前共 {1} 条规则")]
    FilterIndexOutOfBounds(usize, usize),

    /// 配置目录创建失败
    #[error("无法创建配置目录 {0}: {1}")]
    ConfigDirError(String, String),
}
```

## 5. 依赖变更

### 新增依赖

无新增外部依赖。byte-auth 通过 `std::process::Command` 调用，关键词过滤为纯字符串操作。

### 可移除依赖

| 依赖 | 原因 |
|------|------|
| `dotenvy` | 不再需要读取 .env 文件（认证委托 byte-auth） |

### 保留依赖

| 依赖 | 用途 |
|------|------|
| `clap` | CLI 参数解析 |
| `tokio` | 异步运行时 |
| `reqwest` | HTTP 客户端（日志查询） |
| `serde` / `serde_json` | JSON 序列化 |
| `regex` | 消息净化正则 |
| `chrono` | 时间处理 |
| `tracing` | 日志 |
| `anyhow` / `thiserror` | 错误处理 |
| `dirs` | 用户目录路径 |
| `flate2` / `tar` / `zip` / `sha2` | 自更新 |

## 6. 安全设计

- `~/.config/byte-logid/` 目录权限：`0700`
- `~/.config/byte-logid/filters.json` 文件权限：`0600`
- `--token` 参数值不在日志中打印全文，仅打印前 8 字符 + `***`
- byte-auth 子进程调用不通过 shell（直接 `Command::new("byte-auth")`），避免注入风险

## 7. 测试策略

### 单元测试

| 模块 | 测试重点 |
|------|----------|
| `AuthProvider` | token 优先级、byte-auth 未安装处理、执行失败处理 |
| `FilterConfig` | 加载/保存/增删/重置、默认规则生成 |
| `MessageSanitizer` | 各条正则规则的匹配与清除、多规则组合 |
| `KeywordFilter` | 大小写不敏感、OR 逻辑、空关键词不过滤、多关键词组合 |
| `AppConfig` | 配置目录创建、首次初始化、文件权限 |
| CLI 解析 | 位置参数 vs 子命令、参数组合 |

### 集成测试

- 完整查询流程（mock byte-auth + mock 日志服务）
- config filter 子命令 CRUD 操作
- 关键词过滤端到端验证
