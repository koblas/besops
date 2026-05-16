---
paths:
  - "frontend/**"
---
# React Frontend Rules

Every UI change passes two lenses: **Jony Ive** (craft, restraint) and **Don Norman** (usability, error tolerance). The product is **self-service** — no human will rescue the user.

## Ive — design

- Every spacing, radius, weight, and timing is a decision. Justify it or change it.
- Subtract before you add. Best feature is often a removed one.
- One component, one job. If you describe it with "and," split it.
- Motion communicates causality, not decoration.
- Be honest: no fake progress, no infinite spinners, no placeholder content shipped.
- Restraint: neutral palette + one accent, minimal type sizes.
- Empty, error, and 404 states are part of the design.

## Norman — usability

- Clickable looks clickable. Non-clickable doesn't.
- Label behavior — don't make users discover it. Icon-only buttons get tooltips.
- Every action gets feedback within 100ms.
- UI structure mirrors task structure. 4-step task = 4 steps.
- Make wrong actions impossible (disable), not just discouraged.
- Same action looks the same everywhere.
- Assume mis-clicks. Confirm destructive actions. Allow undo. Preserve input on error.
- Every screen answers: *what do I do next?* and *did it work?*

## Self-service rules

1. Never strand the user. Every screen has an obvious next action; empty states include a CTA.
2. Plain language, not jargon. Buttons are verbs the user would say out loud.
3. Errors explain the fix, not the problem. ("Phone needs 10 digits — you entered 9.")
4. Progressive disclosure. 80% path first; advanced behind "More."
5. Any page works cold — no required context from prior screens.
6. Prefer undo over confirm dialogs. If confirming, spell out what's lost.

## React practices

- Name components by user intent, not implementation. `CancelSubscriptionButton`, not `RedButton`.
- One job per component. >5 props or mixed concerns → split.
- State at the lowest altitude that works.
- Loading, empty, error, success are all real states — implement all four.
- Forms preserve input on error; errors inline beside the field.
- Accessibility is part of done: semantic HTML, keyboard reachable, visible focus, labels associated, color never the only signal.
- No magic numbers — use tokens. If a value isn't in the system, fix the system.
- Test user flows, not hook internals.

## Done checklist

- [ ] First-time user can complete this unassisted
- [ ] Affordances are honest (clickable ↔ looks clickable)
- [ ] Feedback within 100ms on every action
- [ ] Loading / empty / error / success all designed
- [ ] Error messages are actionable
- [ ] Destructive actions are recoverable or clearly confirmed
- [ ] Nothing remains that isn't earning its place
- [ ] Matches existing patterns, or new pattern is justified
- [ ] Keyboard + screen-reader sensible

## When in doubt

Simpler. Fewer words. One less step. Show where the user is and what happens next. Would it still work tired, on a phone, in a hurry?