# Claude Code Stats — Data Points Reference

All data lives at `~/.claude/` on macOS (`/Users/{username}/.claude/`).  
Real path confirmed: `/Users/graycup/.claude/`

Operating System	~/.claude resolves to
macOS	/Users/<username>/.claude
Linux	/home/<username>/.claude
WSL (Ubuntu, Debian, etc.)	/home/<username>/.claude
Windows (PowerShell/CMD)	C:\Users\<username>\.claude

---

## Directory Structure

```
~/.claude/
├── stats-cache.json          ← pre-aggregated stats (START HERE)
├── history.jsonl             ← slash command history
├── projects/                 ← per-project conversation logs (richest source)
├── sessions/                 ← session metadata per PID
├── settings.json             ← user config, permissions, theme
├── mcp-needs-auth-cache.json ← MCP server auth status
├── plugins/                  ← plugin ecosystem data
├── telemetry/                ← failed telemetry events (env/error info)
├── ide/                      ← active IDE session locks
├── file-history/             ← file modification tracking (often empty)
├── session-env/              ← env var snapshots per session (often empty)
├── shell-snapshots/          ← shell function dumps (debugging only)
├── backups/                  ← .claude.json config snapshots
├── cache/                    ← static content (changelog.md)
├── debug/                    ← runtime debug logs
└── downloads/                ← downloaded resources
```

---

## 1. `stats-cache.json` — Pre-aggregated Analytics

**Schema version:** `2`  
**Path:** `~/.claude/stats-cache.json`  
**When updated:** On Claude Code shutdown/cleanup

This is the fastest data source — no parsing needed, just read and display.

### Fields

```jsonc
{
  "version": 2,
  "lastComputedDate": "2026-04-24",   // ISO date of last recompute

  // Daily activity rollup — one entry per active day
  "dailyActivity": [
    {
      "date": "2026-02-01",
      "messageCount": 877,       // total messages (user + assistant)
      "sessionCount": 4,         // Claude Code sessions started
      "toolCallCount": 182       // total tool invocations (Bash, Read, Edit, etc.)
    }
  ],

  // Token consumption by day + model
  "dailyModelTokens": [
    {
      "date": "2026-02-01",
      "tokensByModel": {
        "claude-sonnet-4-6": 450000,
        "claude-opus-4-5-20251101": 12000
      }
    }
  ],

  // Aggregate token stats per model (entire history)
  "modelUsage": {
    "claude-sonnet-4-6": {
      "inputTokens": 111828,
      "outputTokens": 4604387,
      "cacheReadInputTokens": 363826598,    // tokens served from cache
      "cacheCreationInputTokens": 19388282, // tokens written to cache
      "webSearchRequests": 0,
      "costUSD": 0,       // NOTE: always 0 — compute from tokens (see pricing section)
      "contextWindow": 0,
      "maxOutputTokens": 0
    }
    // also: claude-opus-4-5-20251101, claude-opus-4-6, claude-sonnet-4-5-20250929
  },

  "totalSessions": 133,
  "totalMessages": 17458,

  "longestSession": {
    "sessionId": "7a7c485d-...",
    "duration": 1673243041,   // milliseconds (can be absurd if session left open — clamp)
    "messageCount": 302,
    "timestamp": "2026-03-08T14:03:01.056Z"
  },

  "firstSessionDate": "2026-02-01T13:28:57.834Z",

  // Sessions started per hour of day (UTC — convert to local for display)
  "hourCounts": {
    "0": 5, "1": 2, "2": 6, ..., "23": 2
  },

  "totalSpeculationTimeSavedMs": 0  // always 0 currently
}
```

### Real Data (this user)
- **133 sessions**, **17,458 messages**, active since **2026-02-01**
- Models: `claude-sonnet-4-6` primary, plus opus-4-5, opus-4-6, sonnet-4-5
- Peak hours: 17–20:00 (evening heavy user)
- `costUSD` is always `0` — must compute from tokens

### Dashboard Widgets
| Widget | Fields Used |
|---|---|
| KPI cards | `totalSessions`, `totalMessages`, `firstSessionDate` |
| Daily activity line chart | `dailyActivity[].{date, messageCount, sessionCount, toolCallCount}` |
| Token trend by model | `dailyModelTokens[].{date, tokensByModel}` |
| Model usage pie | `modelUsage` keys + token sums |
| Hour of day heatmap | `hourCounts` |
| Longest session card | `longestSession.{messageCount, duration}` |
| Cost over time (computed) | `dailyModelTokens` × pricing table |

---

## 2. `projects/{project-id}/{session-uuid}.jsonl` — Conversation Logs

**Path:** `~/.claude/projects/`  
**Format:** One directory per project, one JSONL file per session  
**Project dir naming:** `/` replaced with `-`, leading `-` = `/`  
  e.g. `/Users/graycup/Documents/GitHub/farms-directory` → `-Users-graycup-Documents-GitHub-farms-directory`

**44 projects** in the example dir. Aggregate stats from full parse:
- `claude-sonnet-4-6`: 12,427 assistant messages, `<synthetic>`: 11
- Top projects by messages: opend2c (1427), graybulk (1039), aapka-app (914)
- Top projects by tokens: opend2c (109M), nvault-app-clicksy (84M), graybulk (63M)
- Tool calls: Bash (2111), Edit (2107), Read (1683), Write (896), TodoWrite (174), WebFetch (48), ToolSearch (39), Agent (17), Skill (13), AskUserQuestion (9), Monitor (4)
- 114 of ~130 sessions had extended thinking blocks
- 210 tool result errors across all sessions
- Avg session duration: 72 min (range: 0.3 – 571 min, excluding stale outliers)
- Session distribution: <5min (14), 5–15min (14), 15–60min (37), >1hr (25)

---

### Message Types (all top-level types in JSONL)

```
assistant          12,438  ← Claude responses with token usage
user                8,033  ← User prompts + tool results
ai-title            2,367  ← AI-generated session name
file-history-snapshot 2,270 ← File edit tracking snapshots
last-prompt         1,982  ← Last user prompt state
queue-operation     1,860  ← Prompt queue enqueue/dequeue
attachment          1,622  ← Context injected by Claude Code
mode                  281  ← Session mode changes
system                 47  ← API errors / connection issues
```

---

### Type: `assistant`

```jsonc
{
  "type": "assistant",
  "uuid": "...",
  "parentUuid": "...",     // links conversation thread
  "sessionId": "...",
  "timestamp": "2026-04-21T18:30:05.000Z",
  "requestId": "req_...",
  "message": {
    "model": "claude-sonnet-4-6",
    "role": "assistant",
    "stop_reason": "tool_use",  // "end_turn" | "tool_use" | "stop_sequence"
    "content": [
      { "type": "thinking", "thinking": "..." },  // extended thinking block
      { "type": "text", "text": "response text" },
      {
        "type": "tool_use",
        "id": "toolu_...",
        "name": "Bash",   // Bash | Read | Edit | Write | TodoWrite | WebFetch |
                          // ToolSearch | Agent | Skill | AskUserQuestion | Monitor
        "input": { "command": "npm run build" }
      }
    ],
    "usage": {
      "input_tokens": 1234,
      "output_tokens": 567,
      "cache_read_input_tokens": 89000,
      "cache_creation_input_tokens": 5000,
      "server_tool_use": {
        "web_search_requests": 0,
        "web_fetch_requests": 0
      },
      "service_tier": "standard",
      "cache_creation": {
        "ephemeral_5m_input_tokens": 0,
        "ephemeral_1h_input_tokens": 5000
      },
      "inference_geo": "us"
    }
  }
}
```

**Derivable:** cost, cache hit rate, tool call counts, thinking usage, response latency, stop reason distribution.

---

### Type: `user` (prompt)

```jsonc
{
  "type": "user",
  "uuid": "...",
  "parentUuid": null,          // null = first message in thread
  "isSidechain": false,
  "sessionId": "...",
  "timestamp": "2026-04-21T18:30:00.000Z",
  "promptId": "...",
  "permissionMode": "acceptEdits",  // "acceptEdits" | "default" | "bypassPermissions"
  "userType": "external",
  "entrypoint": "claude-vscode",    // "claude-vscode" | "cli"
  "cwd": "/Users/graycup/Documents/GitHub/farms-directory",
  "gitBranch": "main",
  "version": "2.1.145",        // Claude Code version at time of message
  "message": {
    "role": "user",
    "content": [
      { "type": "text", "text": "the actual user prompt" }
    ]
  }
}
```

**Versions seen:** 2.1.145 (320 msgs), 2.1.168 (143), 2.1.160 (90), 2.1.159 (87), 2.1.143 (85)  
**Branches:** main (1187), HEAD (32), landing (3), pr/1 (2), prod2 (2)  
**Permission modes:** acceptEdits (1168), unknown (58)  
**Entrypoints:** 100% `claude-vscode` in example data

---

### Type: `user` (tool result)

```jsonc
{
  "type": "user",
  "message": {
    "role": "user",
    "content": [
      {
        "type": "tool_result",
        "tool_use_id": "toolu_...",
        "content": "stdout output here",
        "is_error": false    // true = tool failed (210 errors found in example)
      }
    ]
  },
  "toolUseResult": {
    "stdout": "...",
    "stderr": "",
    "interrupted": false,
    "returnCodeInterpretation": "success"  // or "No matches found", etc.
  }
}
```

**Derivable:** tool error rate, which tools fail most, interrupted commands count.

---

### Type: `ai-title`

AI-generated human-readable session title. Great for session browser display.

```jsonc
{
  "type": "ai-title",
  "sessionId": "ac435b43-...",
  "aiTitle": "Set up SEO for coffee brand websites"
}
```

Sample titles from example data:
- "Set up SEO for multiple Bulkchai domain properties"
- "Implement payment gateway integration"
- "Fix TypeScript errors in storefront"
- "Build order management dashboard"

---

### Type: `last-prompt`

The most recent user prompt text in the session — useful as a session preview/summary.

```jsonc
{
  "type": "last-prompt",
  "lastPrompt": "https://developers.google.com/search/docs/...",  // truncated with …
  "leafUuid": "6c62494f-...",   // UUID of the message this prompt belongs to
  "sessionId": "ac435b43-..."
}
```

**Dashboard use:** Session list preview, "resume session" context.

---

### Type: `file-history-snapshot`

Tracks which files Claude has backed up in case of undo. Shows the most-edited files per session and version history.

```jsonc
{
  "type": "file-history-snapshot",
  "messageId": "82d923ca-...",
  "snapshot": {
    "messageId": "82d923ca-...",
    "trackedFileBackups": {
      "src/app/page.tsx": {
        "backupFileName": "53f1fc59531cac07@v1",
        "version": 14,           // how many times this file was backed up
        "backupTime": "2026-05-28T22:13:49.207Z"
      },
      "package.json": {
        "backupFileName": "ee656db8fbb5794d@v7",
        "version": 7,
        "backupTime": "2026-05-28T22:13:54.230Z"
      }
    },
    "timestamp": "2026-05-28T22:09:14.659Z"
  },
  "isSnapshotUpdate": true   // false = initial snapshot, true = file was added
}
```

**Top edited files from example data:**
| File | Max Version | Snapshot Count |
|---|---|---|
| `package.json` | v7 | 232 |
| `src/app/(marketing)/buy-samples/page.tsx` | v11 | 213 |
| `src/components/navbar.tsx` | v8 | 200 |
| `Cargo.toml` | v4 | 172 |
| `apps/console/src/lib/scraper-store.ts` | v11 | 148 |
| `src/app/page.tsx` | v14 | 143 |
| `src/components/footer.tsx` | v8 | 111 |
| `packages/db/schema.ts` | v3 | 105 |

**Dashboard use:** "Most edited files" heatmap, file churn tracker, undo history timeline.

---

### Type: `system` (API errors)

Connection errors and API failures recorded inline in the conversation log.

```jsonc
{
  "type": "system",
  "subtype": "api_error",
  "level": "error",
  "error": {
    "message": "Connection error.",
    "formatted": "Unable to connect to API (ECONNRESET)",
    "connection": {
      "code": "ECONNRESET",        // or "ConnectionRefused"
      "message": "The socket connection was closed unexpectedly.",
      "isSSLError": false
    },
    "isNetworkDown": false,        // true if full network outage
    "rateLimits": null             // populated when rate limited
  },
  "retryInMs": 530.76,
  "retryAttempt": 1,
  "maxRetries": 10,
  "timestamp": "2026-05-28T15:28:59.476Z",
  "uuid": "...",
  "entrypoint": "claude-vscode",
  "cwd": "/Users/graycup/Documents/GitHub/graycup-orders-main",
  "sessionId": "..."
}
```

Error codes seen: `ECONNRESET`, `ConnectionRefused`

**Dashboard use:** Reliability/uptime chart, error rate over time, network outage detection.

---

### Type: `queue-operation`

Tracks when user prompts are queued and dequeued (background prompt queue).

```jsonc
{
  "type": "queue-operation",
  "operation": "enqueue",   // or "dequeue"
  "timestamp": "2026-05-28T22:09:14.617Z",
  "sessionId": "ac435b43-..."
}
```

**Dashboard use:** Queue depth over time, prompt throughput rate.

---

### Type: `mode`

Session mode changes.

```jsonc
{
  "type": "mode",
  "mode": "normal",   // only value seen so far
  "sessionId": "ac435b43-..."
}
```

---

### Type: `attachment` — Subtypes

All have the outer shape: `{ "type": "attachment", "attachment": { "type": "...", ... } }`

#### `todo_reminder` (906 occurrences)
```jsonc
{
  "type": "todo_reminder",
  "content": [
    { "id": "1", "content": "Fix login bug", "status": "in_progress", "priority": "high" }
  ],
  "itemCount": 3
}
```
**Dashboard use:** Todo completion rate per session, task tracking velocity.

#### `hook_additional_context` (155 occurrences)
IDE diagnostics passed to Claude via hooks after each Edit.
```jsonc
{
  "type": "hook_additional_context",
  "hookName": "PostToolUse:Edit",
  "hookEvent": "PostToolUse",
  "toolUseID": "toolu_...",
  "content": [
    "<ide_diagnostics>[{
      \"filePath\": \"/Users/graycup/.../layout.tsx\",
      \"line\": 4,
      \"column\": 8,
      \"message\": \"Cannot find module '@/styles/globals.css'\",
      \"code\": \"2882\",
      \"severity\": \"Error\"
    }]</ide_diagnostics>"
  ]
}
```
**Dashboard use:** TypeScript errors Claude saw and fixed, error severity trends, which files had the most diagnostic noise.

#### `edited_text_file` (160 occurrences)
File snippet shown to Claude after an edit.
```jsonc
{
  "type": "edited_text_file",
  "filename": "/Users/graycup/.../BuyChaiClient.tsx",
  "snippet": "1\t'use client'\n2\t\nimport { useState..."
}
```
**Dashboard use:** Files Claude touched per session, language breakdown.

#### `queued_command` (59 occurrences)
User's next prompt queued while Claude is working.
```jsonc
{
  "type": "queued_command",
  "prompt": [{ "type": "text", "text": "how would i distribute the binary..." }],
  "source_uuid": "f6bfadf6-...",
  "commandMode": "prompt"
}
```
**Dashboard use:** Queue usage frequency, how often users type ahead.

#### `nested_memory` (15 occurrences)
CLAUDE.md project memory files loaded during session.
```jsonc
{
  "type": "nested_memory",
  "path": "/Users/graycup/.../CLAUDE.md",
  "displayPath": "apps/store/CLAUDE.md",
  "content": {
    "type": "Project",
    "content": "Default to using Bun instead of Node.js...",
    "contentDiffersFromDisk": false
  }
}
```
**Dashboard use:** Which projects have CLAUDE.md configured, memory file usage.

#### `date_change` (11 occurrences)
Date crossing detected mid-session (long overnight sessions).
```jsonc
{
  "type": "date_change",
  "newDate": "2026-05-29"
}
```
**Dashboard use:** Identify overnight/marathon sessions.

#### `agent_listing_delta` (8 occurrences)
Available sub-agent types loaded into session.
```jsonc
{
  "type": "agent_listing_delta",
  "addedTypes": ["claude", "claude-code-guide", "Explore", "general-purpose", "Plan", "statusline-setup"],
  "addedLines": ["- claude: Catch-all for any task..."]
}
```
**Dashboard use:** Which agent types are available/used.

#### `deferred_tools_delta` (131 occurrences)
Tools loaded dynamically during session.

#### `skill_listing` (117 occurrences)
Available skills injected into context.

#### `compact_file_reference` (4 occurrences)
```jsonc
{
  "type": "compact_file_reference",
  "filename": "/Users/graycup/.../scraper-store.ts",
  "displayPath": "apps/console/src/lib/scraper-store.ts"
}
```

#### `command_permissions` (13 occurrences)
```jsonc
{ "type": "command_permissions", "allowedTools": [] }
```

---

### What You Can Derive Per Session (summary)

| Metric | How |
|---|---|
| **Duration** | `lastMsg.timestamp − firstMsg.timestamp` |
| **Cost (USD)** | Sum `usage.*tokens` across all assistant msgs × pricing |
| **Cache hit rate** | `cache_read / (input + cache_read)` — typically 95%+ |
| **Tool breakdown** | Count `content[].type === "tool_use"` grouped by `name` |
| **Error rate** | Count `is_error: true` in tool results / total tool calls |
| **Session title** | From `type === "ai-title"` `.aiTitle` field |
| **Last prompt preview** | From `type === "last-prompt"` `.lastPrompt` field |
| **Model used** | `message.model` on assistant messages |
| **Git branch** | `gitBranch` on user messages |
| **Entry point** | `entrypoint` on user messages |
| **Claude Code version** | `version` on user messages |
| **Extended thinking** | Count `content[].type === "thinking"` blocks |
| **Files edited** | From `file-history-snapshot` `.trackedFileBackups` keys |
| **File edit depth** | From `trackedFileBackups[file].version` |
| **API errors** | Count `type === "system"` && `subtype === "api_error"` |
| **Overnight session** | Check for `type === "date_change"` attachments |
| **Todo items** | From `todo_reminder` `content[]` + `itemCount` |
| **IDE diagnostics** | From `hook_additional_context` content |
| **CLAUDE.md usage** | From `nested_memory` entries |
| **Queued prompts** | Count `queued_command` attachments |
| **Response latency** | Time delta from user msg → next assistant msg |

---

## 3. `history.jsonl` — Command History

**Path:** `~/.claude/history.jsonl`  
**Format:** JSONL, one entry per command

```jsonc
{
  "display": "/config",
  "pastedContents": {},
  "timestamp": 1777081035488,     // Unix milliseconds
  "project": "/Users/graycup/Documents/GitHub/bulkgreencoffee-com",
  "sessionId": "8f3c6ca6-..."
}
```

Common `display` values: `/config`, `/stats`, `stats`, `/help`, `/clear`, `/code-review`

### Dashboard Widgets
- Most-used slash commands bar chart
- Project switching frequency (unique projects per day)
- Command usage trends over time

---

## 4. `sessions/{pid}.json` — Active Session Metadata

**Path:** `~/.claude/sessions/`

```jsonc
{
  "pid": 94635,
  "sessionId": "uuid",
  "cwd": "/Users/graycup/Documents/GitHub/farms-directory",
  "startedAt": 1777081035000,
  "procStart": "Mon Jun 15 01:02:45 2026",
  "version": "2.1.177",
  "peerProtocol": 1,
  "kind": "interactive",
  "entrypoint": "claude-vscode"
}
```

### Dashboard Widgets
- "Currently active" sessions badge
- Version adoption timeline (join with JSONL version field)
- Entry point split (IDE vs terminal)

---

## 5. `settings.json` — User Configuration

```jsonc
{
  "permissions": {
    "allow": [
      "Bash(curl -s https://...)",
      "WebFetch(domain:github.com)",
      "Read(/Users/graycup/Documents/GitHub/**)"
    ],
    "additionalDirectories": ["/Users/graycup/Documents/GitHub/farms-directory/..."]
  },
  "effortLevel": "medium",
  "theme": "dark-daltonized"
}
```

### Dashboard Widgets
- Current config display panel
- Permission count / trust level indicator
- Theme + effort level badges

---

## 6. `mcp-needs-auth-cache.json` — MCP Integration Status

```jsonc
{
  "claude.ai Google Calendar": { "timestamp": 1749926568741, "id": "mcpsrv_018j..." },
  "claude.ai Gmail": { "timestamp": ..., "id": "..." },
  "claude.ai Google Drive": { "timestamp": ..., "id": "..." }
}
```

### Dashboard Widgets
- MCP integrations panel with auth status badges

---

## 7. `plugins/` — Plugin Ecosystem

### `install-counts-cache.json`
```jsonc
{
  "counts": [
    { "plugin": "frontend-design@claude-plugins-official", "unique_installs": 277472 },
    { "plugin": "code-review@claude-plugins-official", "unique_installs": 189234 }
  ]
}
```

### `blocklist.json`
```jsonc
{
  "plugins": [
    { "plugin": "code-review@...", "added_at": "...", "reason": "just-a-test" }
  ]
}
```

### Dashboard Widgets
- Installed plugins list with install counts
- Blocked plugins list

---

## 8. `telemetry/` — Failed Telemetry Events

Events that failed to send to Anthropic. Not a complete log — only failures.

```jsonc
{
  "event_data": {
    "event_name": "tengu_feature_ok",
    "model": "claude-sonnet-4-6",
    "betas": "interleaved-thinking-2025-05-14,...",
    "env": {
      "platform": "darwin", "node_version": "22.x",
      "package_managers": "npm,bun", "runtimes": "node,bun",
      "arch": "arm64", "shell": "zsh"
    },
    "auth": { "organization_uuid": "...", "account_uuid": "..." }
  }
}
```

### Dashboard Widgets
- Environment info card (platform, arch, shell, runtimes)
- Active feature flags (`betas` field)

---

## 9. `ide/{pid}.lock` — Active IDE Sessions

```jsonc
{
  "pid": 19269,
  "workspaceFolders": ["/Users/graycup/Documents/GitHub/farms-directory"],
  "ideName": "Visual Studio Code",
  "transport": "ws",
  "runningInWindows": false,
  "authToken": "uuid"
}
```

### Dashboard Widgets
- Currently open workspaces
- IDE name badge

---

## Pricing Reference

`costUSD` is always `0` in the cache — compute from token counts:

| Model | Input /1M | Output /1M | Cache Read /1M | Cache Write /1M |
|---|---|---|---|---|
| claude-sonnet-4-6 | $3.00 | $15.00 | $0.30 | $3.75 |
| claude-opus-4-6 | $15.00 | $75.00 | $1.50 | $18.75 |
| claude-opus-4-5-20251101 | $15.00 | $75.00 | $1.50 | $18.75 |
| claude-sonnet-4-5-20250929 | $3.00 | $15.00 | $0.30 | $3.75 |
| claude-haiku-4-5 | $0.80 | $4.00 | $0.08 | $1.00 |

```ts
function computeCost(usage: TokenUsage, model: string): number {
  const p = PRICING[model]
  return (
    (usage.inputTokens / 1_000_000) * p.input +
    (usage.outputTokens / 1_000_000) * p.output +
    (usage.cacheReadInputTokens / 1_000_000) * p.cacheRead +
    (usage.cacheCreationInputTokens / 1_000_000) * p.cacheWrite
  )
}
```

---

## Complete Dashboard Widget Map

### Overview / Home
| Widget | Source | Notes |
|---|---|---|
| Total sessions KPI | `stats-cache.totalSessions` | |
| Total messages KPI | `stats-cache.totalMessages` | |
| Using Claude since | `stats-cache.firstSessionDate` | |
| Estimated total cost | `stats-cache.modelUsage` × pricing | costUSD is always 0 |
| Active projects count | `projects/` dir count | |
| Daily activity chart | `stats-cache.dailyActivity` | messages + sessions + tools |
| Hour of day heatmap | `stats-cache.hourCounts` | convert UTC → local |
| Model usage split | `stats-cache.modelUsage` | pie by tokens |
| Token trend by model | `stats-cache.dailyModelTokens` | stacked area chart |

### Sessions Browser
| Widget | Source | Notes |
|---|---|---|
| Session list | `projects/**/*.jsonl` | title, project, date, duration |
| Session title | `type === "ai-title"` → `aiTitle` | auto-generated name |
| Last prompt preview | `type === "last-prompt"` → `lastPrompt` | truncated |
| Session cost | sum `usage.*tokens` × pricing | |
| Session duration | last − first timestamp | clamp outliers >10h |
| Git branch | `user.gitBranch` | |
| Entry point badge | `user.entrypoint` | vscode vs cli |
| Claude Code version | `user.version` | |
| Extended thinking badge | count `thinking` blocks > 0 | |
| Overnight session badge | `date_change` attachment present | |
| Permission mode | `user.permissionMode` | acceptEdits vs default |

### Projects Breakdown
| Widget | Source | Notes |
|---|---|---|
| Project list with stats | `projects/` dirs | decode path from dir name |
| Messages per project | count assistant msgs | |
| Tokens per project | sum usage tokens | |
| Tool calls per project | count tool_use blocks | |
| Sessions per project | count JSONL files | |
| Last active date | last timestamp in any JSONL | |
| Cost per project | tokens × pricing | |
| Top branches per project | `gitBranch` field | |

### Tool Analytics
| Widget | Source | Notes |
|---|---|---|
| Tool call frequency bar | count `tool_use.name` | Bash, Edit, Read, Write, etc. |
| Tool error rate | `is_error` on tool_results | 210 errors in example |
| Tool error by type | `tool_use_id` → match errors | which tools fail most |
| Tool usage over time | JSONL timestamps + tool_use | trend by week |
| TodoWrite usage | count `TodoWrite` calls | task management adoption |
| Agent/Skill invocations | count `Agent`/`Skill` calls | sub-agent usage |
| WebFetch/WebSearch calls | count + `server_tool_use` | |

### File Activity
| Widget | Source | Notes |
|---|---|---|
| Most edited files | `file-history-snapshot.trackedFileBackups` keys | |
| Edit depth per file | `trackedFileBackups[file].version` | v14 = edited 14 times |
| Files by project | group by project dir | |
| File type breakdown | parse extension from filenames | .tsx, .ts, .json, etc. |
| Files Claude touched today | filter by `backupTime` | |

### Cache & Performance
| Widget | Source | Notes |
|---|---|---|
| Global cache hit rate | `cacheReadInputTokens / (input + cacheRead)` | usually 95%+ |
| Cache savings ($ value) | `cacheRead × (input_price − cacheRead_price)` | |
| Cache hit rate over time | per-day from JSONL | |
| Avg response latency | user → assistant timestamp delta | |
| Stop reason distribution | `stop_reason` field | tool_use vs end_turn |

### Errors & Reliability
| Widget | Source | Notes |
|---|---|---|
| API error count | `type === "system"` entries | ECONNRESET, ConnectionRefused |
| Error timeline | `system.timestamp` | |
| Error types breakdown | `system.error.connection.code` | |
| Network outage events | `system.error.isNetworkDown === true` | |
| Tool error rate | `is_error: true` on tool_results | |

### Code Quality Signals
| Widget | Source | Notes |
|---|---|---|
| IDE diagnostics seen | `hook_additional_context` content | TS errors, warnings |
| Diagnostic severity breakdown | parse `severity` field | Error vs Warning vs Hint |
| Files with most diagnostics | group by `filePath` in diagnostics | |
| Diagnostics over time | `hookEvent` timestamp | trend of error noise |

### Session Deep Dive (single session view)
| Widget | Source | Notes |
|---|---|---|
| Full conversation thread | all msgs in JSONL, sorted by timestamp | |
| Token usage per turn | `usage` on each assistant msg | |
| Tool calls timeline | extract tool_use from content | |
| Files edited this session | `file-history-snapshot` | |
| Todos in this session | `todo_reminder.content` | |
| CLAUDE.md files loaded | `nested_memory` entries | |
| Queued prompts | `queued_command` entries | |
| Thinking blocks | `thinking` content blocks | |
| Errors in session | `type === "system"` entries | |

### Version & Environment
| Widget | Source | Notes |
|---|---|---|
| Version history timeline | `user.version` over time | which version each session used |
| Upgrade dates | first msg with new version | |
| Platform info | `telemetry.env` | darwin, arch, shell |
| Active feature flags | `telemetry.betas` | comma-separated list |
| Package managers | `telemetry.env.package_managers` | npm, bun, etc. |

### Config & Integrations
| Widget | Source | Notes |
|---|---|---|
| MCP integrations | `mcp-needs-auth-cache.json` | auth status |
| Permissions list | `settings.permissions.allow` | |
| Trusted directories | `settings.permissions.additionalDirectories` | |
| Theme + effort level | `settings.theme`, `settings.effortLevel` | |
| Installed plugins | `plugins/` | |
| CLAUDE.md coverage | `nested_memory` unique paths | which projects configured |

---

## Key Gotchas

1. **`costUSD` is always 0** — always compute from token counts
2. **Project dir names** use `-` as separator — decode: leading `-` → `/`, but project names with `-` in them are ambiguous; use the `cwd` field inside JSONL as ground truth
3. **`longestSession.duration`** can be billions of ms — clamp sessions > 10 hours as stale
4. **`stats-cache.lastComputedDate`** — sessions after this date aren't in the cache; parse recent JSONL for live data
5. **`hourCounts` is UTC** — shift by local timezone offset for display
6. **JSONL lines can be non-JSON** — always try/catch parse; skip non-dict lines
7. **`<synthetic>` model** — appears in 11 messages; represents internally generated responses, not real API calls
8. **`ai-title` appears multiple times** per session (updated as title refines) — take the last one
9. **`last-prompt` appears multiple times** per session — take the last one
10. **`file-history-snapshot.isSnapshotUpdate: false`** = initial empty snapshot; only `true` ones have actual file data
11. **Tool result messages have `type === "user"`** with `content[0].type === "tool_result"` — distinguish from real user prompts by checking content type
12. **Extended thinking** is on for ~88% of sessions in example — nearly always enabled
13. **Cache hit rate** is typically 95%+ — cache read tokens massively dominate input tokens
14. **`queue-operation`** entries are internal plumbing — skip for analytics

---

## 10. Storage Analysis

### Real measurements (`~/.claude/` — 133 sessions, 45 projects)

| Directory | Size | Files | What's stored | Why it grows |
|---|---|---|---|---|
| `projects/` | **85 MB** | 189 JSONL files | Full conversation logs per session | Every message, tool call, token usage, thinking blocks — ~717 KB/session avg |
| `file-history/` | **17 MB** | 2,879 files | Actual file content backups for undo | Each file Claude edits is copied here per version; `package.json` backed up 7×, some tsx files 14× |
| `plugins/` | **5.3 MB** | 385 files | Plugin marketplace registry + all plugin code | Grows when new plugins are installed from marketplace |
| `backups/` | **260 KB** | 5 files | `.claude.json` config snapshots | One backup per config change |
| `cache/` | **252 KB** | 1 file | `changelog.md` — app changelog | Static, rarely changes |
| `telemetry/` | **124 KB** | 4 files | Failed telemetry event payloads | Only failed-to-send events accumulate here |
| `shell-snapshots/` | **16 KB** | 2 files | Zsh/bash function dumps | Captured once per shell init |
| `sessions/` | **12 KB** | 3 files | Active session PID metadata | One small JSON per running process |
| `ide/` | **20 KB** | — | VSCode workspace lock files | One per IDE window open |
| `session-env/` | **24 KB** | — | Environment variable snapshots | Usually empty or tiny |
| `history.jsonl` | **4 KB** | 1 file | Slash command history | One line per command typed |
| `stats-cache.json` | **12 KB** | 1 file | Pre-aggregated analytics | Rewritten on each cleanup |
| `settings.json` | **12 KB** | 1 file | Permissions + config | Grows with more allowed commands |
| `mcp-needs-auth-cache.json` | **4 KB** | 1 file | MCP server auth state | One entry per MCP integration |
| **Total** | **~108 MB** | | | |

### Storage breakdown by what actually costs space

```
projects/     85 MB  (79%)  ← conversation logs — the dominant cost
file-history/ 17 MB  (16%)  ← file backups for undo — proportional to edits
plugins/       5 MB  ( 5%)  ← one-time marketplace download
everything else < 1 MB each
```

### Per-session JSONL metrics (from real data)

```
Average session size:     717 KB
Median line size:       2,719 bytes  (each JSON event)
Average lines/session:    270 lines
```

### Storage growth projection (`projects/` only)

| Sessions | Storage |
|---|---|
| 133 (current) | 85 MB |
| 500 | ~340 MB |
| 1,000 | ~680 MB |
| 2,000 | ~1.4 GB |
| 5,000 | ~3.4 GB |

Growth accelerates with longer sessions (extended thinking blocks are large) and more WebFetch content (HTML pages are stored inline as tool results).

### `file-history/` structure

Organized by session UUID, not by project. Each file is a raw text backup with a content-hash filename + version suffix:

```
~/.claude/file-history/
└── {session-uuid}/
    ├── 53f1fc59531cac07@v1    ← hash of original content, version 1
    ├── 53f1fc59531cac07@v2    ← same file after Claude's first edit
    └── ee656db8fbb5794d@v1    ← different file, first backup
```

The mapping between hash→filename is stored in `file-history-snapshot` entries inside the project JSONL (via `trackedFileBackups[relativePath].backupFileName`).

### Per-project storage (real data, top 15)

| Project | Size | Sessions | MB/Session |
|---|---|---|---|
| graycup-com | 8.7 MB | 8 | 1.09 |
| opend2c | 8.3 MB | 7 | 1.19 |
| graybulk | 7.4 MB | 13 | 0.57 |
| nvault-app-webpimageoptimizer | 5.8 MB | 10 | 0.58 |
| aapka-app | 5.7 MB | 12 | 0.48 |
| fast-graycup-in | 4.7 MB | 4 | 1.18 |
| bulkgreencoffee-com | 4.7 MB | 5 | 0.94 |
| graycup-orders-main | 4.1 MB | 5 | 0.82 |
| bulkctc | 4.0 MB | 4 | 1.00 |
| nvault-app-clicksy | 3.9 MB | 2 | 1.95 |

High MB/session = long sessions with extended thinking or heavy WebFetch content.

### How to compute storage stats in the dashboard

```ts
// Run du on each subdir
const storageByDir = await Promise.all(
  ['projects', 'file-history', 'plugins', ...].map(async dir => ({
    dir,
    bytes: await du(`${CLAUDE_DIR}/${dir}`),
  }))
)

// Per-project from JSONL file sizes
const perProject = await Promise.all(
  projectDirs.map(async dir => ({
    project: decodeProjectDir(dir),
    bytes: await dirSize(`${CLAUDE_DIR}/projects/${dir}`),
    sessions: (await fs.readdir(...)).filter(f => f.endsWith('.jsonl')).length,
  }))
)
```

### Dashboard Widgets (Storage)

| Widget | How |
|---|---|
| Total `.claude/` size gauge | `du -sh ~/.claude/` |
| Storage breakdown donut | `du -sh` each subdirectory |
| Per-project storage bar | `du -sh` each project dir |
| Largest sessions list | file size of each JSONL |
| Storage growth over time | correlate `firstSessionDate` + sessions count |
| File history size vs projects | ratio chart |
| "Claude is using X MB" summary card | sum of all above |

---

## 11. Network / Internet Usage

### What Claude Code connects to

Confirmed via `lsof -i` on live processes:

| Endpoint | IP | Who | Purpose |
|---|---|---|---|
| `api.anthropic.com` | `2607:6bc0::10` (ANTHROPIC-V6) | Anthropic, PBC | LLM inference — all prompt/response traffic |
| Google Cloud edge | `2600:1901:0:3084::` (GOOGLE-CLOUD) | Google LLC | CDN / telemetry delivery |

Claude Code binaries live at `~/.vscode/extensions/anthropic.claude-code-{version}-darwin-arm64/resources/native-binary/claude` and maintain persistent HTTPS connections (port 443) to Anthropic's API.

### Bandwidth estimation from token data

Since `nettop` requires SIP privileges for per-process byte counts, the most practical approach is **estimating from token data** stored in the JSONL files. Each token ≈ 4 bytes of text content, with ~1.3× JSON wrapper overhead.

**Actual computed values for this user (133 sessions, all-time from `stats-cache.json`):**

```
Total input tokens sent:                211,661
Cache write tokens sent:             26,957,557
Cache read tokens (just metadata):  475,355,847
Total output tokens received:         4,612,217

Estimated upstream (→ Anthropic):    195.8 MB
Estimated downstream (← Anthropic):   78.5 MB
Estimated HTTP overhead:               54.5 MB
Total API bandwidth:                  ~329 MB

Without cache, upstream would be:     2.61 GB
Cache bandwidth saved:                2.47 GB   ← cache cuts bandwidth by 88%
```

| Traffic type | Tokens | Est. Bytes | Notes |
|---|---|---|---|
| Prompts sent (input) | 211,661 | ~1.1 MB | Actual new prompt text only |
| Cache writes sent | 26,957,557 | ~140 MB | Context written to Anthropic's cache |
| Output received | 4,612,217 | ~24 MB | Claude's responses streamed back |
| HTTP overhead | ~26,600 requests | ~54 MB | Headers, JSON wrapper, streaming chunks |
| **Total upstream** | | **~195.8 MB** | Sent to Anthropic |
| **Total downstream** | | **~78.5 MB** | Received from Anthropic |
| **Cache saved** | 475,355,847 | **~2.47 GB** | What would have been re-sent without cache |

The cache is the biggest network optimization — 475M tokens read from cache vs 211K sent as fresh input = **99.96% cache hit rate on input tokens**.

### Bandwidth formula

```ts
const BYTES_PER_TOKEN = 4 * 1.3  // 4 chars × JSON overhead

function estimateBandwidth(usage: ModelUsage) {
  const upstream =
    usage.inputTokens * BYTES_PER_TOKEN +          // prompts
    usage.cacheCreationInputTokens * BYTES_PER_TOKEN  // cache writes
  
  const downstream =
    usage.outputTokens * BYTES_PER_TOKEN           // responses

  const saved =
    usage.cacheReadInputTokens * BYTES_PER_TOKEN   // what would have been sent

  return { upstream, downstream, saved }
}
```

### WebFetch & WebSearch traffic

Separate from LLM inference — Claude can fetch web pages inline.

```jsonc
// In assistant message usage:
"server_tool_use": {
  "web_search_requests": 0,   // Brave/Google search API calls
  "web_fetch_requests": 0     // Direct URL fetches
}
```

WebFetch stores the full fetched HTML/text as tool results in the JSONL — large fetched pages (50–500 KB each) can significantly inflate both JSONL storage and bandwidth. Count from `WebFetch` and `WebSearch` tool_use entries, and sum response content lengths in tool_result entries.

### How to monitor network in real-time on macOS

#### Option 1 — `lsof` (no privileges needed)
Shows active connections but **not byte counts**. Use to confirm Claude is connected and to which endpoint.

```bash
lsof -i -n -P 2>/dev/null | grep "claude.*ESTABLISHED"
# claude  99020  ...  TCP [::]:62226->[2607:6bc0::10]:443 (ESTABLISHED)
```

#### Option 2 — `nettop` (recommended for real-time bytes)
Per-process byte counts. Requires running from Terminal (not sandboxed):

```bash
# Snapshot mode — 3 samples, per-process
nettop -P -l 3 -t wifi -t wired

# Stream mode — updates every second
nettop -P -t wifi -t wired
# Look for "claude" rows — shows bytes_in and bytes_out columns
```

In code (for dashboard background monitoring):

```ts
import { spawn } from 'child_process'

function monitorClaudeNetwork(callback: (bytesIn: number, bytesOut: number) => void) {
  const proc = spawn('nettop', ['-P', '-l', '0', '-t', 'wifi', '-t', 'wired'])
  proc.stdout.on('data', (chunk: Buffer) => {
    const lines = chunk.toString().split('\n')
    for (const line of lines) {
      if (line.startsWith('claude')) {
        // parse bytes_in and bytes_out columns
        const cols = line.trim().split(/\s+/)
        callback(parseInt(cols[/* bytes_in col */]), parseInt(cols[/* bytes_out col */]))
      }
    }
  })
}
```

#### Option 3 — `netstat -ibn` (interface totals)
Gives cumulative bytes on the whole network interface (not per-process). Useful as a sanity check:

```bash
netstat -ibn | grep en0
# en0  1500  ...  17124456 pkts  19842195922 bytes-in  4108573 pkts  1759848331 bytes-out
```

Current machine totals on en0: **~18.5 GB received**, **~1.6 GB sent** (all time, all processes).

#### Option 4 — Estimate from JSONL (best for historical analysis)
No OS permissions needed. Computable entirely from the data already in `projects/*.jsonl`. This is what the dashboard should use for historical bandwidth charts.

```ts
// Aggregate across all sessions
let totalUpstream = 0, totalDownstream = 0, totalSaved = 0

for (const msg of assistantMessages) {
  const u = msg.message.usage
  totalUpstream += (u.input_tokens + u.cache_creation_input_tokens) * 5.2
  totalDownstream += u.output_tokens * 5.2
  totalSaved += u.cache_read_input_tokens * 5.2
}
```

### Network connection types Claude Code uses

| Connection | Protocol | Destination | Triggered by |
|---|---|---|---|
| LLM API calls | HTTPS/TLS (port 443) | `api.anthropic.com` | Every user prompt |
| Telemetry | HTTPS (port 443) | Google Cloud CDN | Session events (async, best-effort) |
| WebFetch tool | HTTPS/HTTP | Any URL | User asks Claude to fetch a URL |
| WebSearch tool | HTTPS (port 443) | Anthropic's search proxy | User asks Claude to search |
| Plugin marketplace | HTTPS | GitHub API | Plugin install/update |
| MCP servers | Various (ws/wss) | User-configured | MCP tool calls |
| IDE websocket | `ws://localhost:*` | Local VSCode extension | IDE ↔ Claude bridge |

### Dashboard Widgets (Network)

| Widget | Source | How |
|---|---|---|
| Estimated total API bandwidth | JSONL `usage` fields | `(input + cache_write) × 5.2` bytes/token |
| Bandwidth saved by cache | JSONL `cache_read_input_tokens` | cache read tokens × input price per byte |
| Upstream vs downstream ratio | JSONL `usage` | compare input vs output tokens |
| WebFetch requests count | JSONL `server_tool_use.web_fetch_requests` | sum across sessions |
| WebSearch requests count | JSONL `server_tool_use.web_search_requests` | sum across sessions |
| Real-time bytes in/out | `nettop -P -l 1` | shell spawn, parse claude rows |
| Bandwidth per project | per-project JSONL aggregation | group by project dir |
| Bandwidth over time (daily) | JSONL timestamp + `dailyModelTokens` | align tokens to dates |
| Cache efficiency $ saved | token counts × (input price − cache read price) | e.g. $3.00 vs $0.30/1M |

### Gotchas for network measurement

1. **`nettop` byte counts reset on process restart** — can't get historical per-process totals, only current session
2. **Token-based estimates are slightly low** — JSON field names, UUIDs, and metadata in each API call add ~20% overhead not counted in tokens
3. **Cache read tokens don't cost full bandwidth** — only a tiny cache-lookup token is sent; the response comes from Anthropic's servers but doesn't re-transmit your original context
4. **Streaming responses** — Claude uses SSE streaming, so downstream bandwidth arrives incrementally; each chunk has HTTP chunked-transfer overhead
5. **File content in tool results** — when Claude reads a 50KB file with the Read tool, that content is counted as input tokens AND stored in JSONL, inflating both bandwidth and disk estimates
6. **MCP websocket** is localhost only — zero internet bandwidth

---

## 12. Time Analysis — How Claude Spends Its Time

This is fully computable from JSONL timestamps. Every message has a `timestamp` field, which lets us reconstruct exactly where time went in each session: thinking, writing, running tools, or waiting.

---

### The four time buckets

```
User sends prompt
     ↓
[■■■ API TIME ■■■■■■■■■■■■■■■■■] ← Claude thinking + writing + network latency
     ↓
Assistant responds (tool_use)
     ↓
[■ TOOL TIME ■] ← Bash/Read/Edit/Write executing locally
     ↓
Tool result returned
     ↓
[■■■ API TIME ■■■■■■■■■■■■] ← Claude thinking + writing again
     ↓
...repeat until end_turn...
     ↓
[     IDLE TIME     ] ← User reading response, typing next prompt
```

**API time** = `assistant.timestamp − preceding user/tool_result.timestamp`  
**Tool time** = `tool_result.timestamp − preceding assistant.timestamp`  
**Idle time** = `next user_prompt.timestamp − last assistant end_turn.timestamp`

---

### Measured values (example data, all sessions)

#### API Response Time (user → assistant)

```
Samples:   12,418 turns
Median:     4.9 s
Average:   11.2 s
Min:        0.0 s  (cache hit, near-instant)
Max:      298.2 s  (very long thinking on complex prompt)

Distribution:
  < 2s      751 turns  ( 6%)  ← usually cached/trivial
  2–5s    5,544 turns  (45%)  ← most common, standard response
  5–15s   4,042 turns  (33%)  ← moderate thinking
  15–60s  1,757 turns  (14%)  ← extended thinking / long output
  > 60s     324 turns  ( 3%)  ← complex tasks, deep thinking
```

By stop reason:
```
tool_use   11,522 turns  avg=11.5s  med=4.9s  ← Claude decided to call a tool
end_turn      881 turns  avg= 6.9s  med=5.3s  ← Claude finished responding
```

#### Tool Execution Time (assistant → tool_result)

```
Samples:  7,048 tool calls
Median:   1.01 s
Average:  2.50 s
Min:      0.001 s
Max:    118.5 s  (background server startup)

Distribution:
  < 0.1s    3,102 calls  (44%)  ← Read, TodoWrite, quick checks
  0.1–0.5s    199 calls  ( 3%)
  0.5–2s    3,039 calls  (43%)  ← Edit, Write, short Bash
  2–10s       402 calls  ( 6%)
  > 10s       306 calls  ( 4%)  ← npm install, cargo build, server starts
```

#### Per-tool execution times

| Tool | Calls | Avg Time | Median | Max | What takes the time |
|---|---|---|---|---|---|
| `Read` | 1,683 | 0.02s | 0.01s | 0.9s | Pure disk I/O — nearly instant |
| `TodoWrite` | 174 | 0.00s | 0.00s | 0.0s | In-memory write — instant |
| `Write` | 896 | 1.01s | 1.02s | 2.0s | Disk write + permission prompt |
| `Edit` | 2,106 | 1.03s | 1.02s | 13.4s | Diff apply + permission prompt |
| `Bash` | 2,064 | 6.05s | 0.07s | 118.5s | Huge variance — most fast, some very slow |
| `ToolSearch` | 39 | 2.97s | 0.01s | 115.6s | Schema loading, occasionally slow |
| `Skill` | 13 | 0.61s | 0.02s | 4.0s | Skill loading |
| `WebFetch` | 45 | 14.79s | 6.31s | 105.5s | Network round trip to fetch URL |
| `AskUserQuestion` | 7 | 46.35s | 48.59s | 97.2s | Waiting for human to respond |
| `Agent` | 15 | 61.56s | 46.89s | 111.3s | Spawning + completing sub-agent |

`Bash` median is 0.07s but avg is 6.05s — shows the long tail: most commands are fast (`ls`, `grep`, `cat`) while a few are very slow (server starts, package installs, compilation).

Slow Bash commands seen (> 60s): background server start (`bun src/server.ts`), `npm create`, `cargo test`, `next dev` startup, `cargo build`.

#### Overall time budget (all example sessions)

```
Total measured active time:    43.54 hours
  API time (think + write):    38.64 hrs  (88.7%)
  Tool execution time:          4.90 hrs  (11.3%)
```

This is time Claude was actively working. Idle/read time between sessions is not captured.

---

### Thinking vs Writing breakdown

Claude's output per assistant turn has two components: `thinking` blocks (internal reasoning, never shown to user) and `text` blocks (actual response). Both are stored in the JSONL.

**Computed from character counts across all thinking + text content:**

```
Thinking chars:     2,134,934   (74.6% of all output chars)
Text output chars:    727,504   (25.4%)
Ratio:              2.93× more thinking than writing

Average per turn:
  Thinking block:  1,186 chars  ≈ 297 tokens
  Text response:     206 chars  ≈  52 tokens
```

Claude thinks ~3× more than it writes. The thinking is essentially invisible work.

**Thinking presence per turn:**

```
Turns with thinking blocks:     1,800  (14.5%)
Turns without thinking:        10,638  (85.5%)
  └ Tool-only turns (no text or think): 7,102  (56.9% of all turns)
```

Most turns are tool-only — Claude silently calls Bash/Read/Edit without any thinking or text output. When Claude does think, those turns are the "decision points" in a session.

**Thinking depth grows with session length:**

```
Early turns (0–9):   avg 722 chars of thinking
Mid turns (10–49):  avg 1,227 chars of thinking
Late turns (50+):   avg 1,288 chars of thinking
```

Claude thinks harder as the session goes on and context grows — likely because late-session problems are more complex or require integrating earlier context.

---

### How to compute in the dashboard

```ts
interface TurnTiming {
  sessionId: string
  turnIndex: number
  apiMs: number          // user → assistant (think + write + network)
  toolMs: number         // assistant → tool_result
  thinkingChars: number  // chars in thinking blocks
  textChars: number      // chars in text blocks
  toolCalls: string[]    // tool names called this turn
  stopReason: string     // 'tool_use' | 'end_turn'
}

function analyzeTiming(messages: Message[]): TurnTiming[] {
  // Sort by timestamp, then for each assistant message:
  // 1. API time = ts(assistant) - ts(prev user/tool_result)
  // 2. Tool time = ts(tool_result) - ts(this assistant)
  // 3. Thinking chars = sum len(block.thinking) for thinking blocks
  // 4. Text chars = sum len(block.text) for text blocks
}
```

For response latency histogram, filter `apiMs` to `0 < apiMs < 300_000` (5 min) to remove stale/outlier sessions.

---

### Daily time budget (sample from example data)

| Date | API Hours | Tool Hours | API% |
|---|---|---|---|
| 2026-05-13 | 2.32 | 0.23 | 91% |
| 2026-05-31 | 2.86 | 0.52 | 85% |
| 2026-06-02 | 2.78 | 0.36 | 88% |
| 2026-06-03 | **4.14** | **0.62** | 87% |
| 2026-06-08 | 3.25 | 0.37 | 90% |

Busy days have 3–4 hours of actual API time (Claude actively thinking/writing). Tool execution is consistently 10–15% of total time.

---

### Dashboard Widgets (Time)

| Widget | Source | Notes |
|---|---|---|
| Total hours Claude worked | sum all `apiMs` across sessions | e.g., "38.6 hrs of API time" |
| API response latency histogram | `apiMs` per turn | 5 buckets: <2s, 2-5s, 5-15s, 15-60s, >60s |
| Tool execution time by tool | `toolMs` grouped by tool name | bar chart with avg+median |
| Thinking vs writing ratio | `thinkingChars / textChars` | donut chart: 74.6% think, 25.4% write |
| Daily time budget | `daily_api_ms` + `daily_tool_ms` | stacked bar by day |
| Thinking depth over session | `thinkingChars` by `turnIndex` | line chart — grows with session length |
| Slowest Bash commands | filter `Bash` toolMs > 10s | table: duration + command output preview |
| Turns with vs without thinking | count thinking blocks | bar: decision turns vs tool-only turns |
| Total tool execution time | sum all `toolMs` | "Claude ran tools for 4.9 hrs total" |
| Longest single API wait | max `apiMs` | "longest thinking: 298s" |
| AskUserQuestion wait times | filter tool=AskUserQuestion `toolMs` | shows how long user takes to respond |
| Agent sub-task durations | filter tool=Agent `toolMs` | "sub-agents took avg 62s each" |

---

## Build Priority

### Tier 1 — Zero parsing, instant value
Read `stats-cache.json` only:
- KPI cards, daily activity chart, model split, hour heatmap, cost estimate

### Tier 2 — Light JSONL parse
Scan each JSONL for top-level fields only (`type`, `timestamp`, `sessionId`, first/last timestamps):
- Session list, session browser, project breakdown, durations

### Tier 3 — Full JSONL parse
Read `message.usage` and `message.content` on every assistant message:
- Per-session cost, token trends, tool frequency, cache hit rate, error tracking

### Tier 4 — Deep attachment parse
Parse `attachment` subtypes:
- File activity heatmap, todo tracking, IDE diagnostic signals, CLAUDE.md coverage, overnight sessions
