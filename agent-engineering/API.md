# API 参考 (V2)

Base URL: `http://localhost:8000`

---

## REST

### `GET /`
服务健康检查 + 版本信息。

```json
{ "name": "Agent Engineering Platform", "version": "2.0.0" }
```

### `GET /api/capabilities`
6 大工程化能力清单 + V2 用途说明。

### `POST /api/v2/pipeline/run`
启动一条流水线。

**请求体：**
```json
{
  "requirement": "做一个待办事项 App",
  "engine": "langchain",
  "project_id": "proj-1"
}
```
- `engine`: `"langchain"` (默认) | `"agentscope"`

**响应：**
```json
{ "run_id": "run-527e9c97", "status": "running", "ws_url": "/ws/pipeline/run-527e9c97" }
```

### `GET /api/v2/pipeline/{run_id}`
查询流水线整体状态（含 stages 摘要）。

### `GET /api/v2/pipeline/{run_id}/stages`
各阶段详情，供前端流水线可视化。

**响应：**
```json
{
  "run_id": "run-xxx",
  "engine": "langchain",
  "status": "finished",
  "stages": [
    {
      "stage": "coding",
      "name": "coding",
      "agent": "Coder Agent (×3 并行)",
      "status": "success",
      "duration_ms": 320,
      "capabilities_used": [1, 3],
      "events": [...],
      "artifact": { "files": [...], "tests_passed": 12 }
    }
  ]
}
```

### `POST /api/chat`
智能聊天：检测到开发需求自动走流水线，否则普通对话。

---

## WebSocket

### `WS /ws/pipeline/{run_id}`

连接后立即收到：
```json
{ "type": "connected", "run_id": "run-xxx" }
```

随后推送事件流：

| type | payload |
|------|---------|
| `pipeline.start` | `{run_id, requirement, engine}` |
| `stage.start` | `{stage, agent}` |
| `stage.event` | `{stage, message, level}` |
| `agent.thought` | `{agent, thought}` |
| `agent.action` | `{agent, action}` |
| `permission.request` | `{agent, action, reason}` |
| `stage.end` | `{stage, status, duration_ms, artifact}` |
| `pipeline.end` | `{run_id, status}` |

**前端示例（PipelineView.vue）：**
```js
const ws = new WebSocket(`${VITE_WS_URL}/ws/pipeline/${runId}`)
ws.onmessage = (e) => {
  const evt = JSON.parse(e.data)
  // 根据 evt.type 更新 UI
}
```

---

## 错误码

| 状态码 | 含义 |
|--------|------|
| 200 | 成功 |
| 400 | 请求体校验失败 |
| 404 | run_id 不存在 |
| 500 | 服务器内部错误（含流水线异常） |

错误响应格式：
```json
{ "detail": "错误描述" }
```
