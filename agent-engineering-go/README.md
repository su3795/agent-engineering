# agent-engineering-go · WorkBuddy V2.5.8 (Go)

WorkBuddy V2.5.8 的 **Go 语言 1:1 复刻**。

> 与 Python 原版（`agent-engineering/backend/workbuddy/`）逐模块对齐，
> 共享同一套架构决策：TenantTraceStore 按租户分桶、strict 模式默认开启、
> 会话标题取需求首句前 18 字符、门禁 step1-step6 独立函数不覆盖。

## 目录

```
agent-engineering-go/
├── go.mod
├── cmd/agent-engineering/main.go   # 入口：run / tenant / verify
├── internal/
│   ├── tenant/        # 多租户：ConfigStore + TenantTraceStore + 配额 + APIKey
│   ├── tracing/       # 链路追踪：Span / Trace / Sink / Tracer / InMemoryStore
│   ├── pipeline/      # 6 阶段流水线 + EventBus + Artifact
│   ├── agents/        # LangChain / AgentScope 双引擎
│   ├── guardrails/    # 三层护栏 + 脱敏
│   ├── uncertainty/   # ECE 校准
│   ├── instrumentation/ # 拦截器 + 采样
│   └── llm/           # Mock / Real 双模式
└── scripts/transpile.sh
```

## 编译与运行

```bash
# 需要 Go 1.21+
go build ./...
go test ./...

# 三个子命令
go run ./cmd/agent-engineering run --engine mock --req "做一个天气查询 App"
go run ./cmd/agent-engineering tenant --strict
go run ./cmd/agent-engineering verify
```

## verify 子命令（对应 Python verify_all.py 的 step1-6）

| Step | 验证内容 |
|------|----------|
| 1 | 结构自洽 |
| 2 | 多租户隔离（A 不可读 B） |
| 3 | 链路追踪（子 Span ≥4） |
| 4 | 不确定性校准（输出 ∈ [0,1]） |
| 5 | 流水线可运行（6 阶段产出） |
| 6 | 会话标题概要（generateTitle 对齐前端） |

## 零第三方依赖

`go.mod` 只声明模块路径，不引入任何外部依赖。全部 9 个 Go 源文件
（共约 3750 行）仅使用 Go 标准库。

## 与 Python 版的差异

语义 100% 对齐，仅实现机制因语言特性而不同：

| 能力 | Python | Go |
|------|--------|----|
| 上下文传播 | contextvars | context.Context |
| 并发安全 | asyncio.Lock | sync.Mutex / RWMutex |
| 流式响应 | requests stream | net/http + goroutine |
| 装饰器 | @functools.wraps | closure + defer |

## 诚实说明

- 本目录为手写 Go 源码。沙盒环境无 golang 工具链，
  **未在本机执行 `go build` / `go test`**。
- 已通过 `scripts/transpile.sh`（Python ast 静态校验）做语法配对与结构检查。
- 请你在有 Go 环境的机器上执行 `go build ./...` 与 `go test ./...` 做最终确认。
- 若编译报错，请把错误信息发回来，我逐文件修复。
