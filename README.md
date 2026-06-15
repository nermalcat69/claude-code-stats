# claude-code-stats

![cover](static/cover.png)

A local dashboard for your Claude Code usage — sessions, costs, tool calls, network activity, and more. Reads `~/.claude` directly, nothing leaves your machine.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/nermalcat69/claude-code-stats/main/install.sh | sh
```

Installs to `~/.local/bin/claude-stats`. Or download a binary directly from [Releases](../../releases).

> Make sure `~/.local/bin` is in your `$PATH`. Add `export PATH="$HOME/.local/bin:$PATH"` to your `~/.zshrc` or `~/.bashrc` if needed.

Opens at **http://localhost:6967**.
