# lattice

Mental models engine — apply Charlie Munger's latticework of 98 cognitive frameworks to any problem.

## Install

```bash
brew install cyperx84/tap/lattice
# or
go install github.com/cyperx84/lattice@latest
```

## Usage

```bash
# Think through a problem (surfaces top 3 models + applies thinking steps)
lattice think "Should we build or buy our auth system?"

# Use specific models
lattice think "Scaling our API" --models inversion,second-order-thinking,bottlenecks

# Search for relevant models
lattice search "decision making"

# Apply a single model
lattice apply inversion "designing microservices architecture"

# Get model suggestions for a situation (no LLM needed)
lattice suggest "I'm deciding whether to hire more engineers or outsource"

# Get more suggestions
lattice suggest "team conflict resolution" --count 10

# List all models
lattice list

# Filter by category
lattice list --category "General Thinking Tools"

# Start MCP server (for Claude Desktop, Cursor, Claude Code, etc.)
lattice serve
```

## Flags

All commands support:
- `--json` — structured JSON output
- `--llm-cmd` — LLM command for synthesis (default: `claude -p`)
- `--verbose` — show progress

## Configuration

`~/.config/lattice/config.yml`:

```yaml
llm_cmd: "claude -p"
default_models: 3
```

## Model Categories

| Category | Models | Examples |
|----------|--------|----------|
| General Thinking Tools | m01-m09 | First principles, inversion, second-order thinking |
| Physics/Chemistry/Biology | m10-m29 | Leverage, inertia, feedback loops |
| Systems Thinking | m30-m40 | Bottlenecks, scale, margin of safety |
| Mathematics | m41-m47 | Randomness, regression to mean |
| Economics | m48-m59 | Trade-offs, scarcity, creative destruction |
| Art | m60-m70 | Framing, audience, contrast |
| Warfare & Strategy | m71-m75 | Asymmetric warfare, seeing the front |
| Human Nature & Judgment | m76-m98 | Cognitive biases, incentives, social proof |

## Suggest Command

`lattice suggest` recommends which mental models to use for a situation **without applying them**. It's a fast, LLM-free recommender.

```bash
$ lattice suggest "should I hire or outsource?"

Suggested models for: "should I hire or outsource?"

1. Trade-offs & Opportunity Cost (Economics)
   Every decision closes other doors. Map the opportunity cost of each path.
   → Why: Relevant to: hiring decisions, resource allocation

2. Circle of Competence (General Thinking Tools)
   Know what you're good at. Outsource what falls outside your circle.
   → Why: Model directly addresses competence

3. Specialization (Economics)
   Focus drives mastery. Consider where specialization serves you best.
   → Why: Relevant to: outsourcing, team structure
```

Supports `--json` for structured output and `--count N` to control how many suggestions (default 5).

## MCP Server Mode

lattice includes a built-in [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server, so AI assistants can use mental models as tools.

```bash
lattice serve              # start MCP server on stdio
lattice serve --verbose    # with debug logging to stderr
```

The MCP server exposes 5 tools: `think`, `suggest`, `search`, `apply`, and `list`. No LLM or API keys required — all tools return the model's built-in thinking steps and coaching questions directly.

### Setup for Claude Desktop

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

### Setup for Cursor

Add to `.cursor/mcp.json` in your project:

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

### Setup for Claude Code

```bash
claude mcp add lattice lattice serve
```

## Integration

lattice integrates with:
- **multiplan** — Phase 0 mental model framing before parallel planning
- **content-breakdown** — Mental models lens + `--think` flag
- **clwatch** — `think` command for changelog mental model analysis
