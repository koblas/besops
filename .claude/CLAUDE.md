Standards for generating and modifying code in this repo. Rooted in **Bryan Finster** (testing, continuous delivery) and **Martin Fowler** (design, refactoring).

Core idea: **code is finished only when it can be changed safely tomorrow.** Tests prove behavior. Design keeps change cheap.

## Prime directives
1. Optimize for ease of change, not cleverness.
2. Untested new behavior is unfinished. Write the test.
3. Every change small enough to merge to trunk today. If not, split it and show the sequence.
4. Don't mix refactors with behavior changes in one diff.

## Testing (Finster)
- **Test behavior, not implementation.** A pure refactor must not break tests. If it does, the test was wrong.
- **TDD for new logic**: red → green → refactor. Use Given/When/Then names for business behavior.
- **Coverage is a diagnostic, not a goal.** Every test must fail if the behavior breaks — otherwise it's noise. Cover negative and edge cases explicitly.
- **Pyramid, not hourglass**: many fast unit tests, few integration, minimal E2E. Unit suite runs in seconds.
- Test names should tell on-call what broke without reading the implementation.

## Change size & integration (Finster)
- Hours of work per change, not days.
- For incomplete features, prefer **feature flags**, **branch by abstraction**, or **expand/contract** migrations over long-lived branches.
- Never weaken or delete a test to make a change pass. Never silently change a public contract — propose a deprecation path.

## Design (Fowler)
**Four rules of simple design**, in priority order — code should:
1. Pass its tests.
2. Reveal intent (if a comment explains *what*, rename instead).
3. Have no duplicated knowledge.
4. Have the fewest elements. No abstraction without a concrete reason.

**Functions**: short, one level of abstraction, no flag arguments, few parameters (introduce a parameter object past 3–4), no hidden side effects.

**Modules**: single reason to change, high cohesion, low coupling. Business logic doesn't know about HTTP, SQL, or framework annotations.

**Names**: say what, why, how it's used. No abbreviations except universal ones. Booleans as predicates (`isReady`), functions as verbs (`calculateTax`).

**Comments**: explain *why* only. Acceptable: business rules, trade-offs, links to tickets, public API docs. Delete commented-out code — git remembers.

## Smells that signal "refactor now"
Duplicated code · long function / large class · long parameter list · divergent change · shotgun surgery · feature envy · data clumps · primitive obsession · switch on type code · speculative generality · mysterious name · middle man · inappropriate intimacy.

## Errors
Fail loudly at the boundary. Never swallow exceptions — handle, wrap with context, or rethrow. Validate at trust boundaries; inside, prefer types that make invalid states unrepresentable. Error messages include what was attempted, expected, and seen.

## Workflow per task
**Before**: state the behavior change in one sentence; identify the smallest change; list the tests.
**During**: failing test first; refactors and behavior changes in separate diffs; flag smells in the touched area.
**After**: confirm each new behavior has a test that fails without the new code; confirm the diff is trunk-mergeable today; note follow-up refactors not done.

## Don't
- Ship new behavior without a test.
- Chase coverage with assertion-free tests.
- Add interfaces, configs, or plug-ins for needs that don't yet exist.
- Mix unrelated changes in one diff.
- Leave commented-out code.

---

## Project conventions
<!-- Language, framework, formatter, test command, directory layout, domain glossary. -->