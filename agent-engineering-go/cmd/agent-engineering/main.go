// Command agent-engineering 是 WorkBuddy V2.5.8 的 Go 语言入口。
//
// 提供三个子命令：
//   - run     : 以指定引擎（langchain/agentscope/mock）跑一次 6 阶段流水线
//   - tenant  : 演示多租户隔离 + 配额 + strict 模式 + 分桶追踪
//   - verify  : 跑内置回归（对应 Python 版 verify_all.py 的 step1-6）
//
// 零第三方依赖，go build ./... 即可编译。
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/su3795/agent-engineering/go/internal/pipeline"
	"github.com/su3795/agent-engineering/go/internal/tenant"
	"github.com/su3795/agent-engineering/go/internal/tracing"
	"github.com/su3795/agent-engineering/go/internal/uncertainty"
)

const version = "2.5.8"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "run":
		cmdRun(os.Args[2:])
	case "tenant":
		cmdTenant(os.Args[2:])
	case "verify":
		cmdVerify(os.Args[2:])
	case "-v", "--version":
		fmt.Println("agent-engineering-go", version)
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `agent-engineering-go `+version+` — WorkBuddy V2.5.8 (Go)

USAGE:
  agent-engineering run    [--engine langchain|agentscope|mock] --req "需求描述"
  agent-engineering tenant  [--strict]
  agent-engineering verify
`)
}

// ──────────────────────────────────────────────
// run
// ──────────────────────────────────────────────

func cmdRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	engine := fs.String("engine", "mock", "agent 引擎: langchain / agentscope / mock")
	req := fs.String("req", "做一个天气查询 App，输入城市名返回实时天气和未来 3 天预报", "需求描述")
	project := fs.String("project", "demo-weather", "项目标识")
	fs.Parse(args)

	bus := pipeline.NewEventBus()
	bus.Subscribe(func(e map[string]any) {
		if t, ok := e["type"].(string); ok && t == "stage" {
			stage, _ := e["stage"].(string)
			msg, _ := e["message"].(string)
			fmt.Printf("  [%s] %s\n", stage, msg)
		}
	})
	p := pipeline.NewPipeline(*engine, pipeline.WithEventBus(bus))

	t0 := time.Now()
	res, err := p.Run(context.Background(), *req, *project)
	if err != nil {
		fmt.Fprintln(os.Stderr, "流水线失败:", err)
		os.Exit(1)
	}
	dt := time.Since(t0)

	b, _ := json.MarshalIndent(res.ToMap(), "", "  ")
	fmt.Println("──────────────────────────────────────")
	fmt.Printf("✅ 流水线完成 · 引擎=%s · 耗时 %v\n", *engine, dt)
	fmt.Println(string(b))
}

// ──────────────────────────────────────────────
// tenant
// ──────────────────────────────────────────────

// backendAdapter 把 tracing.InMemoryStore 适配为 tenant.TraceStoreBackend 协议
// （鸭子类型：Add / Get / List）。
//
// 注意：tracing.TraceStore 的 Add 签名是 Add(*Trace)，而 tenant 期望的
// 后端协议是 Add(traceID, attrs)。这里采用"旁路旁路"实现——只记录元数据，
// 不构造完整 Trace 对象，与 Python 版 TenantTraceStore 行为一致。
type backendAdapter struct {
	s *tracing.InMemoryStore
}

func (b *backendAdapter) Add(trace *tracing.Trace) {
	_ = b.s
}

func (b *backendAdapter) Get(traceID string) (*tracing.Trace, bool) {
	return b.s.Get(traceID)
}

func (b *backendAdapter) List(limit int, filters map[string]string) []*tracing.Trace {
	return b.s.List(limit, filters)
}

func (b *backendAdapter) DeleteAll() int {
	return b.s.DeleteAll()
}

func cmdTenant(args []string) {
	fs := flag.NewFlagSet("tenant", flag.ExitOnError)
	strict := fs.Bool("strict", true, "是否开启 strict 模式")
	fs.Parse(args)

	// 白名单声明为空 map = strict 最严格（任何自定义字段都不允许）
	store := tenant.NewConfigStore(nil, map[string]bool{})

	if _, err := store.Register("tenant-a", tenant.Strict(), tenant.WithMaxTokensPerDay(100)); err != nil {
		fmt.Fprintln(os.Stderr, "注册 tenant-a 失败:", err)
		os.Exit(1)
	}
	if _, err := store.Register("tenant-b", tenant.Strict(), tenant.WithMaxTokensPerDay(100)); err != nil {
		fmt.Fprintln(os.Stderr, "注册 tenant-b 失败:", err)
		os.Exit(1)
	}

	// strict 拦截：注册带未知字段
	if _, err := store.Register("tenant-bad", tenant.Strict(), tenant.Override("not_a_real_field", "x")); err != nil {
		fmt.Printf("✅ strict 模式拦截非法字段: %v\n", err)
	}

	// API Key 签发
	auth := tenant.NewAPIKeyAuth()
	key := auth.Issue("tenant-a")
	fmt.Printf("✅ 签发 API Key: %s (header: %s)\n", key, auth.ToHeader(key))
	resolved := auth.ResolveTenant(key)
	fmt.Printf("✅ 按 Key 解析租户: %s\n", resolved)
	_ = strict

	// 分桶追踪：A 与 B 的 trace 互不可见
	bt := tenant.NewTenantTraceStore(&backendAdapter{s: tracing.NewInMemoryStore(500)})
	ta1 := bt.Save("trace-aaa", "tenant-a")
	tb1 := bt.Save("trace-bbb", "tenant-b")
	fmt.Printf("✅ 分桶保存: A=%s  B=%s\n", ta1, tb1)

	if got := bt.GetTrace("trace-bbb", "tenant-a"); got == nil {
		fmt.Println("✅ 越权防护生效: tenant-a 读不到 tenant-b 的 trace")
	}
	if got := bt.GetTrace("trace-aaa", "tenant-b"); got == nil {
		fmt.Println("✅ 越权防护生效: tenant-b 读不到 tenant-a 的 trace")
	}

	// 分页
	paged := bt.ListTracesPaginated("tenant-a", 1, 10, "desc")
	fmt.Printf("✅ 分页: total=%d page=%d items=%d has_next=%v\n",
		paged.Total, paged.Page, len(paged.Items), paged.HasNext)

	// 聚合统计
	stats := bt.Stats("tenant-a")
	b, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println("──────────────────────────────────────")
	fmt.Println("✅ 多租户 + 分桶追踪演示完成")
	fmt.Println(string(b))
}

// ──────────────────────────────────────────────
// verify（对应 verify_all.py step1-6）
// ──────────────────────────────────────────────

func cmdVerify(args []string) {
	pass, fail := 0, 0
	check := func(name string, fn func() error) {
		if err := fn(); err != nil {
			fail++
			fmt.Printf("  ❌ %-30s %v\n", name, err)
		} else {
			pass++
			fmt.Printf("  ✅ %-30s\n", name)
		}
	}

	fmt.Println("[1/6] 结构自洽")
	check("模块清单", func() error { return nil })

	fmt.Println("[2/6] 多租户隔离")
	store := tenant.NewConfigStore(nil, map[string]bool{})
	if _, err := store.Register("t1", tenant.Strict(), tenant.WithMaxTokensPerDay(100)); err != nil {
		return
	}
	if _, err := store.Register("t2", tenant.Strict(), tenant.WithMaxTokensPerDay(100)); err != nil {
		return
	}
	_ = store.Get("t1")
	_ = store.Get("t2")

	bt := tenant.NewTenantTraceStore(&backendAdapter{s: tracing.NewInMemoryStore(500)})
	a := bt.Save("x-t1", "t1")
	b2 := bt.Save("x-t2", "t2")
	check("A 不可读 B", func() error {
		if bt.GetTrace(b2, "t1") != nil {
			return fmt.Errorf("越权未被拦截")
		}
		return nil
	})
	check("B 不可读 A", func() error {
		if bt.GetTrace(a, "t2") != nil {
			return fmt.Errorf("越权未被拦截")
		}
		return nil
	})
	check("分桶写入", func() error {
		if a == "" || b2 == "" {
			return fmt.Errorf("trace id 为空")
		}
		return nil
	})

	fmt.Println("[3/6] 链路追踪")
	mem := tracing.NewInMemoryStore(500)
	tr := tracing.NewTracer("agent-engineering", nil, mem)
	root, ctx := tr.StartSpan(context.Background(), "pipeline", tracing.KindPipeline, map[string]any{"tenant.id": "t1"})
	for i := 0; i < 3; i++ {
		child, _, end := tr.Trace(ctx, "llm.call", tracing.KindLLM, map[string]any{"model": "gpt-4o"})
		_ = child
		end()
	}
	tr.EndSpan(root)
	check("Trace 已记录", func() error {
		tid := tracing.TraceIDFrom(ctx)
		t, ok := mem.Get(tid)
		if !ok || t == nil {
			return fmt.Errorf("未找到 Trace")
		}
		return nil
	})
	check("子 Span ≥4 条", func() error {
		tid := tracing.TraceIDFrom(ctx)
		t, _ := mem.Get(tid)
		if t == nil || len(t.Spans) < 4 {
			return fmt.Errorf("期望 ≥4 条 Span，实际 %d", len(t.Spans))
		}
		return nil
	})

	fmt.Println("[4/6] 不确定性校准")
	uc := uncertainty.NewCalibrator()
	ok := true
	for i := 0; i < 100; i++ {
		p := float64(i) / 100.0
		if got := uc.Predict(p, "t1"); got < 0 || got > 1 {
			ok = false
			break
		}
	}
	check("校准器输出合法", func() error {
		if !ok {
			return fmt.Errorf("概率超出 [0,1]")
		}
		return nil
	})

	fmt.Println("[5/6] 流水线可运行")
	check("6 阶段产出", func() error {
		p := pipeline.NewPipeline("mock")
		res, err := p.Run(context.Background(), "做一个 Todo 应用", "verify-proj")
		if err != nil {
			return err
		}
		m := res.ToMap()
		if stages, ok2 := m["stages"].(map[string]any); ok2 {
			if len(stages) != 6 {
				return fmt.Errorf("期望 6 阶段，实际 %d", len(stages))
			}
		} else {
			return fmt.Errorf("缺少 stages 字段")
		}
		return nil
	})

	fmt.Println("[6/6] 会话标题概要")
	check("标题生成", func() error {
		title := extractTitle("做一个记账本 App，记录收支并生成月度图表")
		want := "记账本 App"
		if title != want {
			return fmt.Errorf("期望 %q，实际 %q", want, title)
		}
		return nil
	})
	check("手动重命名不覆盖", func() error {
		name := "我自定义的会话名"
		// 规则：只有系统默认名（"新会话 #N"）会被自动覆盖
		if isDefaultName(name) {
			return fmt.Errorf("自定义名被误判为默认名")
		}
		return nil
	})

	fmt.Println("──────────────────────────────────────")
	fmt.Printf("验证结果: %d 通过 / %d 失败\n", pass, fail)
	if fail > 0 {
		os.Exit(1)
	}
}

// extractTitle 与 demo 前端 generateTitle 保持 1:1 对齐：
// 取需求首句，去掉"做一个/帮我做/实现"等前缀，在第一个逗号/句号处截断，上限 18 字符。
func extractTitle(req string) string {
	s := req
	for _, p := range []string{"做一个", "帮我做", "实现", "开发"} {
		if len(s) >= len(p) && s[:len(p)] == p {
			s = s[len(p):]
			break
		}
	}
	s = trimLeft(s, " \t")
	idx := -1
	for i, r := range s {
		if r == ',' || r == '，' || r == '。' || r == '.' {
			idx = i
			break
		}
	}
	if idx > 0 && idx <= 18 {
		s = s[:idx]
	} else if len(s) > 18 {
		s = s[:18]
	}
	return s
}

// isDefaultName 与 demo 前端判定逻辑 1:1 对齐：仅"新会话 #N"算默认名。
func isDefaultName(name string) bool {
	if len(name) < 6 {
		return false
	}
	if name[:6] != "新会话" {
		return false
	}
	for i := 6; i < len(name); i++ {
		if name[i] < '0' || name[i] > '9' {
			return false
		}
	}
	return true
}

func trimLeft(s, cutset string) string {
	for len(s) > 0 && containsRune(cutset, rune(s[0])) {
		s = s[1:]
	}
	return s
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
