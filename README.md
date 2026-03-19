# logid

内部日志查询 CLI 工具，通过 Log ID（Trace ID）查询分布式调用链路日志。

> 仅供内部使用，请勿对外传播。

## 功能特性

- **极简命令** - `logid <LOGID> -r us`，无需子命令
- **多区域支持** - 美区 (us)、国际化 (i18n)、欧洲 (eu)、中国 (cn)
- **认证委托** - 自动调用 byte-auth 获取 JWT，也支持 `--token` 手动传入
- **消息净化** - 可配置的正则规则，自动剔除冗余日志内容
- **关键词过滤** - 按关键词筛选日志条目，快速定位关键信息
- **PSM 过滤** - 按服务名过滤（服务端执行）
- **消息截断** - 默认截断超长消息（1000 字符），节省 token 开销
- **JSON 输出** - 结构化 JSON 格式，方便解析和 Agent 调用
- **自更新** - 内置 update 命令

## 安装

### 一键安装（推荐）

```bash
go install github.com/DreamCats/byte-logid/cmd/logid@latest
```

### 从源码构建

```bash
git clone git@github.com:DreamCats/byte-logid.git
cd byte-logid
make build
```

### 安装到 ~/.local/bin

```bash
make install
```

### 前置依赖

需要安装 byte-auth 用于自动认证：

```bash
go install <byte-auth-repo>@latest
byte-auth config set --region us --cookie "your_cas_session_value"
```

## 快速开始

```bash
# 1. 确保 byte-auth 已配置好区域凭证
byte-auth config show

# 2. 查询日志
logid <trace-id> --region us

# 3. 带 PSM 过滤
logid <trace-id> -r us -p my.service

# 4. 带关键词过滤
logid <trace-id> -r us -p my.service -k error
```

## 命令参考

### 查询日志（默认行为）

```
logid <LOGID> [OPTIONS]

Arguments:
  <LOGID>                   要查询的 Log ID（Trace ID）

Options:
  -r, --region <REGION>     查询区域 (us/i18n/eu/cn)  [必填]
  -p, --psm <PSM>           按 PSM 服务名过滤（可多次指定，服务端过滤）
  -k, --keyword <KEYWORD>   按关键词过滤日志条目（可多次指定，OR 关系）
  -t, --token <TOKEN>       手动指定 JWT Token（跳过 byte-auth 调用）
      --max-len <N>         消息最大长度，超出截断（默认 1000，0 表示不截断）
  -h, --help                显示帮助信息
  -V, --version             显示版本号
```

### 示例

```bash
# 基础查询
logid <trace-id> --region us

# 按 PSM 过滤
logid <trace-id> -r us -p my.service

# 多个 PSM 过滤
logid <trace-id> -r i18n -p service.a -p service.b

# 关键词过滤 - 只看包含指定关键词的日志
logid <trace-id> -r us -k error

# 多关键词过滤（OR 关系）
logid <trace-id> -r us -k error -k timeout

# PSM + 关键词组合过滤
logid <trace-id> -r us -p my.service -k error

# 查看完整消息内容（不截断）
logid <trace-id> -r us --max-len 0

# 自定义截断长度
logid <trace-id> -r us --max-len 2000

# 手动传入 token
logid <trace-id> -r us --token "your-jwt-token"

# 配合 byte-auth 手动获取 token
logid <trace-id> -r us --token $(byte-auth token -r us --raw)

# 配合 jq 提取特定字段
logid <trace-id> -r us | jq '.messages[].values[].value'

# 只看消息数量
logid <trace-id> -r us | jq '.total_items'
```

### 配置管理

```bash
# 查看消息净化过滤规则
logid config filter list

# 添加自定义过滤规则
logid config filter add '<regex-pattern>'

# 删除过滤规则（按索引）
logid config filter remove <index>

# 重置为默认规则
logid config filter reset
```

### 自更新

```bash
# 检查是否有新版本
logid update --check

# 更新到最新版本
logid update

# 强制更新
logid update --force
```

### 版本信息

```bash
logid version
```

## 输出格式

```json
{
  "logid": "<trace-id>",
  "region": "us",
  "region_display_name": "美区",
  "total_items": 16,
  "filtered_items": 1,
  "messages": [
    {
      "id": "item1-val1",
      "group": {
        "psm": "my.service",
        "pod_name": "pod-abc",
        "ipv4": "10.0.0.1",
        "env": "production"
      },
      "values": [
        {
          "key": "_msg",
          "value": "RPC Call method=DoSomething info: req=... resp=...[truncated, original: 33395 chars]"
        }
      ],
      "level": "Info",
      "location": "handler.go:42"
    }
  ],
  "meta": {
    "scan_time_range": [{"start": 1710000000, "end": 1710000600}],
    "level_list": ["Info", "Warn", "Error"]
  },
  "timestamp": "2026-03-20T12:00:00Z"
}
```

**字段说明：**

| 字段 | 说明 |
|------|------|
| `logid` | 查询的 Log ID |
| `region` | 查询区域 |
| `region_display_name` | 区域中文名 |
| `total_items` | 净化后的总条目数 |
| `filtered_items` | 关键词过滤后的条目数（仅在使用 `-k` 时出现） |
| `messages` | 日志消息列表 |
| `messages[].group` | 分组信息（PSM、Pod、IP 等） |
| `messages[].values` | 日志键值对（`_msg` 为主要内容，已净化） |
| `messages[].values[].value` | 消息内容，超长时截断并标记 `...[truncated, original: N chars]` |
| `messages[].level` | 日志级别 |
| `messages[].location` | 代码位置 |
| `meta` | 元数据（扫描时间范围、日志级别列表） |
| `timestamp` | 查询时间戳 |

## 支持的区域

| 区域 | 标识 | 状态 |
|------|------|------|
| 美区 | `us` | 可用 |
| 国际化区域 | `i18n` | 可用 |
| 欧洲区 | `eu` | 可用 |
| 中国区 | `cn` | 待上线 |

## 认证方式

logid 使用 byte-auth 进行统一认证。

### 认证流程

```
logid <LOGID> -r us
  │
  ├─ --token 参数存在？→ 直接使用
  │
  └─ 自动调用: byte-auth token --region us --raw
      ├─ 成功 → 使用返回的 JWT Token
      └─ 失败 → 提示检查 byte-auth 配置
```

### Token 优先级

1. `--token` 参数（最高优先级）
2. byte-auth 自动获取

### 常见认证问题

```bash
# byte-auth 未安装
# → 安装 byte-auth

# byte-auth 未配置区域凭证
# → byte-auth config set --region us --cookie "your_cas_session"

# Token 过期
# → byte-auth token --region us --force
```

## 过滤体系

logid 提供三层过滤 + 消息截断机制：

```
服务端返回原始数据
  │
  ├─ 1. PSM 过滤（服务端，--psm）
  │     只返回指定服务的日志
  │
  ├─ 2. 消息净化（本地，filters.json）
  │     从 _msg 字段中剔除冗余内容
  │
  ├─ 3. 关键词过滤（本地，--keyword）
  │     只保留包含指定关键词的条目
  │     注意：关键词匹配在截断之前，基于完整文本
  │
  └─ 4. 消息截断（本地，--max-len）
        截断超长消息，默认 1000 字符
        末尾标记 ...[truncated, original: N chars]
```

### 消息净化规则

配置文件位于 `~/.config/logid/filters.json`，首次运行自动生成默认规则。

通过命令管理：

```bash
logid config filter list       # 查看规则
logid config filter add '...'  # 添加规则
logid config filter remove 0   # 删除规则
logid config filter reset      # 重置为默认
```

### 关键词过滤

- 大小写不敏感
- 纯字符串匹配（非正则）
- 多个关键词为 OR 关系
- 匹配范围：净化后的完整 `_msg` 字段（截断前）

### 消息截断

- 默认截断长度：1000 字符
- `--max-len 0`：不截断，查看完整内容
- `--max-len 2000`：自定义截断长度
- 截断后末尾标记：`...[truncated, original: N chars]`
- Agent 看到截断标记后，如需完整内容可用 `--max-len 0` 重查

## 在脚本中使用

```bash
#!/bin/bash
# 批量查询多个 logid
for id in "<id-1>" "<id-2>" "<id-3>"; do
  echo "=== Querying $id ==="
  logid "$id" -r us -p my.service -k error 2>/dev/null
done
```

```bash
# 提取所有错误消息
logid <trace-id> -r us -k error | jq -r '.messages[] | "\(.level) \(.location): \(.values[0].value)"'
```

## 文件结构

```
~/.config/logid/
└── filters.json       # 消息净化规则配置
```

## 项目结构

```
logid/
├── main.go                        # 程序入口
├── Makefile                       # 构建脚本
├── cmd/                           # CLI 命令定义（Cobra）
│   ├── root.go                    # 根命令 + 查询逻辑
│   ├── config.go                  # config filter 子命令
│   ├── update.go                  # update 子命令
│   └── version.go                 # version 子命令
├── internal/
│   ├── auth/
│   │   └── provider.go            # 认证提供者（byte-auth 委托）
│   ├── config/
│   │   ├── region.go              # 区域枚举 + 服务配置
│   │   ├── app_config.go          # 应用配置管理
│   │   └── filter.go              # 过滤规则配置
│   ├── filter/
│   │   ├── sanitizer.go           # 消息净化器
│   │   └── keyword.go             # 关键词过滤器
│   ├── query/
│   │   ├── client.go              # 日志查询客户端
│   │   └── types.go               # 数据结构定义
│   └── updater/
│       └── updater.go             # 自更新逻辑
└── docs/
    ├── requirements.md            # 需求文档
    └── design.md                  # 设计方案
```

## 开发

```bash
# 运行测试
make test

# 构建
make build

# 跨平台构建
make build-all

# 清理
make clean
```

## 声明

本工具仅供内部使用。
