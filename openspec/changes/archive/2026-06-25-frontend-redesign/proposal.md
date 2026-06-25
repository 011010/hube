# Proposal: Frontend UI Redesign

## Intent

Hube's UI is functional but visually generic — Tailwind dashboard pattern with system-ui font, flat gray surfaces, Unicode icons, and no motion. The redesign gives it a **distinctive visual identity** that makes a personal hub feel like *your space*, not a corporate tool.

## Scope

### In Scope
- Warm ambient color system (Solaris direction — OKLCH warm dark base)
- Typography: Instrument Sans (headings) + Inter (body) + JetBrains Mono (code)
- Icon migration: Unicode → Lucide React (tree-shaken)
- Atom redesign: Button, Badge, Modal + new Input/Select/Textarea/Spinner/IconButton
- New molecules: EmptyState, Toast, PageHeader, Card, StatCard
- Layout rework: all 11 pages, sidebar overhaul, global layout wrapper
- CSS-native motion: page transitions, micro-interactions, staggered reveals
- `App.css` removal, `index.css` theme tokens

### Out of Scope
- Backend or data model changes
- New features or pages
- Tauri config changes
- CI/CD or deployment changes
- Test infrastructure (pre-existing gap)

## Capabilities

### New Capabilities
- `visual-theme`: Warm ambient color tokens, font system, dark/light mode refinements
- `ui-component-library`: Reusable Input/Select/Textarea/Spinner/IconButton/EmptyState/Toast/Card atoms & molecules
- `icon-system`: Lucide React migration, sidebar icon replacement, status/empty-state icons

### Modified Capabilities
None — no existing specs to modify.

## Approach

**Direction selected: Solaris (Warm Ambient).** Shift neutral grays to warm OKLCH undertones (hue ~60). Replace flat gray-950/900 surfaces with warm charcoal. Redesign components bottom-up: atoms first, then molecules, then pages. One page at a time to minimize regression risk.

Deliver in phases: (1) theme tokens + typography + fonts, (2) atom components, (3) molecule components + sidebar, (4) page-by-page rollout with layout rework.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `frontend/src/index.css` | Modified | Theme tokens, fonts, global styles |
| `frontend/src/components/atoms/*` | Modified | Button, Badge, Modal redesign |
| `frontend/src/components/molecules/*` | Modified | TaskModal, EventModal redesign |
| `frontend/src/components/organisms/*` | Modified | Sidebar, widget visual refresh |
| `frontend/src/pages/*` | Modified | All 11 pages layout/spacing rework |
| `frontend/src/App.tsx` | Modified | Layout wrapper, page transitions |
| `frontend/package.json` | Modified | New deps: lucide-react, fontsource packages |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Bundle size from fonts + icons | Med | Variable fonts (single file), Lucide tree-shaking |
| Tauri WebView rendering | Low | Avoid `backdrop-filter`, test CSS in Tauri |
| Regressions across 11 pages | Low | Page-by-page rollout, git branches |
| Light mode quality lag | Med | Design light mode in parallel with dark |

## Rollback Plan

Each page is an isolated change — revert a single page's files if needed. If theme tokens cause widespread issues, revert `index.css` and `package.json` to restore previous visual state.

## Dependencies

- `lucide-react` — icon library
- `@fontsource-variable/instrument-sans` — heading font
- `@fontsource-variable/inter` — body font
- `@fontsource/jetbrains-mono` — code font

## Success Criteria

- [ ] Every Unicode icon replaced with Lucide equivalent
- [ ] Font loading confirmed (no FOUT, all weights available)
- [ ] All 11 pages render with warm theme tokens (dark + light)
- [ ] Sidebar redesigned with Lucide icons + active indicators
- [ ] All form inputs use new Input/Select/Textarea atoms
- [ ] Empty states have illustration or icon + message (not text-only)
- [ ] Page transitions animate on route change
