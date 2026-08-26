# CommandCode Proxy Server

[English](README.md) | 简体中文

CommandCode API 的 OpenAI 兼容代理服务器。对外提供 `/v1/chat/completions` 与 `/v1/models` 端点，让 OpenAI 兼容的客户端通过本服务调用 CommandCode 模型。

仓库地址：https://github.com/dev2k6/command-code-proxy-server

版本：`v1.0.9`

## 功能特性

- OpenAI 兼容的 chat completions 端点
- 支持流式与非流式响应
- OpenAI 兼容的模型列表端点
- 模型短名映射
- 可选的服务端默认 API key
- 通过 `Authorization` 请求头传递每请求密钥
- **客户端密钥白名单**（`-auth-keys` / 环境变量 `CCP_AUTH_KEYS`）——限制谁可以使用本代理；上游 CommandCode 密钥保存在服务端
- 空/空字符串消息内容自动规范化（绝不会序列化出 `"content": null`）
- 超大请求体直接返回明确的 `413` 错误
- 上游返回空响应时，转为带说明的明确 `502` 错误，而不是让客户端收到莫名其妙的空回复
- 主机与端口可配置
- 启动时检查 GitHub 最新版本号并显示在当前版本旁

## 环境要求

- Go 1.26.2 或更新版本

## 运行

```bash
go run main.go
```

默认监听地址：

```text
http://127.0.0.1:55990
```

## 命令行参数

```bash
go run main.go [options]
```

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `-host` | `127.0.0.1` | 服务绑定主机 |
| `-port` | `55990` | 服务监听端口 |
| `-api-key` | 空 | 用于上游调用的 CommandCode 密钥（建议仅保存在服务端） |
| `-auth-keys` | 空 | 允许使用本代理的客户端密钥白名单，逗号分隔；留空 = 开放代理。环境变量：`CCP_AUTH_KEYS` |
| `-base-url` | `https://api.commandcode.ai` | 覆盖上游基础地址（用于测试或自建） |
| `-version` | `false` | 打印版本后退出 |

示例：

```bash
# 默认主机和端口运行
go run main.go

# 自定义端口
go run main.go -port 8080

# 监听所有网卡
go run main.go -host 0.0.0.0

# 未携带 Authorization 的请求统一使用该默认上游密钥
go run main.go -api-key your-commandcode-api-key

# 要求客户端出示白名单中的密钥（真实上游密钥只留在服务端，不外泄）
go run main.go -auth-keys "client-key-1,client-key-2" -api-key your-commandcode-api-key

# 通过环境变量达到同样效果
CCP_AUTH_KEYS="client-key-1,client-key-2" go run main.go

# 打印版本
go run main.go -version
```

## 构建

构建当前平台：

```bash
go build -o bin/command-code-proxy
```

交叉编译 Windows / Linux：

```bash
GOOS=windows GOARCH=amd64 go build -o bin/command-code-proxy.exe
GOOS=linux GOARCH=amd64 go build -o bin/command-code-proxy
```

## 密钥机制说明

整个链路涉及两种不同的密钥：

1. **客户端密钥** —— 调用本代理的一方在 `Authorization` 头里带的 key。
   配置了 `-auth-keys`（或环境变量 `CCP_AUTH_KEYS`）时，该 key 必须在白名单内，
   否则请求被拒绝并返回 `401`。
   未配置白名单时接受任意 key（开放代理模式）。
2. **上游密钥** —— 访问 `api.commandcode.ai` 时使用的 key。解析顺序：
   1. `-api-key` 命令行参数（推荐：只在服务端保管，不发给客户端）
   2. 客户端 `Authorization` 头里的 key（透传）

请求头格式：

```http
Authorization: Bearer your-key
```

如果既没有可用密钥也没有配置任何来源，则返回 `401 Unauthorized`。

## 内容规范化

OpenAI 客户端有时会发送 `"content": null` 或 `"content": ""` 的消息
（例如只包含工具调用的 assistant 回合）。CommandCode 上游会拒绝 null 内容，
因此本代理保证每条发出的消息都携带有效内容——空内容会被替换为单个文本片段。

## 请求体大小限制

请求体上限为 32 MB。超限上传会得到明确的 `413` 错误，
不会挂死，也不会被中间层静默丢弃。

## 上游空响应处理

当 CommandCode API 返回 `200` 但响应体为空时（实测发生在对话上下文超出模型上限的场景），
代理会将其转换为带有说明信息的 `502` 错误，让客户端可以做出反应
（例如压缩历史记录），而不是挂在一条死掉的流上。

## 安全警告

`-host 0.0.0.0` 且未配置 `-auth-keys` 时，本代理是**开放中转**：任何扫到地址的人都能烧光你的 CommandCode 配额。公网部署务必配合 `-auth-keys`（或保持 `-host 127.0.0.1` 挂在自己的鉴权层后面）。

## 部署

### 快速启动（二进制）

```bash
# 下载 release 二进制或自行构建（见构建章节）
chmod +x command-code-proxy
./command-code-proxy -port 55990 -auth-keys "client-key-1" -api-key "cc-secret-key"
```

### systemd（Linux）

```ini
# /etc/systemd/system/ccproxy.service
[Unit]
Description=CommandCode Proxy Server
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/opt/ccproxy/command-code-proxy -port 55990 -host 0.0.0.0 -auth-keys change-me-client-key -api-key cc-secret-key
Environment=CCP_AUTH_KEYS=change-me-client-key
Restart=always
RestartSec=3
User=nobody
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now ccproxy
journalctl -u ccproxy -f        # 跟踪日志（[REQ] 行包含模型/消息数/字节数）
```

### 国内构建加速

```bash
GOPROXY=https://goproxy.cn,direct go build -o command-code-proxy .
```

## 端点

### 健康检查

```http
GET /health
```

响应：

```json
{"status":"ok"}
```

### 模型列表

```http
GET /v1/models
```

返回 OpenAI 兼容的模型列表。

### Chat Completions

```http
POST /v1/chat/completions
```

非流式请求示例：

```bash
curl http://127.0.0.1:55990/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-commandcode-api-key" \
  -d '{
    "model": "deepseek-v4-pro",
    "messages": [
      {"role": "system", "content": "You are helpful."},
      {"role": "user", "content": "Hello"}
    ],
    "stream": false
  }'
```

流式请求示例：

```bash
curl -N http://127.0.0.1:55990/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-commandcode-api-key" \
  -d '{
    "model": "deepseek-v4-pro",
    "messages": [
      {"role": "user", "content": "Write a short poem."}
    ],
    "stream": true
  }'
```

## 支持的模型别名

代理接受完整模型 ID 以及以下短别名：

| 别名 | 映射到 |
| --- | --- |
| `deepseek-v4-pro`, `deepseek-v4`, `deepseek-pro` | `deepseek/deepseek-v4-pro` |
| `deepseek-v4-flash`, `deepseek-flash` | `deepseek/deepseek-v4-flash` |
| `minimax-m2.7`, `minimax2.7` | `MiniMaxAI/MiniMax-M2.7` |
| `minimax-m2.5`, `minimax2.5`, `minimax` | `MiniMaxAI/MiniMax-M2.5` |
| `glm-5.1` | `zai-org/GLM-5.1` |
| `glm-5` | `zai-org/GLM-5` |
| `kimi-k2.6`, `kimi2.6` | `moonshotai/Kimi-K2.6` |
| `kimi-k2.5`, `kimi2.5` | `moonshotai/Kimi-K2.5` |
| `qwen-3.6-max-preview`, `qwen3.6-max` | `Qwen/Qwen3.6-Max-Preview` |
| `qwen-3.6-plus`, `qwen3.6-plus`, `qwen3.6` | `Qwen/Qwen3.6-Plus` |
| `step-3.5-flash`, `step3.5` | `stepfun/Step-3.5-Flash` |
| `gemini-3.1-flash-lite`, `gemini-flash-lite` | `google/gemini-3.1-flash-lite` |
| `minimax-m3`, `minimax3` | `MiniMaxAI/MiniMax-M3` |
| `qwen-3.7-max-free`, `qwen3.7-max-free` | `Qwen/Qwen3.7-Max-Free` |
| `qwen-3.7-max`, `qwen3.7-max` | `Qwen/Qwen3.7-Max` |
| `step-3.7-flash`, `step3.7` | `stepfun/Step-3.7-Flash` |
| `mimo-v2.5-pro`, `mimo-pro` | `xiaomi/mimo-v2.5-pro` |
| `mimo-v2.5`, `mimo` | `xiaomi/mimo-v2.5` |

未识别的模型名原样透传。

## 项目结构

```text
.
├── README.md
├── README.zh-CN.md
├── go.mod
├── go.sum
├── main.go
├── bin
│   └── （预编译产物，如 command-code-proxy、.exe）
└── internal
    ├── api
    │   ├── commandcode.go
    │   └── openai.go
    ├── proxy
    │   ├── convert.go        # OpenAI -> CommandCode 消息转换（含内容规范化）
    │   ├── convert_test.go   # 回归测试："content" 绝不允许序列化为 null
    │   ├── model.go          # 模型别名映射
    │   ├── model_test.go
    │   └── proxy.go          # 鉴权、请求处理、SSE 流式转发
    ├── server
    │   └── server.go         # HTTP 服务、路由、体积上限、超时
    ├── update
    │   └── update.go
    └── version
        └── version.go        # npm 版本号请求头缓存
```

## 工作原理

1. 客户端向本代理发送 OpenAI 兼容请求。
2. 代理提取 system 消息、映射模型名，并把消息转换为 CommandCode 格式。
3. 代理将请求发送到 `https://api.commandcode.ai/alpha/generate`。
4. CommandCode 的 NDJSON 流式事件被转换回 OpenAI 兼容的 SSE 分块，
   或汇总为单个 JSON 响应。

## 版本检查

启动时以及执行 `-version` 时，代理会请求：

```text
https://api.github.com/repos/dev2k6/command-code-proxy-server/tags
```

如果 GitHub 最新 tag 比当前版本新，版本行会显示为：

```text
v1.0.9 (latest: v1.x.x)
```

## CommandCode 版本请求头

发往上游的请求包含：

```http
x-command-code-version: <npm command-code 最新版本号>
```

该值来自：

```text
https://registry.npmjs.org/command-code/latest
```

获取到的版本会缓存 30 分钟。若 registry 请求失败，则使用上一次缓存值；
从未成功获取过时使用 `unknown`。
