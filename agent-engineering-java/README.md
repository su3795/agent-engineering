# Agent-Engineering · Java V2.5.8

> 1:1 Java port of [Agent-Engineering](https://github.com/su3795/agent-engineering) V2.5.8.
> 对应 Python 原版与 Go 版（`v2.5.8-python` / `v2.5.8-go` 分支）。

## 模块映射

| 模块 | 包路径 | 对应 Python | 对应 Go |
|---|---|---|---|
| 多租户 | `com.agentengineering.tenant` | `backend/workbuddy/tenant.py` | `internal/tenant` |
| 链路追踪 | `com.agentengineering.tracing` | `backend/workbuddy/tracing.py` | `internal/tracing` |
| 6 阶段流水线 | `com.agentengineering.pipeline` | `backend/workbuddy/pipeline.py` | `internal/pipeline` |
| Agent 双引擎 | `com.agentengineering.agents` | `langchain_agents.py` / `agentscope_agents.py` | `internal/agents` |
| 可靠性护栏 | `com.agentengineering.guardrails` | `backend/workbuddy/guardrails_v2.py` | `internal/guardrails` |
| 不确定性校准 | `com.agentengineering.uncertainty` | `backend/workbuddy/uncertainty.py` | `internal/uncertainty` |
| 埋点 | `com.agentengineering.instrumentation` | `backend/workbuddy/instrumentation.py` | `internal/instrumentation` |
| LLM 配置 | `com.agentengineering.llm` | `backend/workbuddy/llm_config.py` | `internal/llm` |

## 构建与运行

```bash
# 编译
mvn compile

# 测试（Tenant 15 个断言）
mvn test

# 打包可执行 jar
mvn package
java -jar target/agent-engineering-java-2.5.8.jar

# 或指定子命令
java -jar target/agent-engineering-java-2.5.8.jar tenant
java -jar target/agent-engineering-java-2.5.8.jar verify
java -jar target/agent-engineering-java-2.5.8.jar pipeline
java -jar target/agent-engineering-java-2.5.8.jar trace
```

## 依赖

- 运行时：**纯 JDK 标准库，零第三方依赖**
- 编译：JDK 11+
- 测试：JUnit 5.10.2（test scope 自动下载）

## 诚实说明

本实现为 Java 语言层面的 1:1 功能复刻（语义等价，非逐字节翻译）。
- `LangChain` / `AgentScope` 引擎、`LLMConfig.real()` 流式调用为**接口骨架**，真实对接 langchain4j 等需另行实现。
- `TracerStore.OtelExporter` 为**占位 Sink**，未实现 OTLP 网络发送。
- 沙盒环境无 `javac` / `mvn`，**本次未在本环境跑通 `mvn compile` / `mvn test`**；编译与测试结果需在你本机验证。

## License

Apache 2.0
