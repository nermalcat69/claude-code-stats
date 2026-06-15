# Claude Code Stats — Data Points Reference

All data lives at `~/.claude/` on macOS (`/Users/{username}/.claude/`).  
Real path confirmed: `/Users/graycup/.claude/`

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
