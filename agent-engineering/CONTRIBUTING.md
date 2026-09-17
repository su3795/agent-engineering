# 贡献指南

## 开发流程

1. Fork → 创建分支 `feature/xxx` 或 `fix/xxx`
2. 编写代码 + 补充测试
3. 运行 `python run_tests.py` 确保全部通过
4. 提交 PR

## 代码规范

- **Python**：遵循 PEP8，使用类型注解，新模块放在 `backend/workbuddy/` 或 `backend/langchain_impl/`
- **Vue**：组件单文件 `<script setup>`，样式 scoped，主题色用 CSS 变量
- **提交信息**：`feat: ...` / `fix: ...` / `docs: ...` / `refactor: ...`

## 新增一个能力

1. 在 `langchain_impl/` 实现（能力7、8...）
2. 在 `workbuddy/pipeline.py` 的 `_capabilities_for()` 映射中加入
3. 如需 AgentScope 版本，在 `agentscope_impl/` 实现
4. 补充 `test_v2.py` 测试用例
5. 更新 `docs/structure.html` 的能力表格

## 新增一个流水线阶段

1. 在 `pipeline.py` 的 `STAGES` 中注册
2. 实现 `_run_xxx` 方法，产出 Artifact
3. 更新 `AgentScopeWorkBuddy._define_workflow`
4. 前端 `PipelineView.vue` 阶段卡片自动渲染

## 测试要求

- 单元测试 `test_v2.py`：覆盖核心逻辑，**不依赖外部 LLM**（用 Mock）
- E2E `test_e2e.py`：覆盖 HTTP API
- WS `test_ws.py`：覆盖 WebSocket 事件流
- 运行：`make all`（等价于 `python run_tests.py`）

## 文档

- 架构变更 → 更新 `docs/structure.html`（Canvas 图 + 目录树）
- API 变更 → 更新 `README.md` API 速览表
- 版本变更 → 更新 `CHANGELOG.md`

## 报告问题

提交 Issue 时请包含：
- 复现步骤
- 期望 / 实际行为
- 日志（`backend/logs/` 或终端输出）
- Python / Node 版本
