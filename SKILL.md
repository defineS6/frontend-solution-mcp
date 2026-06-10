---
name: frontend-mcp-uiux-collab
description: "在前端 UI/UX 任务中调用 mcp__frontend_mcp__frontend_solution 生成方案、组件拆分、样式建议和代码草案，再由 Codex 进行工程化实现、重构与验证。用于页面布局、组件设计、交互状态、动效、响应式、可访问性和设计系统优化；不用于纯后端逻辑、算法或数据处理任务。"
---

# Frontend MCP UI/UX 协作

## 目标

- 先使用 `frontend_mcp` 快速产出 UI/UX 方案、组件拆分、代码片段或小范围 Unified Diff 草案。
- 再由 Codex 完成可维护实现，确保符合项目代码风格、工程约束与质量标准。
- 明确禁止把 `frontend_mcp` 输出直接视为最终代码；它只提供草案，最终落地由 Codex 完成。

## 执行边界

- 只在前端 UI/UX 相关需求触发本技能。
- 遇到纯后端、算法、数据处理任务时，不使用本技能流程。
- 不要让 `frontend_mcp` 直接修改项目文件；只请求方案、片段或 diff 草案。
- 不要把本地 `.env` 文件内容作为上下文发给 `frontend_mcp`；代码逻辑中仍可保留读取环境变量的写法。

## 标准流程

### 1) 调用 frontend_mcp 获取草案

始终调用 `mcp__frontend_mcp__frontend_solution`。

默认参数：
- `return_all_messages=false`
- `SESSION_ID`：同一功能迭代时复用上次返回的值；新功能或会话污染时留空新建
- `model`：仅在用户明确指定或环境已有固定配置时传入

使用以下输入模板（`# Task` 必须英文）：

    # Task (English)
    [Describe the UI/UX task, constraints, expected output format, and acceptance criteria.]

    # User Original Request
    """
    [粘贴用户原话]
    """

    # Context
    - Project Root: [前端项目根目录绝对路径]
    - Tech Stack: [React/Vue/Vanilla/...]
    - Existing Components: [相关组件路径或关键代码片段]
    - Design Constraints: [颜色/字体/间距/动效/a11y/响应式限制]

调用后立即记录返回的 `SESSION_ID`，用于同一功能后续迭代。

### 2) 防超时调用规范

- 默认关闭冗长返回；不需要调试轨迹时，不启用 `return_all_messages`。
- `# Task (English)` 保持简短，用 1-3 句描述目标、约束和验收。
- `# Context` 只提供必要信息，不粘贴大段无关代码。
- 单次优先只解决一个子问题，例如仅布局、仅动效、仅 a11y。
- 每个相关文件只提供关键片段，建议 40-120 行。
- 优先提供接口、符号、样式变量和组件结构，不提供整文件。
- 大任务拆分为多次调用：设计决策 -> 代码片段 -> diff 草案。

推荐输入模板：

    # Task (English)
    [One focused subtask only. Keep it concise.]

    # User Original Request
    """
    [粘贴用户原话]
    """

    # Context (Minimal)
    - Project Root: [前端项目根目录绝对路径]
    - Tech Stack: [React/Vue/Vanilla/...]
    - Files In Scope: [仅列 1-3 个关键文件]
    - Key Constraints: [最多 3-5 条]

    # Output Needed
    1) [简短设计决策]
    2) [最小代码片段或最小 diff]
    3) [验收检查点]

### 3) 整理 frontend_mcp 输出

优先提取以下内容：
- 前端代码片段，必须覆盖 a11y、响应式、动效或交互状态要点。
- 对已有代码的小变更 Unified Diff。
- 关键设计决策与取舍说明。

清理输出格式：
- 删除 Markdown 代码围栏，去掉三反引号标记。
- 修复路径、变量名、组件名与仓库实际结构不一致的问题。
- 回收不符合现有设计系统的颜色、间距、字体和阴影。
- 过滤无法落地的占位、TODO、伪依赖和不存在的组件 API。

### 4) 工程化实现与验证

将草案转化为可维护实现：
- 按项目分层组织组件、样式和状态逻辑。
- 复用现有组件、工具函数、样式变量和设计系统。
- 补齐 `loading`、`empty`、`error`、`disabled`、`hover`、`active`、`focus` 等状态。
- 验证可访问性，包括语义标签、键盘可达、焦点管理和 `aria` 属性。
- 验证响应式与跨端展示，避免文本溢出、布局跳动和内容重叠。
- 执行项目已有校验命令；是否运行构建按项目约束和用户要求决定，不要无脑每次构建。

## 迭代策略

复用同一个 `SESSION_ID`：
- 微调样式、交互、间距、层级或文案。
- 同一功能的增量优化。
- 同一组件不同状态，例如 hover、active、disabled、focus。

新开会话：
- 切换到不同功能模块。
- 需求目标或设计方向明显变化。
- 旧会话上下文污染或输出质量持续下降。

## 错误处理

`frontend_mcp` 调用失败时：
1. 检查 MCP 服务状态、工具名和参数结构。
2. 走降级方案：基于现有设计系统手动实现。
3. 记录失败场景，包括任务类型、报错和输入模板，便于后续优化。

`frontend_mcp` 超时时：
1. 第一轮重试：缩短 `# Task` 与 `# Context`，只保留一个子任务。
2. 第二轮重试：进一步拆分输出要求，只请求“设计决策”或“单文件代码”。
3. 连续两次超时后，停止继续扩大请求，改走手工工程化实现，并在交付中注明降级原因。

输出质量不达标时：
1. 补充更具体的最小上下文后重试。
2. 将大任务拆成布局、样式、动效、a11y 等更小子问题。
3. 明确禁止项和验收标准后再次生成。

`SESSION_ID` 失效时：
1. 新建会话。
2. 在新输入中先总结旧会话关键决策，再继续迭代。

## 交付要求

- 始终说明：`frontend_mcp` 产物是草案，最终代码由 Codex 工程化落地。
- 对用户展示改动时，优先给出改动点、原因、验证方式、风险与回滚点。
- 保持改动最小化与可追溯，避免无关重构。
