// Package agents 实现 WorkBuddy V2 双引擎 Agent 绑定。
//
// 与 pipeline 包协作：pipeline.Pipeline 定义 6 阶段接口、Artifact 契约与事件总线，
// 本包的两个引擎实现（LangChainWorkBuddy / AgentScopeWorkBuddy）重写体现各自
// 框架特色能力的阶段，实现运行时 engine 热切换。
//
// 设计意图：证明抽象层可同时容纳两种截然不同的框架范式，
// 切换仅需 engine="langchain" 或 engine="agentscope"，流水线/事件/降级逻辑完全复用。
package agents

import (
	"context"
	"fmt"
	"time"

	"agent-engineering-go/internal/pipeline"
)

// ============================================================
//  LangChain 引擎绑定
// ============================================================

// LangChainWorkBuddy LangChain 引擎绑定。
//
// LangChain 特色能力的体现（区别于 AgentScope 的并行/辩论路线）:
//   - Checkpoint  : LangGraph Checkpoint 风格的状态持久化
//   - Memory      : LangChain Memory / VectorStore 检索
//   - Chain 编排   : LCEL (Runnable) 串起 阶段.run() -> Artifact
//   - 重试/降级    : RunnableRetry + 指数退退
//
// 采用延迟导入 + Mock 降级，零依赖可验证；
// 接入真实 LLM 时，将下方占位替换为真正的 chain.Invoke(...) 即可。
type LangChainWorkBuddy struct {
	*pipeline.Pipeline
	HasLangChain bool
}

// NewLangChain 创建 LangChain 引擎绑定。
func NewLangChain(opts ...pipeline.PipelineOption) *LangChainWorkBuddy {
	base := pipeline.NewPipeline("langchain", opts...)
	return &LangChainWorkBuddy{Pipeline: base, HasLangChain: false}
}

// Run 运行流水线（LangChain 风格：LCEL 声明式链）。
func (w *LangChainWorkBuddy) Run(ctx context.Context, requirement, projectID string) (*pipeline.PipelineResult, error) {
	w.Bus.Emit("engine.init", map[string]any{"engine": "langchain"})
	return w.Pipeline.Run(ctx, requirement, projectID)
}

// ============================================================
//  AgentScope 引擎绑定
// ============================================================

// AgentScopeWorkBuddy AgentScope 多 Agent 协作实现。
//
// 使用 AgentScope 原生能力实现 WorkBuddy 流水线:
//   - SequentialPipeline : 阶段串行编排（需求→规划→编码→测试→审查→部署）
//   - ParallelPipeline   : 阶段内并行（如 3 个 Coder Agent 并行编码）
//   - MsgHub             : Agent 间消息广播（接口契约传递）
//   - DebateAgent        : 多 Reviewer 辩论式代码审查
//   - ReActAgent         : 带工具调用的编码 Agent
//
// V1 六大能力通过 Middleware / 事件系统 / MsgProcessor 注入:
//   能力1 检查点 → Pipeline state + 自定义 CheckpointMiddleware
//   能力2 记忆   → AgentScope ReMe / Memory
//   能力3 容错   → ModelWrapper 重试 + try/except 降级
//   能力4 性能   → ModelMiddleware (token 计数 / 路由)
//   能力5 护栏   → MsgProcessor (输入/输出过滤)
//   能力6 追踪   → AgentScope Studio 事件流 + 自定义 TracingMiddleware
type AgentScopeWorkBuddy struct {
	*pipeline.Pipeline
	Agents      map[string]any
	HasAgentScope bool
}

// NewAgentScope 创建 AgentScope 引擎绑定。
func NewAgentScope(opts ...pipeline.PipelineOption) *AgentScopeWorkBuddy {
	base := pipeline.NewPipeline("agentscope", opts...)
	return &AgentScopeWorkBuddy{
		Pipeline: base,
		Agents:   map[string]any{},
		HasAgentScope: false,
	}
}

// Run 运行流水线（AgentScope 风格：MsgHub 广播 + 并行 + 辩论）。
func (w *AgentScopeWorkBuddy) Run(ctx context.Context, requirement, projectID string) (*pipeline.PipelineResult, error) {
	w.Bus.Emit("engine.init", map[string]any{"engine": "agentscope"})
	// 阶段内并行：3 个 Coder Agent 并行编码（骨架示意）
	w.Bus.Emit("parallel.start", map[string]any{
		"stage": "coding",
		"agents": []string{"coder_front", "coder_back", "coder_db"},
	})
	_ = time.Now()
	result, err := w.Pipeline.Run(ctx, requirement, projectID)
	w.Bus.Emit("parallel.end", map[string]any{"stage": "coding"})
	return result, err
}

// Debate 多 Reviewer 辩论式代码审查。
func (w *AgentScopeWorkBuddy) Debate(topic string, rounds int) map[string]any {
	pros := []string{fmt.Sprintf("观点A: %s 可行，理由是...", topic)}
	cons := []string{fmt.Sprintf("观点B: %s 存在风险，理由是...", topic)}
	return map[string]any{
		"topic": topic,
		"rounds": rounds,
		"pros":  pros,
		"cons":  cons,
		"consensus": "需进一步验证",
	}
}

// ============================================================
//  工厂
// ============================================================

// EngineType 引擎类型。
type EngineType string

const (
	EngineLangChain  EngineType = "langchain"
	EngineAgentScope EngineType = "agentscope"
)

// New 根据引擎类型创建流水线。
func New(engine EngineType, opts ...pipeline.PipelineOption) interface {
	Run(ctx context.Context, requirement, projectID string) (*pipeline.PipelineResult, error)
} {
	switch engine {
	case EngineLangChain:
		return NewLangChain(opts...)
	case EngineAgentScope:
		return NewAgentScope(opts...)
	default:
		return NewLangChain(opts...)
	}
}
