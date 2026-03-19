# logid v2.0 需求文档

> 版本: v2.0.0
> 作者: maifeng@bytedance.com
> 创建日期: 2026-03-20
> 基于: logid v0.1.2 (Rust 版本)

## 1. 项目背景

logid 是字节跳动内部的日志查询 CLI 工具，通过 Log ID（Trace ID）在内部日志服务（StreamLog）中检索分布式调用链路日志，帮助开发者快速排查问题。

当前 v0.1.x 版本存在以下可优化点：

- **命令结构冗余**：必须使用 `logid query <LOGID>` 子命令形式，日常使用中 `query` 是多余的
- **认证逻辑内置**：自己管理 `.env` 中的 `CAS_SESSION`，与 byte-auth 统一认证工具重复
- **过滤规则硬编码**：消息净化规则写死在代码中，用户无法灵活调整
- **缺少关键词筛选**：返回数据量大时，无法按关键词快速定位关键信息

v2.0 的目标是解决上述问题，使 logid 更简洁、更灵活，特别是对 AI Agent 场景更友好。

## 2. 目标用户

- 字节跳动内部后端开发工程师
- AI Agent（Claude Code 等 CLI 工具调用场景）

## 3. 功能需求

### 3.1 命令结构优化

**去掉 `query` 子命令**，logid 直接接受位置参数作为 Log ID：

```bash
# v1.x（旧）
logid query <LOGID> --region us

# v2.0（新）
logid <LOGID> --region us
```

保留 `update` 等子命令不变，clap 优先匹配子命令名称，匹配不到则当作 Log ID 位置参数处理。

#### 完整命令格式

```
logid <LOGID> [OPTIONS]
logid update [--check] [--force]
logid config [SUBCOMMAND]
logid --version
logid --help
```

### 3.2 认证委托 byte-auth

**认证方式改为委托 byte-auth CLI 获取 JWT Token**，支持两种模式：

| 模式 | 说明 | 使用场景 |
|------|------|----------|
| 自动获取 | logid 内部调用 `byte-auth token --region <region> --raw` | 默认模式，用户无感 |
| 手动传入 | `--token <TOKEN>` 参数直接指定 | CI/CD、调试、已有 token 场景 |

优先级：`--token` 参数 > 自动调用 byte-auth

#### 认证流程

```
用户执行: logid <LOGID> -r us

  ├─ 检查 --token 参数？
  │   ├─ 有 → 直接使用该 token
  │   └─ 无 → 调用 byte-auth token --region us --raw
  │           ├─ 成功 → 使用返回的 token
  │           └─ 失败 → 报错提示用户检查 byte-auth 配置
  │
  └─ 携带 token 请求日志服务
```

#### 错误处理

- byte-auth 未安装：提示用户安装 byte-auth
- byte-auth 执行失败：展示 byte-auth 的错误输出，引导用户执行 `byte-auth config set` 配置凭证
- token 无效/过期：提示用户执行 `byte-auth token --region <region> --force` 刷新

### 3.3 消息净化规则配置化

**将硬编码的过滤规则移到用户级配置文件**：`~/.config/logid/filters.json`

#### 配置文件格式

```json
{
  "msg_filters": [
    "_compliance_nlp_log",
    "_compliance_whitelist_log",
    "_compliance_source=footprint",
    "(?s)\"user_extra\":\\s*\"\\{.*?\\}\"",
    "(?m)\"LogID\":\\s*\"[^\"]*\"",
    "(?m)\"Addr\":\\s*\"[^\"]*\"",
    "(?m)\"Client\":\\s*\"[^\"]*\""
  ]
}
```

#### 行为规则

- **首次运行**：如果 `~/.config/logid/filters.json` 不存在，自动生成包含默认规则的配置文件
- **配置文件为唯一来源**：代码中不再硬编码默认规则，所有规则从配置文件读取
- **用户可编辑**：直接编辑 JSON 文件增删规则
- **命令行管理**：通过 `logid config filter` 子命令管理规则

#### config filter 子命令

```bash
# 查看当前过滤规则
logid config filter list

# 添加过滤规则
logid config filter add '<正则表达式>'

# 删除过滤规则（按索引）
logid config filter remove <index>

# 重置为默认规则
logid config filter reset
```

### 3.4 关键词过滤

**新增关键词过滤功能**，在日志条目级别进行筛选，只保留包含指定关键词的条目。

与消息净化是两个不同层次：

| 层次 | 作用 | 操作对象 | 时机 |
|------|------|----------|------|
| 消息净化 | 从 `_msg` 中**剔除**冗余内容 | 单条消息文本 | 清洗阶段 |
| 关键词过滤 | **只保留**包含关键词的条目 | 整条日志条目 | 筛选阶段 |

#### 命令行参数

```bash
# 只看包含 ForLive 的日志
logid <LOGID> -r us -k ForLive

# 多个关键词（OR 关系，包含任一即保留）
logid <LOGID> -r us -k ForLive -k ErrorCode

# 与 PSM 过滤组合使用
logid <LOGID> -r us -p my.service -k ForLive
```

| 参数 | 简写 | 类型 | 说明 |
|------|------|------|------|
| `--keyword` | `-k` | 可多次指定 | 关键词过滤，多个之间为 OR 关系 |

#### 处理流程

```
服务端返回原始数据
  │
  ├─ 1. PSM 过滤（服务端完成）
  │
  ├─ 2. 消息净化（本地，按 filters.json 规则剔除冗余内容）
  │
  └─ 3. 关键词过滤（本地，只保留包含指定关键词的条目）
      │
      └─ 输出最终结果
```

#### 匹配规则

- 大小写不敏感
- 纯字符串匹配（非正则）
- 匹配范围：日志条目中的 `_msg` 字段值（净化后的文本）
- 多个关键词为 **OR** 关系：包含任意一个关键词即保留该条目

### 3.5 PSM 过滤（保留）

PSM 过滤通过请求参数传递，由服务端完成过滤，保持现有行为不变：

```bash
logid <LOGID> -r us -p service.a -p service.b
```

### 3.6 update 子命令（保留）

保持现有自更新功能不变：

```bash
logid update           # 更新到最新版本
logid update --check   # 仅检查是否有新版本
logid update --force   # 强制更新
```

## 4. 非功能需求

### 4.1 Agent 友好

- 命令简洁，减少不必要的参数
- JSON 结构化输出，方便 Agent 解析
- 错误信息清晰，包含修复建议
- README 详细，包含完整使用示例，方便 Agent 理解和使用

### 4.2 安全性

- 配置文件权限 `0600`（仅当前用户可读写）
- 不在日志中打印 token 全文

### 4.3 跨平台

- 支持 macOS（darwin/amd64, darwin/arm64）
- 支持 Linux（linux/amd64, linux/arm64）

### 4.4 性能

- byte-auth 子进程调用开销控制在 1 秒内（byte-auth 有 token 缓存机制）
- 关键词过滤为内存操作，不引入额外 IO

## 5. 支持的区域

| 区域 | 标识 | 日志服务 URL | 状态 |
|------|------|-------------|------|
| 美区 | `us` | `https://logservice-tx.tiktok-us.org/streamlog/platform/microservice/v1/query/trace` | 可用 |
| 国际化 | `i18n` | `https://logservice-sg.tiktok-row.org/streamlog/platform/microservice/v1/query/trace` | 可用 |
| 欧洲 | `eu` | `https://logservice-eu-ttp.tiktok-eu.org/streamlog/platform/microservice/v1/query/trace` | 可用 |
| 中国 | `cn` | 待配置 | 待上线 |

## 6. 配置文件

### 6.1 存储位置

```
~/.config/logid/
├── filters.json       # 消息净化规则配置
```

### 6.2 filters.json 格式

```json
{
  "msg_filters": [
    "_compliance_nlp_log",
    "_compliance_whitelist_log",
    "_compliance_source=footprint",
    "(?s)\"user_extra\":\\s*\"\\{.*?\\}\"",
    "(?m)\"LogID\":\\s*\"[^\"]*\"",
    "(?m)\"Addr\":\\s*\"[^\"]*\"",
    "(?m)\"Client\":\\s*\"[^\"]*\""
  ]
}
```

## 7. 完整参数一览

```
logid <LOGID> [OPTIONS]

Arguments:
  <LOGID>                 要查询的 Log ID（Trace ID）

Options:
  -r, --region <REGION>   查询区域 (us/i18n/eu/cn)
  -p, --psm <PSM>         按 PSM 服务名过滤（可多次指定，服务端过滤）
  -k, --keyword <KEYWORD> 按关键词过滤日志条目（可多次指定，OR 关系）
  -t, --token <TOKEN>     手动指定 JWT Token（跳过 byte-auth 调用）
  -h, --help              显示帮助信息
  -V, --version           显示版本号

Subcommands:
  update                  更新 logid 到最新版本
  config                  配置管理
```
