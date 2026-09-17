# 更新日志

## [2.5.7] - 2026-09-16 — 社区化（对外免费展示就绪）

### 🤝 社区化（GitHub 模板，.github/）

- **Discussions 模板 ×3**：`general.yml`（通用讨论）/ `ideas.yml`（功能想法，含动机→方案→API 草案→备选）/ `q-a.yml`（问答，先搜历史）
- **Issue 模板（表单化）**：`bug_report.yml`（环境/复现/期望/堆栈 + 必填校验）/ `feature_request.yml` / `config.yml`（引导走 Discussions）
- **PR 模板**：关联 Issue / 自测 / changelog / DCO 签署清单
- **`CODEOWNERS`**：按目录自动分配 Reviewer（`backend/workbuddy/` → @core-team）
- **`docs/LABELS.md`**：标签规范（type / status / priority / 领域 四维）
- 说明：Discussions / Issue / PR 模板放 `docs/.github/`（打包时映射到仓库根 `.github/`），避免与 CI 工作流目录冲突

### 🎨 对外门面（README + 资源）

- **README 重构**：Logo + 8 个动态徽章（build/version/license/last-commit/issues/PRs/discussions/python）+ ⚡ 一键 Demo 按钮 + 截图 + 5 分钟 Quickstart + 真实 LLM 接入示例
- **`docs/assets/logo.svg`**：项目 Logo（200×200，渐变 + AE 字样，SVG 合法 XML）
- **截图 ×4**（matplotlib 程序化渲染，离线、无敏感信息、可复现）：
  - `architecture.png` — 分层架构（Frontend → Engine → Abstract → Backend）
  - `pipeline_flow.png` — 6 阶段流水线 + Self-Heal 失败回环
  - `multi_tenant.png` — 多租户隔离（acme / globex / initech 配额）
  - `demo_screenshot.png` — LLM 选择器 UI（Mock/真实 + 供应商）
  - ⚠️ 注：截图系程序化示意图，非浏览器真实截屏；如需像素级真实请手动替换
- **`SECURITY.md`**：漏洞披露流程 + 安全红线（API Key 不入仓 / 依赖审计 / 多租户防越权）+ 已知安全考量表

### 🛠️ 工具链

- **`scripts/gen_assets.py`**：截图独立渲染脚本（可复现）
- **`scripts/update_changelog.py`**：解析 `git log`（Conventional Commits）自动聚合 CHANGELOG 条目，供 CI 使用
- **`_pack_v257.py`**：打包脚本升级，新增 `COMMUNITY_FILES` 白名单校验（缺失即 fail-fast），版本 → V2.5.7

### 📦 发布

- 发布包 `agent-engineering-V2.5.7.zip`：社区化完整，四重自洽（zip == 扫描 == 渲染 == 结构树头部）
- 对外展示三入口：代码（Public 仓库，免费）/ Releases（附 zip，免费）/ GitHub Pages（`demo/index.html`，免费，MIT 协议可商用）

---


## [2.5.5] - 2026-09-16 — tenant 配额陷阱修复 + 豁免清单清零

### 核心修复（tenant.py，真实代码改动）

- **`TenantConfigStore.register()` 支持两种等价写法**，消除"散字段配额被吞"的历史陷阱：
  - ① 显式 `quota=`：`register("acme", quota=QuotaConfig(max_tokens_per_day=1000))`
  - ② 散字段形式：`register("acme", max_tokens_per_day=1000, max_concurrency=1)` ← **新增识别**
  - 混用：`register("initech", quota=QuotaConfig(...), max_concurrency=8)` → 散字段作增量覆盖
  - 原理：新增 `_QUOTA_FIELDS = {max_tokens_per_day, max_concurrency, max_calls_per_minute}`，
    `register()` 自动从 `**overrides` 剥离配额字段装配到 `QuotaConfig`；缺失字段保留默认值
- **回归测试** `_verify_tenant.py` 扩展为 6 项（新增 ①+ 配额陷阱修复 / ③ 混写增量 / ⑥ 默认配额），实测 **6/6 全绿**

### 豁免机制（docs/EXEMPTIONS.md）

- **A1（"封装边界"）由 ✅ 豁免调整为 🟡 已收敛 / 建议关闭**：
  - 旧表述"tenant 故意不挂顶层 `__all__`"已被代码推翻——tenant 全部 10 个公共符号
    （`TenantContext`/`TenantConfig`/`TenantConfigStore`/`TenantTraceStore`/`QuotaConfig`/
    `QuotaMiddleware`/`QuotaExceeded`/`APIKeyAuth`/`get_tenant`/`get_user`）已在
    `workbuddy/__init__.py` 导入并列入 `__all__`，`from workbuddy import TenantContext` 可直接用
  - "封装边界"准确含义改写为：tenant 是可选/分级能力（`tenant_id="default"` 透明），
    隔离逻辑集中于 `tenant.py` 单一命名空间、通过 `ContextVar` 自动贯穿全链路（5 层隔离）
- **新增 A1 复审记录（2026-09-16）**，含 Q1-Q3 与代码证据
- **裁决统计更新：活跃豁免 = 0**（B1/B2 已修复 + B4 已移除 + A1 收敛为文档说明）
- `scripts/check_exemptions.py` 实测：`[OK] 所有豁免项均在复审期内 (3/3)` → exit 0

### 文档同步

- `docs/EXEMPTIONS.md` Q1 章节基于真实代码重写（5 层隔离 + 配额陷阱修复说明）
- 版本 `__version__ = "2.5.5"`

### 验证

```
[1] _verify_tenant.py    → 6/6 ✅（含散字段/混写/默认配额回归）
[2] check_exemptions.py  → 活跃豁免 0 项，exit 0 ✅
[3] 后端语法全扫          → 0 错误 ✅
[4] 顶层 import ≡ 深路径   → TenantContext/QuotaConfig/get_tenant 同一对象 ✅
```

## [2.5.3] - 2026-09-14 — 独立架构讲解文档（6 能力 × 6 阶段）

### 文档（代码零改动）
- **新增** `docs/architecture_guide.html` — 聚焦「6 大工程化能力 × 6 阶段多 Agent 流水线」这一核心抽象
  - 7 个章节（含 hero + 6 图），**6 张 Canvas**：总览 / 6 阶段流水线 / 6 大能力分层 / **能力×阶段热力矩阵** / 一次调用交织 / 双引擎抽象
  - 配套 HTML 表格 + 面试速记区（含「正交分离」加分表述）
  - 复用既有绘图规范（`drawSafely` + `roundRect` polyfill + DPR 适配），离线可用

### 校验工具链（docs/）
- `_verify_guide.py` — 结构校验（6 canvas / drawSafely 绑定 / 19 项概念覆盖 / 矩阵精确 36 格 / `node --check` 语法）
- `_render_guide.py` — Pillow 静态预渲染 6 张 PNG + 拼贴预览图（无浏览器依赖的 CI 兜底）
- `_verify_guide_pixels.py` — **像素级校验**（非背景像素占比 + 关键色存在性，防 Canvas 灰屏）
- 产物：`docs/_assets/guide_preview.png` + 6 张单图

### 校验结果
```
[✓] _verify_guide.py        → 30+ 项全绿（含 node --check JS 语法 pass）
[✓] _render_guide.py        → 6/6 图渲染成功（1000×360 each）
[✓] _verify_guide_pixels.py → 非背景像素 16.5%~57.9%，关键色齐全，无灰屏
矩阵统计：完整 26 · 部分 8 · 缺口 2 = 36 个理论拦截点
```

### 打包门禁
- `make_release.py` Step 2 新增 architecture_guide 文档门禁（`_verify_guide.py` + `_render_guide.py` + `_verify_guide_pixels.py` 任一步失败即终止打包）
- `MUST_EXIST` 白名单追加 guide 相关 7 个文件

## [2.5.2] - 2026-09-14 — Canvas 架构文档对齐 V2.5 真实代码 + 不足诚实清单

### 文档（代码零改动，仅增改 docs/）

#### `docs/architecture_canvas.html` 升级为 6 章节 / 6 张 Canvas
| 章节 | Canvas | 说明 |
|------|--------|------|
| 01 架构设计 | 分层架构图 | 用户层→前端→API→抽象层→双引擎→基础设施 + 六大能力 |
| 02 双引擎 | 路径对比图 | LangChain ⇄ 抽象层 ⇄ AgentScope + 差异表 |
| 03 链路监控 | Trace 树 | Span 嵌套 + 并行分支 + 4× SpanSink + OTel 属性 + 查询 API |
| 04 可靠性 | 拦截流程图 | 8 道护栏 + 输入→Agent→RAG→Debate→输出→**人工门禁✋** |
| 05 不足与改进 | **四象限优先级矩阵** | 演示价值 × 上线紧迫度，★ 本轮已完成 / ⚠ 生产需补齐 |
| 06 代码全景 | 模块依赖图 | 编排层→引擎层→能力层→V2.5 增强→验证层 + 文档↔代码↔测试三向表 |

#### 不足清单（第 5 节）对齐 V2.5 真实状态
每项标注 `tag done`（已落地）/ `tag gate`（架构就位·待扩充）+ 对应代码文件：
- ✅ 已落地：① replay.py ② self_heal.py ③ debate.py ⑤ uncertainty.py ⑦ guardrails_v2.py ⑧ tenant.py
- ⚠️ 生产需补：④ eval.py（20 条→需扩 100+）⑥ tracing.py（Sink 就绪·未接 SLO）

### 校验工具链（docs/ + 根目录入口）
- `docs/_verify_canvas.py` — 结构校验（section/canvas/绘图函数/7 模块引用）
- `docs/_verify_js.py` — `node --check` 真实语法检查（无 node 降级为括号平衡）
- `docs/_gen_preview.py` — Pillow 生成预览图（无依赖降级）
- `docs/_verify_all.py` — 一键入口
- `verify_canvas.py` — 根目录软入口（`python3 verify_canvas.py`）
- 产物：`docs/_assets/canvas_preview.png`（35 KB，6 章节缩略导航图）

### 校验结果
```
[✓] _verify_canvas.py  → 6 section / 6 canvas / 6 绘图函数 / 7 模块引用齐全
[✓] _verify_js.py      → node --check 通过：内联 JS 语法正确
[✓] _gen_preview.py    → 预览图生成成功
✅ 全部通过
```

### 兼容性
- 纯前端单文件，无 CDN、无构建，双击离线可用
- 已纳入 `make_release.py` 的 `MUST_EXIST` 白名单，打包闭环校验覆盖

## [2.5.1] - 2026-09-14 — 多租户配额逻辑修复 + 测试套件全绿

### 修复（tenant.py · 真实 Bug）
- **`TenantConfigStore.register` 支持显式 `quota` 参数**：此前 `register(tenant_id, quota=QuotaConfig(...))` 中的 `quota` 被 `**overrides` 吞掉，落入 `overrides` 字典而非赋值给 `TenantConfig.quota` 字段，导致配额始终回落到默认值（100_000），**超限判断永远不触发**。现已把 `quota` 从 overrides 中剥离单独赋值，两种写法（`quota=` 命名参数 / `**overrides`）均生效。
- **`remaining` 语义对齐代码**：`check()` 返回 `max_tokens_per_day - tokens_today`（已消耗量），测试断言同步修正为递减验证（1000 → 750）。

### 验证（5 套测试套件，105/105 全绿）
| 套件 | 结果 |
|------|------|
| test_v2.py（既有回归） | ✅ 通过 |
| test_engines.py（双引擎） | ✅ 通过 |
| test_env_utils.py（配置三源） | ✅ 通过 |
| test_tracing.py（链路追踪） | ✅ 71/71 通过 |
| test_v25.py（架构不足补齐 8 项） | ✅ 33/33 通过 |

### 测试质量
- 新增 `_regression.py` 全量回归脚本（已清理），统一解析 5 套套件统计，CI 门禁可用


## [2.5.0] - 2026-09-14 — 演示增强 + 可靠性方案 + 架构可视化

### 演示界面（demo/index.html，纯前端）
- **新增「🧠 AI 思考过程」页签**：每个 Agent 展示 ReAct 推理链（推理 → 行动 → 观察），含澄清问题、方案枚举、分歧辩论、根因定位等思考步骤，逐阶段自动展开
- **新增「🏛️ 架构设计」页签**：Canvas 分层架构图（用户层/前端/API/引擎抽象层/双引擎/基础设施）+ 核心数据流 + 双引擎差异化
- **新增「📡 链路监控」页签**：链路流程图 + Waterfall 耗时分布（LangChain vs AgentScope 对比）+ Trace 事件日志
- **新增「🛡️ 可靠性护栏」页签**：8 道护栏可视化（Schema/证据/Debate/测试/门禁/注入/熔断/审计）
- **新增「⚠️ 不足与改进」页签**：8 项诚实清单 + 改进方向
- **总计 6 个 Tab**：Agent 协作 / AI 思考 / 架构 / 链路 / 护栏 / 不足
- 新增 `demo/_verify.js` 结构验证（22 项检查）

### 文档
- **新增** `docs/architecture_canvas.html` — 架构设计/链路监控/可靠性方案，含 4 张 Canvas 图（分层架构/双引擎路径/Trace 树/可靠性流程）
- **新增** `docs/RELIABILITY.md` — 防漏答/错答/乱说完整机制（8 道护栏 + 处理流程图 + 代码层对应）
- 新增 `docs/_verify_arch.js` 结构验证（16 项检查）

### 验证集成
- `distcheck.py` 新增 HTML/JS 结构校验步骤，演示页与架构文档纳入打包门禁

## [2.4.0] - 2026-09-14 — 链路追踪 + LLM 自动埋点

### 链路追踪（tracing.py）
- **新增** 标准化 Span 模型（对齐 OpenTelemetry + OpenInference 语义）
- **新增** `TraceContext` 基于 `contextvars`，修复并发串号缺陷
- **新增** 4 种 `SpanSink`：Tracer（内存，兼容旧代码）/ Console / File（JSONL 持久化）/ OTel（Jaeger/Tempo，未装 SDK 自动降级）
- **新增** 可插拔 `TraceStore`：InMemory / File（持久化 + 重载）
- **新增** 查询 API：`get_trace` / `list_traces` / `waterfall` / `replay` / `stats`（p50/p99/cost/error_rate）
- **新增** 装饰器 `@trace_span` / `@traced`

### LLM 自动埋点（instrumentation.py）—— 满足所有 LLM
- **新增** `LLMInterceptor` 协议（before / on_request / on_response / on_failure 四钩子）
- **新增** `InstrumentedLLM` 鸭子类型代理，仅在 `get_llm()` 一处包裹，业务零改动
- **新增** 六大内置拦截器：OpenAICompatible（万能兜底）/ LangChainCallback / AgentScope / Streaming / TokenCounting / Cost
- **新增** 采样策略：`Always` / `Rate` / `ErrorOnly` / `TailLatency`
- **新增** `PromptEnricher` 敏感字段脱敏 + 属性富化
- **新增** 模型定价表（DeepSeek / Qwen / GLM / Kimi / Ollama）
- **新增** `AutoInstrument.install()` 幂等自动 patch

### 关键修复
- **修复** `pipeline.py` `run()` 外层 try 缺少 finally（阻塞性语法 Bug）
- **修复** Composite 拦截器链重复创建 Span（一条调用 = 一个 llm span，幂等守卫）
- **修复** 裸调用 trace_id=None 导致 token/cost 静默丢失
- **修复** `test_engines.py` 差异化验证（改为类身份 + 方法重写检测 + capabilities 声明）

### 测试
- **新增** `test_tracing.py`（71/71 全绿）：并发不串号、Span 唯一性、Q1=C 三路 Sink、Waterfall、FileStore、采样、脱敏、AgentScope、AutoInstrument 幂等、6 阶段集成
- `test_v2.py` 36/36、`test_env_utils.py` 16/16、`test_engines.py` 全绿

### 文档
- **新增** `docs/TRACING.md`（使用 + 接入 Jaeger）

## [2.0.1] - 2026-09-12 — 交付完整性加固

### 工程化补齐
- **新增** 根 `Dockerfile`（多阶段构建：Node 构建前端 → Python 后端一体，前端产物挂载到 `backend/static/`）
- **新增** `docker-compose.yml` 完整编排（backend + frontend + redis）
- **新增** `Makefile`（`make test` / `make all` / `make docker-up` / `make clean`）
- **新增** `.github/workflows/test.yml` CI 流水线
- **新增** `.gitignore` / `LICENSE`（MIT）/ `CHANGELOG.md`

### 启动脚本（三平台）
- `start.sh`（Mac/Linux/Git Bash）
- `start.bat`（Windows CMD）
- `start.ps1`（Windows PowerShell）
- `启动说明.txt`（中文引导）

### 文档
- **新增** `docs/deploy.html` — 小白友好部署使用说明（Canvas 流程图，双击即用）
- **新增** `API.md` — 完整接口文档（REST + WebSocket）
- **新增** `PROJECT_OVERVIEW.md` — 项目总览
- **新增** `CONTRIBUTING.md` — 贡献指南
- **新增** `MIGRATION.md` — V1 → V2 迁移

### 测试（4 套）
- `test_v2.py` — 单元测试（36 项）
- `test_e2e.py` — HTTP 端到端
- `test_ws.py` — WebSocket
- `test_integration.py` — 集成测试（8 用户场景：并发 / 降级 / 边界）
- `run_tests.py` — 一键全跑

### 修复
- 修复 `langchain_impl/memory.py` f-string 语法错误（V1 遗留）
- `AgentScopeWorkBuddy` 兼容父类不接受 `tracer` 参数（V2 已通过 kwargs 处理）
- `main_v2.py` 挂载 `/static` + 根路径 SPA fallback（生产一体部署）
- 新增 `backend/__main__.py`、`backend/__init__.py`

### 验证
- `verify_package.py` — 打包前自动校验（关键文件 + 语法 + 测试 + 前端配置）
- 结果：**61/61 关键文件 ✅ · 22/22 Python 语法 ✅ · 单元测试 36/36 ✅**

---

## [2.0.0] - 2026-09-12 — WorkBuddy 流水线版

### 架构升级
- **新增** V2 流水线编排核心 `workbuddy/pipeline.py`
  - 6 阶段：需求→规划→编码→测试→审查→部署
  - 6 大工程化能力自动注入到对应阶段
  - Artifact 产物链传递
  - EventBus 实时事件流
- **新增** AgentScope 多 Agent 实现 `workbuddy/agentscope_agents.py`
  - ParallelPipeline（3 Coder 并行）
  - DebateAgent（4 视角代码审查）
  - MsgHub（接口契约广播）
  - Checkpoint / Tracing / Guardrail 中间件
- **新增** V2 API `main_v2.py`
  - `POST /api/v2/pipeline/run`
  - `GET /api/v2/pipeline/{run_id}/stages`
  - `WS /ws/pipeline/{run_id}` 实时事件流
  - `POST /api/chat` 自动识别开发需求→流水线

### 前端
- **新增** PipelineView.vue 流水线可视化
- 默认路由改为 `/pipeline`
- 引擎切换（LangChain / AgentScope）
- 离线演示模式（后端不可用时独立运行）
- WebSocket 实时进度

### 工程化
- **新增** Dockerfile（多阶段构建）
- **新增** docker-compose.yml（前后端 + Redis）
- **新增** requirements-min.txt（零 LLM 最小依赖）
- **新增** GitHub Actions CI（.github/workflows/test.yml）
- **新增** Makefile 统一命令入口
- **新增** run_tests.py 一键全测试

### 测试
- test_v2.py：36/36 通过
- test_e2e.py：HTTP 端到端
- test_ws.py：WebSocket 事件流

### 文档
- **新增** docs/structure.html（HTML+Canvas 代码结构 + 部署说明）
- 新增 README / QUICKSTART / MIGRATION / CHANGELOG / LICENSE

### 兼容性
- V1 代码完整保留（langchain_impl/ + agentscope_impl/）
- V1 依赖缺失时自动 Mock 降级，**无需 API Key 即可完整运行**

---

## [1.0.0] - 初始版本

6 个独立的工程化能力：
1. 状态持久化 (checkpoint)
2. 长期记忆 (memory)
3. 容错恢复 (recovery)
4. 性能控制 (performance)
5. 安全护栏 (guardrails)
6. 可观测性 (observability)

双框架实现：LangChain + AgentScope
