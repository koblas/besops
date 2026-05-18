# CLAUDE.md

Rules for working in this repo. Rooted in **Bryan Finster** (testing, continuous delivery) and **Martin Fowler** (design, refactoring).

**Core idea:** code is finished only when it can be changed safely tomorrow. Tests prove behavior. Design keeps change cheap.

## Prime directives (IMPORTANT — apply to every task)
1. **YOU MUST** write a failing test before implementing new behavior. If a test would not have failed before your change, the test is wrong.
2. **YOU MUST** keep every change small enough to merge to trunk today. If not, split it and show the sequence first.
3. **Don't** mix refactors with behavior changes in one diff. Separate commits, separate diffs.
4. **Don't** modify or delete a test to make a change pass. Fix the implementation, or fix the test in a separately reviewable commit with explicit reasoning.
5. **Don't** silently change a public contract. Propose a deprecation path.

## Testing (Finster)
- **Test behavior, not implementation.** A pure refactor must not break tests. If it does, the tests assert the wrong thing.
- **TDD for new logic**: red → green → refactor. Use Given/When/Then names for business behavior.
- **Coverage is a diagnostic, not a goal.** Every test must fail when the behavior breaks. Cover negative and edge cases explicitly.
- **Pyramid, not hourglass**: many fast unit tests, few integration, minimal E2E.
- Test names should tell on-call what broke without reading the implementation.

## Change shape (Finster)
- Hours of work per change, not days.
- For incomplete features, prefer **feature flags**, **branch by abstraction**, or **expand/contract** migrations over long-lived branches.

## Design (Fowler)
**Four rules of simple design**, in priority order — code should:
1. Pass its tests.
2. Reveal intent (if a comment explains *what*, rename instead).
3. Have no duplicated knowledge.
4. Have the fewest elements. No abstraction without a concrete reason.

**Functions**: short, one level of abstraction, no flag arguments, few parameters (introduce a parameter object past 3–4), no hidden side effects.

**Modules**: single reason to change, high cohesion, low coupling. Business logic doesn't know about HTTP, SQL, or framework annotations.

Smell-driven refactoring: if you see duplicated code, long function, large class, long parameter list, divergent change, shotgun surgery, feature envy, data clumps, primitive obsession, switch on type code, speculative generality, middle man, or inappropriate intimacy → refactor in the touched area or flag it explicitly.

**Naming, formatting, and style**: defer to the formatter/linter (see `## Tooling` below). Don't burn cycles formatting — let the tool do it.

## Errors
Fail loudly at the boundary. Never swallow exceptions — handle, wrap with context, or rethrow. Validate at trust boundaries; inside, prefer types that make invalid states unrepresentable. Error messages include what was attempted, expected, and seen.

## Workflow per task
**Before**: state the behavior change in one sentence; identify the smallest change; list the tests.
**During**: failing test first; refactors and behavior changes in separate diffs; flag smells in the touched area.
**After**: confirm each new behavior has a test that fails without the new code; confirm the diff is trunk-mergeable today; note follow-up refactors not done.

## Self-improvement
After I correct you, propose an update to this file (or the relevant doc in `docs/agents/`) so the same mistake doesn't recur. Don't update without my okay.

## Project context (WHAT / WHY / HOW)
<!-- Fill in. Keep tight. Pointers > copies. -->
- **What**: <one-line description of the system; key modules and their purpose>
- **Why**: <product purpose; the constraint that drives architecture>
- **How**: <stack; entrypoints; how to run, test, lint; where deeper docs live>

## Tooling (defer to these — don't reimplement in prose)
- Formatter — run before commit; never argue with its output
- Linter — fix lints, don't disable them
- Tests — full suite must pass before "done"

## Deeper docs (read on demand, not every session)
<!-- Pointers, not embedded content. Use file:line refs instead of pasting code. -->
- `docs/agents/architecture.md` — system map, module boundaries
- `docs/agents/testing.md` — test patterns, fixtures, integration setup
- `docs/agents/domain-glossary.md` — ubiquitous language for this domain