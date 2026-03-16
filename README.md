# lattice

[![CI](https://github.com/cyperx84/lattice/actions/workflows/ci.yml/badge.svg)](https://github.com/cyperx84/lattice/actions/workflows/ci.yml)

Mental models engine — apply Charlie Munger's latticework of 98 cognitive frameworks to any problem.

Part of the [CyperX CLI ecosystem](#ecosystem-integration): works standalone, as an MCP server, or integrated with multiplan, content-breakdown, clwatch, and OpenClaw agents.

## Install

```bash
brew install cyperx84/tap/lattice
# or
go install github.com/cyperx84/lattice@latest
```

## Quick Start

```bash
# What models should I use? (instant, no LLM needed)
lattice suggest "should I hire or outsource"

# Think through a problem (applies top 3 models)
lattice think "should we build or buy our auth system" --no-llm

# Apply a specific model
lattice apply inversion "designing microservices architecture" --no-llm

# Search for models
lattice search "scaling"

# Record a decision and track it
lattice decide "raise prices 20%" --quick --prediction "churn under 5%"

# Browse everything
lattice list
```

## Command Reference

| Command | Description | Needs LLM? |
|---------|-------------|:---:|
| `suggest <situation>` | Recommend models for a situation | No |
| `think <problem>` | Surface top models + apply thinking steps | Optional |
| `apply <slug> <context>` | Apply one model to a context | Optional |
| `search <keyword>` | Search model index | No |
| `list [--category "..."]` | List all models | No |
| `info <slug>` | Show full model details | No |
| `add <name> [--from URL]` | Add a custom model | Yes |
| `remove <slug>` | Remove a user-added model | No |
| `decide <decision>` | Record a decision with model analysis + prediction | Optional |
| `journal [--due]` | View decision journal, filter by review date | No |
| `journal review <id>` | Review a past decision, record outcome | No |
| `serve` | Start MCP server (stdio) | No |
| `history [--limit N]` | View session history | No |
| `history clear` | Delete all history | No |
| `completion <shell>` | Generate shell completion script | No |

## Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--json` | Structured JSON output | off |
| `--no-llm` | Skip LLM, return static steps/questions | off |
| `--llm-cmd` | LLM command for synthesis | `claude -p` |
| `--timeout` | LLM timeout in seconds | 60 |
| `--verbose` | Show progress | off |
| `--no-history` | Skip saving to history | off |

## Model Categories

98 models across 8 disciplines:

| Category | IDs | Examples |
|----------|-----|----------|
| General Thinking Tools | m01-m09 | First principles, inversion, second-order thinking |
| Physics/Chemistry/Biology | m10-m29 | Leverage, inertia, feedback loops, ecosystems |
| Systems Thinking | m30-m40 | Bottlenecks, scale, margin of safety, emergence |
| Mathematics | m41-m47 | Randomness, regression to mean, local vs global maxima |
| Economics | m48-m59 | Trade-offs, scarcity, creative destruction |
| Art | m60-m70 | Framing, audience, contrast, rhythm |
| Warfare & Strategy | m71-m75 | Asymmetric warfare, seeing the front |
| Human Nature & Judgment | m76-m98 | Cognitive biases, incentives, social proof |

## Custom Models

Add your own models — they're stored in `~/.config/lattice/models/` and merged with built-in models at runtime:

```bash
# Generate a new model via LLM
lattice add "Network Effects"
lattice add "Lindy Effect" --from "https://fs.blog/lindy-effect/"

# Remove a custom model
lattice remove network_effects
```

Custom models are immediately searchable. They follow the same markdown format as built-in models.

## Configuration

`~/.config/lattice/config.yml`:

```yaml
llm_cmd: "claude -p"    # or "gemini -p", "codex exec", etc.
default_models: 3        # how many models to apply in `think`
vault_path: ""           # optional: Obsidian vault path for decision sync
vault_folder: "decisions" # folder within vault (default: "decisions")
```

## Decision Journal

Track decisions, apply mental models, record predictions, and review outcomes over time.

```bash
# Record a decision (guided: models + thinking steps + prediction prompt)
lattice decide "raise prices 20%"

# Quick mode: skip thinking steps, just capture prediction
lattice decide "raise prices 20%" --quick --prediction "churn under 5%"

# Project-local (ADR-style, saves to ./decisions/)
lattice decide "use postgres over mongo" --project

# Force specific models
lattice decide "hire vs outsource" --models inversion,trade-offs

# View recent decisions
lattice journal
lattice journal --limit 50

# Show decisions due for review
lattice journal --due

# Review a past decision
lattice journal review d-20260317-001
```

Decisions are saved as markdown files with YAML frontmatter in `~/.config/lattice/journal/`. Review dates are automatically set to 30 and 90 days out.

If `vault_path` is set in config, entries are also copied to your Obsidian vault.

## MCP Server

lattice includes a [Model Context Protocol](https://modelcontextprotocol.io/) server so AI assistants can use mental models as tools. No API keys required — all responses use the built-in model data.

```bash
lattice serve              # start on stdio
lattice serve --verbose    # debug logging to stderr
```

Exposes 5 tools: `think`, `suggest`, `search`, `apply`, `list`.

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "lattice": {
      "command": "lattice",
      "args": ["serve"]
    }
  }
}
```

### Cursor

Add to `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "lattice": {
      "command": "lattice",
      "args": ["serve"]
    }
  }
}
```

### Claude Code

```bash
claude mcp add lattice lattice serve
```

## Ecosystem Integration

Lattice is designed to interoperate with the CyperX CLI ecosystem. Each integration is optional — if lattice isn't on PATH, the other tools skip it gracefully.

### multiplan

[multiplan](https://github.com/cyperx84/multiplan) runs 4 models in parallel with lens-based prompts. Lattice adds **Phase 0 — Mental Model Framing** before planning starts.

```bash
# Automatic: multiplan detects lattice on PATH
multiplan "Design a rate limiting system"
# → Phase 0: lattice surfaces relevant models (e.g., Bottlenecks, Scale, Trade-offs)
# → Phase 1: each model's lens prompt includes the mental model framing
# → Output: lattice_framing.md in the run directory

# Skip lattice framing
multiplan "Design a cache layer" --skip-lattice
```

### content-breakdown

[content-breakdown](https://github.com/cyperx84/content-breakdown) transforms videos/articles into structured notes. Lattice adds:

1. **Mental models lens** — a new lens (`--lens mental-models`) that surfaces which cognitive frameworks the content uses
2. **`--think` flag** — appends a mental models analysis section to the output

```bash
# Apply mental models lens
breakdown run "https://youtube.com/watch?v=..." --lens mental-models

# Add mental models to any breakdown
breakdown run "https://youtube.com/watch?v=..." --think
```

### clwatch

[clwatch](https://github.com/cyperx84/clwatch) tracks AI coding CLI changelogs. Lattice adds mental model tagging:

```bash
# Tag recent changes with mental models
clwatch think claude-code
# → "Claude Code /simplify = Friction Reduction + Efficiency"

# Also appears in diff output
clwatch diff claude-code --since 7d
# → includes 🧠 Mental Models section
```

### OpenClaw Agents

Lattice is installed as a shared [OpenClaw](https://github.com/openclaw/openclaw) skill at `~/.openclaw/skills/lattice/`. Any OpenClaw agent (including those created by [ClawForge](https://github.com/cyperx84/clawforge)) can use it automatically.

Trigger phrases: "think through", "apply mental models", "inversion", "second-order thinking", "what framework should I use", etc.

### How it all connects

```
┌──────────────────────────────────────────────────┐
│                   OpenClaw Agents                 │
│         (Claw, Builder, any ClawForge agent)      │
│                        │                          │
│              lattice skill (auto)                  │
└────────────────────────┬─────────────────────────┘
                         │
              ┌──────────┴──────────┐
              │      lattice        │
              │   98 mental models  │
              │   suggest/think/    │
              │   apply/search      │
              └──────────┬──────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
    multiplan    content-breakdown    clwatch
   (Phase 0)      (--think flag)    (think cmd)
   model framing  model analysis    model tagging
```

## Shell Completions

Lattice supports shell completion for bash, zsh, and fish via Cobra's built-in completion:

```bash
# Zsh (add to ~/.zshrc)
eval "$(lattice completion zsh)"
# or for persistent completion:
lattice completion zsh > "${fpath[1]}/_lattice"

# Bash (add to ~/.bashrc)
eval "$(lattice completion bash)"

# Fish
lattice completion fish > ~/.config/fish/completions/lattice.fish
```

## Session History

Lattice saves a history of `think` and `apply` sessions to `~/.config/lattice/history/`:

```bash
lattice history              # show last 20 sessions
lattice history --limit 50   # show more
lattice history --json        # structured JSON output
lattice history clear         # delete all history
```

To skip saving a session: `lattice think "..." --no-history`

## Color Output

Lattice uses ANSI colors for readable terminal output. To disable:

```bash
NO_COLOR=1 lattice list    # disable color (https://no-color.org/)
lattice list --json         # JSON output is always color-free
```

## Releases

Lattice uses [GoReleaser](https://goreleaser.com/) for cross-platform releases. Binaries are available for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, and windows/amd64.

Releases are triggered automatically when a version tag is pushed:

```bash
git tag v0.x.0
git push --tags
```

## Development

```bash
git clone https://github.com/cyperx84/lattice
cd lattice
go build ./...     # build
go test ./...      # test (19 tests)
go vet ./...       # lint
```

## License

MIT — [CyperX](https://github.com/cyperx84)
