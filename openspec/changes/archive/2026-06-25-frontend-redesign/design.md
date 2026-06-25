# Design: Frontend UI Redesign (Solaris)

## Technical Approach

Replace generic Tailwind dashboard aesthetics with a warm Solaris color system, custom typography, Lucide icons, and redesigned components — bottom-up from tokens to pages. Each page is an isolated change to minimize regression. No backend or data model changes.

## Architecture Decisions

| Option | Tradeoffs | Decision |
|--------|-----------|----------|
| Token source | Tailwind `@theme` block vs CSS custom properties in `index.css` | **`@theme` block** — Tailwind v4 resolves these into utilities like `bg-surface-card`. Falls back to `var(--color-surface-*)` for inline styles.  |
| Font strategy | `@fontsource` npm (bundled) vs CDN (`<link>`) | **fontsource npm** — works offline, no CDN dependency, Tauri-friendly. Fontsource v5 generates `@font-face` at build time. |
| Icon approach | Lucide named imports vs `react-icons` barrel | **Named imports from `lucide-react`** — tree-shaken by Vite/Rollup. No barrel files, no dynamic resolution. |
| Component structure | Atomic redesign in-place vs wrapper layer | **In-place redesign** — existing atoms get new props/classes; new atoms/molecules are new files. No wrapping layer needed. |
| Modal pattern | Existing inline modals vs new Modal atom | **New Modal atom** with animation + Lucide close. Inline modals (Projects, Wishlist) migrate to it. |
| Form inputs | Raw `<input>` vs new atom components | **New Input/Select/Textarea atoms** — consistent warm styling, error state, focus ring. Migrate all pages. |

## Token Architecture (Tailwind v4 `@theme` block)

```css
/* index.css — replaces current flat-gray tokens */
@import "tailwindcss";
@custom-variant dark (&:where(.dark, .dark *));

@theme {
  --font-sans: 'Instrument Sans', 'Inter', sans-serif;
  --font-mono: 'JetBrains Mono', monospace;

  --color-surface-base:    oklch(14% 0.015 60);
  --color-surface-elevated: oklch(18% 0.02 55);
  --color-surface-card:    oklch(22% 0.025 55);
  --color-text-primary:    oklch(92% 0.01 60);
  --color-text-secondary:  oklch(65% 0.02 55);
  --color-border:          oklch(28% 0.025 55);
}

@theme light { /* or html:not(.dark) overrides */
  --color-surface-base:    oklch(97% 0.01 60);
  --color-surface-elevated: oklch(94% 0.015 55);
  --color-surface-card:    oklch(92% 0.015 55);
  --color-text-primary:    oklch(18% 0.015 60);
  --color-text-secondary:  oklch(45% 0.02 55);
  --color-border:          oklch(85% 0.015 55);
}
```

Accent colors (6 classes) remain on `:root` / `.theme-*` as CSS custom properties — no change needed. `ThemeContext.tsx` already handles toggling `.dark` and accent classes, which aligns perfectly.

## Font Strategy

```json
{
  "dependencies": {
    "@fontsource-variable/instrument-sans": "^5.x",
    "@fontsource-variable/inter": "^5.x",
    "@fontsource/jetbrains-mono": "^5.x"
  }
}
```

Import in `main.tsx` (not `index.css`) for explicit ordering:

```ts
import '@fontsource-variable/instrument-sans'
import '@fontsource-variable/inter'
import '@fontsource/jetbrains-mono'
```

Then `font-sans` and `font-mono` in `@theme` resolve via Tailwind. No CDN. No FOUT if fallback `system-ui` renders first (acceptable — fontsource adds `font-display: swap`).

## Data Flow

```
ThemeContext
  └─ applies .dark / accent classes on <html>
     └─ @theme vars respond to .dark selector
        └─ components use bg-surface-card / text-primary / etc.

Sidebar nav click ─→ react-router <NavLink>
  └─ active class sets accent bg + white text
  └─ icon uses Lucide component (18px)

Form submit ─→ Input/Select/Textarea (controlled via prop)
  └─ error state: red border + error message slot
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `frontend/src/index.css` | Modify | Replace flat grays with warm `@theme` tokens; add light mode overrides |
| `frontend/src/App.css` | Delete | Unused template styles, all moved to index.css |
| `frontend/src/main.tsx` | Modify | Add fontsource imports |
| `frontend/package.json` | Modify | Add 3 fontsource packages + lucide-react |
| `frontend/src/components/atoms/Button.tsx` | Modify | Add `icon` prop, `outline` variant, warm hover states |
| `frontend/src/components/atoms/Badge.tsx` | Modify | Use OKLCH warm values instead of `bg-emerald-900` |
| `frontend/src/components/atoms/Modal.tsx` | Modify | Add scale/fade animation, Lucide `X` close button, warm surface tokens |
| `frontend/src/components/atoms/Input.tsx` | Create | Warm input with `error` prop, accent focus ring |
| `frontend/src/components/atoms/Select.tsx` | Create | Warm select with `error` prop |
| `frontend/src/components/atoms/Textarea.tsx` | Create | Warm textarea with `error` prop |
| `frontend/src/components/atoms/Spinner.tsx` | Create | CSS-only rotation, `sm`/`md`/`lg` sizes |
| `frontend/src/components/atoms/IconButton.tsx` | Create | Square icon-only button for close/delete actions |
| `frontend/src/components/molecules/EmptyState.tsx` | Create | Icon + title + description + optional action slot |
| `frontend/src/components/molecules/Toast.tsx` | Create | Auto-dismiss (4s), 4 variants, fade-out |
| `frontend/src/components/molecules/PageHeader.tsx` | Create | Title + description + optional actions slot |
| `frontend/src/components/molecules/Card.tsx` | Create | Generic card container with surface-elevated |
| `frontend/src/components/molecules/StatCard.tsx` | Create | Numeric stat card (replaces inline ones in Dashboard) |
| `frontend/src/components/molecules/TaskModal.tsx` | Modify | Use Input/Select/Textarea atoms |
| `frontend/src/components/molecules/EventModal.tsx` | Modify | Use Input/Select/Textarea atoms |
| `frontend/src/components/organisms/Sidebar.tsx` | Modify | Lucide icons, active indicator, warm tokens, account for collapse |
| `frontend/src/components/organisms/FinanceWidgets.tsx` | Modify | `surface-elevated` / `border` tokens |
| `frontend/src/components/organisms/CardTrackerWidget.tsx` | Modify | `surface-elevated` / `border` tokens |
| `frontend/src/pages/Dashboard/index.tsx` | Modify | Use StatCard, surface tokens, Lucide |
| `frontend/src/pages/Tasks/index.tsx` | Modify | Surface tokens, Lucide delete/plus, EmptyState |
| `frontend/src/pages/Calendar/index.tsx` | Modify | Lucide nav arrows, surface tokens |
| `frontend/src/pages/Notes/index.tsx` | Modify | Lucide icons, surface tokens, EmptyState |
| `frontend/src/pages/Projects/index.tsx` | Modify | Use Modal + Input atoms, Lucide icons, EmptyState |
| `frontend/src/pages/AI/index.tsx` | Modify | Lucide ✦ → Sparkles, surface tokens |
| `frontend/src/pages/Settings/index.tsx` | Modify | Use Input atoms, surface tokens |
| `frontend/src/pages/Wishlist/index.tsx` | Modify | Use Modal + Input atoms, Lucide icons |
| `frontend/src/pages/Network/index.tsx` | Modify | Lucide icons, surface tokens |
| `frontend/src/pages/Docs/index.tsx` | Modify | Surface tokens only (no icons to replace) |
| `frontend/src/pages/Launcher/index.tsx` | Modify | Surface tokens |
| `frontend/src/pages/Projects/Detail.tsx` | Modify | Lucide icons, surface tokens |
| `frontend/src/App.tsx` | Modify | Remove App.css import, add page transition wrapper |

**Total: ~30 files** (5 new atoms, 4 new molecules, 20+ modified)

## Component API Contracts

```tsx
// Atoms
interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'ghost' | 'danger' | 'outline'
  size?: 'sm' | 'md'
  icon?: ReactNode   // Lucide component, placed before children
}

interface BadgeProps {
  label: string
  variant?: 'default' | 'success' | 'warning' | 'danger'
}

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  error?: string
}

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  error?: string
}

interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  error?: string
}

interface ModalProps {
  open: boolean
  onClose: () => void
  title: string
  children: ReactNode
}

interface SpinnerProps {
  size?: 'sm' | 'md' | 'lg'
}

// Molecules
interface EmptyStateProps {
  icon?: ReactNode  // Lucide icon at 40-48px
  title: string
  description?: string
  action?: ReactNode
}

interface ToastProps {
  message: string
  variant?: 'success' | 'error' | 'info' | 'warning'
  onDismiss: () => void
}

interface StatCardProps {
  label: string
  value: string | number
  accent?: string  // optional accent color
}
```

## Icon Migration Map

| Location | Unicode (old) | Lucide (new) | Size |
|----------|---------------|--------------|------|
| Sidebar nav | ⊞ ◈ ✓ ◷ ✎ ◫ ✦ ♡ ⊛ ☰ ⚙ | LayoutDashboard, Grid3X3, ListChecks, Calendar, FileText, FolderKanban, Brain, Heart, Network, BookOpen, Settings | 18px |
| Modal close | ✕ | X | 16px |
| Delete buttons | ✕ | Trash2 | 14px |
| Check/cross status | ✓ | Check, X | 14px |
| Calendar nav | ‹ › | ChevronLeft, ChevronRight | 16px |
| Plus buttons | + | Plus | 16px |
| Arrow up/right | ↑ ↗ | ArrowUpRight | 14px |
| Back nav | ← | ArrowLeft | 16px |
| AI header | ✦ | Sparkles | 18px |
| Empty states | ◻ (text) | Inbox, FileText, Heart, FolderKanban, Network | 40-48px |

## Component Dependency Tree

```
Tokens (index.css)
 └── Atoms: Button, Badge, Modal, Input, Select, Textarea, Spinner, IconButton
      └── Molecules: EmptyState, Toast, PageHeader, Card, StatCard, TaskModal, EventModal
           └── Organisms: Sidebar, FinanceWidgets, CardTrackerWidget
                └── Pages: Dashboard, Tasks, Calendar, Notes, Projects, AI, etc.
                     └── App.tsx (layout + page transitions)
```

## Implementation Order

| Phase | What | Depends On |
|-------|------|------------|
| **0** | Install deps: fontsource packages, lucide-react | — |
| **1** | Tokens + fonts: `index.css` @theme, `main.tsx` imports, delete `App.css` | Phase 0 |
| **2** | Atom redesigns: Button, Badge, Modal + new Input/Select/Textarea/Spinner/IconButton | Phase 1 |
| **3** | New molecules: EmptyState, Toast, PageHeader, Card, StatCard | Phase 2 |
| **4** | Sidebar redesign: Lucide icons, warm tokens, active indicator | Phase 2 |
| **5** | Organisms: FinanceWidgets, CardTrackerWidget migrate to surface tokens | Phase 1 |
| **6** | Modal pages: TaskModal, EventModal use Input/Select/Textarea | Phase 2 |
| **7** | Page-by-page: Dashboard → Tasks → Calendar → Notes → Projects → Detail → AI → Wishlist → Network → Docs → Launcher → Settings | Phases 1-6 |
| **8** | App.tsx: page transition wrapper, remove App.css import | Phase 1 |

## Animation Strategy

- **Modal**: CSS transition only — `opacity` + `scale` (200ms ease-out) via Tailwind classes. No JS animation library.
- **Page transitions**: `animate-fadeIn` utility class on route `<div>` wrapper in `App.tsx`. CSS-only `@keyframes fadeIn`.
- **Hover states**: Use Tailwind `transition-colors` (standard), `hover:scale-105` for interactive cards, `hover:opacity-80` for icons.
- **Spinner**: CSS `@keyframes spin` — no dependency.
- **Staggered reveals**: Optional `animation-delay` via inline style or `delay-*` utilities for list items.

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Visual | Token application, font loading | Manual review in dark/light mode per page |
| Build | Token resolution, tree-shaking | `pnpm run build` — verify bundle, no dead code |
| Regression | All pages render | Visual check on all 11 pages across modes |

## Migration / Rollout

Phase 7 is page-by-page commits. Each page gets its own commit so reverts are isolated. The full diff across 30 files can be stacked as a single PR or chained PRs depending on total size.

## Open Questions

- [ ] Tauri WebView: test `oklch()` support and `@fontsource` in Tauri environment (should work with modern WebKit)
- [ ] Light mode quality: warm cream might need adjustment for contrast — verify `text-secondary` oklch values in light mode
- [ ] `@theme` light variant syntax: Tailwind v4 `@theme light {}` vs manual overrides — need to confirm which works with `@custom-variant dark`
