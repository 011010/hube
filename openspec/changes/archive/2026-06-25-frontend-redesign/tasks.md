# Tasks: Frontend UI Redesign (Solaris)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~1100–1500 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (Theme+Atoms) → PR 2 (Molecules+Sidebar+Widgets) → PR 3 (Pages+Layout+Polish) |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

```
Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High
```

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Deps + Theme Tokens + Fonts + Atom Redesigns | PR 1 | Base = feature/frontend-redesign. Self-contained: new index.css, 5 new atoms, 3 modified atoms, main.tsx |
| 2 | New Molecules + Sidebar + Widgets + Modals | PR 2 | Base = PR 1 branch. Uses atoms from PR 1. EmptyState, Toast, Card, Sidebar Lucide, widget tokens, TaskModal/EventModal |
| 3 | All 11 Pages + App Layout + Polish | PR 3 | Base = PR 2 branch. Page-by-page rollout, App.tsx transitions, light mode tuning, cleanup |

## Phase 0: Dependencies

- [x] 0.1 Install `@fontsource-variable/instrument-sans`, `@fontsource-variable/inter`, `@fontsource/jetbrains-mono`, `lucide-react` in `frontend/package.json`

## Phase 1: Tokens + Fonts

- [x] 1.1 Replace `index.css` body/root styles with warm `@theme` block (8 surface tokens, dark + light), keep accent classes intact
- [x] 1.2 Add fontsource imports to `main.tsx` — Instrument Sans, Inter, JetBrains Mono
- [x] 1.3 Delete `App.css`, update `App.tsx` to remove its import

## Phase 2: Atom Redesigns

- [x] 2.1 **Button.tsx**: add `icon`/`iconRight`/`loading` props, add `outline` variant and `lg` size, warm hover states with token classes
- [x] 2.2 **Badge.tsx**: replace hardcoded `bg-gray-700`/`bg-emerald-900` with OKLCH warm values, add dot indicator
- [x] 2.3 **Modal.tsx**: add scale/fade CSS transition, Lucide `X` close button, `surface-card` + `border` tokens, focus trap
- [x] 2.4 Create **Input.tsx**: warm input with `error`/`label`/`helperText`/`prefix`/`suffix` props, accent focus ring, error message slot
- [x] 2.5 Create **Select.tsx**: warm select with `error`/`label`/`placeholder`/`options` props, accent focus ring, custom ChevronDown icon
- [x] 2.6 Create **Textarea.tsx**: warm textarea with `error`/`label`/`helperText` props, accent focus ring, auto-resize option
- [x] 2.7 Create **Spinner.tsx**: CSS-only rotation with SVG, `sm`/`md`/`lg` sizes
- [x] 2.8 Create **IconButton.tsx**: square icon-only button, `ghost`/`outline`/`solid` variants, `sm`/`md`/`lg` sizes

## Phase 3: New Molecules

- [x] 3.1 Create **EmptyState.tsx**: icon (Lucide 40-48px) + title + description + optional action slot
- [x] 3.2 Create **Toast.tsx**: auto-dismiss 4s, 4 variants (success/error/info/warning), fade-out transition
- [x] 3.3 Create **PageHeader.tsx**: title + description + optional actions slot
- [x] 3.4 Create **Card.tsx**: generic card container with `surface-elevated` + `border` tokens
- [x] 3.5 Create **StatCard.tsx**: numeric stat card, replaces inline ones in Dashboard

## Phase 4: Sidebar + Lucide

- [x] 4.1 Rewrite **Sidebar.tsx**: replace 11 Unicode icons with Lucide named imports (LayoutDashboard, Grid3X3, ListChecks, Calendar, FileText, FolderKanban, Brain, Heart, Network, BookOpen, Settings)
- [x] 4.2 Add accent active indicator, warm surface tokens, collapse-ready structure

## Phase 5: Widgets

- [x] 5.1 **FinanceWidgets.tsx**: migrate `bg-gray-900 border-gray-800` → `surface-elevated` / `border` tokens
- [x] 5.2 **CardTrackerWidget.tsx**: migrate `bg-gray-900 border-gray-800` → `surface-elevated` / `border` tokens

## Phase 6: Task/Event Modals

- [x] 6.1 **TaskModal.tsx**: replace raw `<input>`/`<textarea>`/`<select>` with Input/Select/Textarea atoms
- [x] 6.2 **EventModal.tsx**: replace raw `<input>`/`<textarea>` with Input/Textarea atoms

## Phase 7: Page Rollout

- [x] 7.1 **Dashboard**: use StatCard, surface tokens, Lucide icons in Quick Access
- [x] 7.2 **Tasks**: surface tokens, Lucide `Trash2` + `Plus`, EmptyState for zero tasks
- [x] 7.3 **Calendar**: Lucide `ChevronLeft`/`ChevronRight` for nav, surface tokens
- [x] 7.4 **Notes**: Lucide icons, surface tokens, EmptyState
- [x] 7.5 **Projects**: Modal + Input atoms, Lucide icons, EmptyState
- [x] 7.6 **AI**: Lucide `Sparkles`, surface tokens
- [x] 7.7 **Wishlist**: Modal + Input atoms, Lucide icons
- [x] 7.8 **Network**: Lucide icons, surface tokens
- [x] 7.9 **Docs**: surface tokens only
- [x] 7.10 **Launcher**: surface tokens
- [x] 7.11 **Settings**: Input atoms, surface tokens
- [x] 7.12 **Project Detail**: Lucide icons, surface tokens

## Phase 8: App Layout

- [x] 8.1 **App.tsx**: add `animate-fadeIn` page transition wrapper, polish layout container, confirm no `App.css` import

## Phase 9: Polish + Cleanup + Review

- [x] 9.1 Tune light mode contrast: verify `text-secondary` oklch values, adjust cream surface if needed
- [x] 9.2 Verify `@theme` light variant works with `@custom-variant dark` syntax
- [x] 9.3 Self-review: grep for remaining `bg-gray-9` / Unicode icons, verify tree-shaking
- [x] 9.4 Update README with Solaris design system notes
