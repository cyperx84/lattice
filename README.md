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

# List all models
lattice list

# Filter by category
lattice list --category "General Thinking Tools"
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

## Integration

lattice integrates with:
- **multiplan** — Phase 0 mental model framing before parallel planning
- **content-breakdown** — Mental models lens + `--think` flag
- **clwatch** — `think` command for changelog mental model analysis
