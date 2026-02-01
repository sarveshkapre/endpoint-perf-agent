# AGENTS

## Scope
- Work only inside this repository.
- Keep changes small and ship incremental improvements.

## Commands
- Setup: `make setup`
- Dev: `make dev`
- Tests: `make test`
- Lint: `make lint`
- Typecheck: `make typecheck`
- Build: `make build`
- Quality gate: `make check`

## Conventions
- Prefer standard library unless a dependency is clearly justified.
- Keep CLI flags stable; document any changes in `docs/CHANGELOG.md`.
- Update `docs/PLAN.md` for architectural changes.
