## Exploration: Frontend UI Redesign

### Current State

The current frontend is **functional but visually generic**. It follows a standard Tailwind dashboard pattern—dark gray backgrounds (`bg-gray-950`, `bg-gray-900`), flat surfaces with thin gray borders (`border-gray-800`), system-ui font, and Unicode symbols for navigation icons. The layout is a conventional sidebar + main content split with no visual hierarchy differentiation between pages. The theme system (dark/light + 6 OKLCH accent colors) is well-architected but underutilized—accents appear primarily on buttons, the AI avatar, and link hover states. The HubeLogo SVG (hub-and-spoke design) is the most distinctive visual element.

**What works well:**
- Clean atomic design separation (atoms/molecules/organisms)
- Proper TypeScript types across all entities
- Consistent hooks layer with TanStack Query
- Theme context with dark/light + 6 accent colors (OKLCH-based)
- Good keyboard handling and focus outlines
- Modern stack (React 19, Tailwind v4 Vite plugin, react-router-dom v7)
- No bloated dependencies—lean bundle
- The hub-and-spoke logo SVG is distinctive and worth keeping/evolving
- Calendar grid algorithm and layout is solid

**What needs improvement:**
- Typography is completely generic (`system-ui, -apple-system, sans-serif`)
- Icon system is ad-hoc Unicode symbols (⊞, ◈, ✓, ◷, ✎, ◫, ✦, ♡, ⊛)
- No motion/animation—zero transitions or micro-interactions
- Flat color palette—all gray-950/900/800/700 with no gradients, textures, or depth
- Sidebar is basic text + Unicode icons
- Empty states are text-only messages with no illustration or personality
- Loading states are mostly text ("Loading...") or missing—only Finance/CardTracker have skeletons
- Form inputs are repetitive inline className patterns—no global `Input`, `Select`, `Textarea` components
- Page layouts are inconsistent (some use `p-8`, others use `p-6`, some full-height flex)
- Modal component is functional but visually plain
- No toast/notification system exists
- The Docs page is a massive single-file component (~940 lines) that should be split
- ErrorBoundary is minimal
- No global layout wrapper—each page handles its own spacing

### Affected Areas

All frontend files will be touched in some capacity. Major categories:

- `frontend/src/index.css` — Typography system, CSS variables, global styles, dark mode refinements
- `frontend/src/App.tsx` — Layout wrapper, page transition animations
- `frontend/src/App.css` — Likely deletable (vestigial, contains placeholder Vite demo styles)
- `frontend/index.html` — Font loading, meta tags
- `frontend/src/contexts/ThemeContext.tsx` — Minor enhancements for redesign
- `frontend/src/components/atoms/*` — Button, Badge, Modal need visual overhaul; new atoms: Input, Select, Textarea, IconButton, Spinner
- `frontend/src/components/molecules/*` — TaskModal, EventModal redesigned; new molecules: Toast/Notification, EmptyState, Card, PageHeader
- `frontend/src/components/organisms/*` — Sidebar (major overhaul), FinanceWidgets, CardTrackerWidget (visual refresh)
- `frontend/src/pages/*` — All 11 pages need layout/spacing rework and visual refreshes
- `frontend/src/store/` — (Currently empty) May add UI state slices (sidebar collapse, toasts)
- `frontend/package.json` — New dependencies: icon library, animation library, fonts

### Design Directions

#### Direction A: "Instrument Panel" — Industrial Precision

**Mood:** Precision tooling, cockpit instrumentation, data-rich environment. Think high-end dashboard design with technical credibility.

**Feel:** Dark-first, information-dense, monochromatic with sharp accent punches. Structured grids, tabular data aesthetics, subtle glows on active elements. Sidebar behaves like a mixer board—dense, efficient, with visual indicators.

**Typography:**
- Display/Headings: **IBM Plex Mono** or **JetBrains Mono** — monospace for technical authority
- Body: **Inter** (refined execution, wide weight range)
- Rationale: Monospace for headings communicates "tool/terminal/precision"—perfect for a developer's personal hub. Inter is the most-produced neutral sans with excellent legibility at small sizes.

**Color:** Dark-slate base (not pure black), neon accent for active states, subtle monochromatic surface hierarchy using luminance rather than hue.

**Strength:** Plays to the power-user audience, emphasizes the data nature of the app, distinctive without being trendy.

#### Direction B: "Solaris" — Warm Ambient

**Mood:** Warm, organic, atmospheric. Think personal space rather than corporate tool. Dark with amber/warm undertones, soft glows, depth through light and shadow.

**Feel:** Low contrast, comfortable, restful. Backgrounds shift from deep warm charcoal rather than cool neutral gray. Cards have subtle inner glow, borders are luminous rather than hard. Sidebar feels like a control panel on a spaceship at night.

**Typography:**
- Display/Headings: **Instrument Sans** or **Plus Jakarta Sans** — warm, rounded humanist
- Body: **Satoshi** or **Inter** — neutral, warm-leaning
- Rationale: Warm sans-serifs for a tool that lives in your daily workflow—approachable, personal, not cold.

**Color:** Warm dark base (`oklch(15% 0.02 60)`), amber/copper accents, rose gold as secondary. Light mode shifts to warm cream/paper tones.

**Strength:** Distinctive and cozy—memorable because no one else designs dark mode with warm undertones. Feels like "your space."

#### Direction C: "Lucid" — Glass + Precision Minimalism

**Mood:** Crystal clear, layered, premium. Think modern financial dashboards and creative tools. Spatial depth through transparencies and blur.

**Feel:** Generous whitespace, translucent card surfaces (glass morphism), sharp typography against blurred backgrounds. Sidebar becomes a translucent floating panel. Content feels like it sits on multiple glass planes.

**Typography:**
- Display/Headings: **Editorial New** or **Departure Mono** for contrast, or **Space Grotesk** (only if used with extreme distinction)
- Body: **Inter** or **Onest**
- Rationale: Headings need character to stand out against glass surfaces. A slight grotesk or even editorial serif for large headings provides the "premium tool" feel.

**Color:** True dark background, cards with `backdrop-blur` and subtle border highlights. Acrylic/glass aesthetic. Light mode uses frosted glass on warm white.

**Strength:** Premium feel, highly visual, separates hube from every other flat Tailwind dashboard. The glass effect is technically achievable with CSS and performs well on desktop/Tauri.

### Recommended Direction

**Direction B: "Solaris" (Warm Ambient)** is the strongest choice for hube because:

1. **Personality match:** A personal hub should feel like YOUR space, not a corporate dashboard. Warm dark mode with amber/rose undertones is inherently welcoming.
2. **Unforgettable:** Every other dashboard is cool dark gray + indigo. Warm dark is distinctive and memorable.
3. **Tauri-friendly:** No heavy browser APIs needed—no `backdrop-filter` reliance (glass effects in Lucid can be expensive on Tauri WebView). Warm ambient uses pure CSS gradients and box-shadows.
4. **Accent system synergy:** The 6 existing accent colors (indigo, violet, rose, emerald, sky, orange) work beautifully against a warm dark base—each accent reads clearly and distinctly.
5. **Extension:** Warm ambient scales well—financial data looks serious, notes feel cozy, AI feels intelligent. It's versatile across all 11 pages.

### Typography Recommendation

**Primary:** `Instrument Sans` (headings, buttons, nav) — warm rounded humanist with excellent variable weight range
**Fallback:** `Inter` as body/code—proven, legible, pairs naturally with Instrument Sans
**Code:** `JetBrains Mono` — for notes code blocks, settings values, docs

**Loading strategy:** `@fontsource-variable/instrument-sans`, `@fontsource-variable/inter`, `@fontsource/jetbrains-mono` — loaded as npm packages (no external CDN, Tauri-friendly, works offline).

These fonts are:
- Distinctive without being distracting
- Variable-weight (load one file, get all weights)
- Available as npm packages (no Google Fonts dependency)
- Performance-friendly (Latin subset, variable = single file per family)

### Color Strategy

**Dark mode (primary):**
- Base surface: `oklch(14% 0.015 60)` — warm dark charcoal
- Elevated surface: `oklch(18% 0.02 55)` — one level up with warm tint
- Card surface: `oklch(22% 0.025 55)` — interactive surfaces
- Text primary: `oklch(92% 0.01 60)` — warm white
- Text secondary: `oklch(65% 0.02 55)` — warm gray
- Borders: `oklch(28% 0.025 55)` — subtle warm lines
- Accent: Dynamic via `--color-accent` OKLCH variable (existing system)

**Light mode (secondary):**
- Base: `oklch(97% 0.01 60)` — warm cream
- Elevated: `oklch(94% 0.015 55)` — subtle warm
- Text: `oklch(18% 0.015 60)` — warm near-black

The key insight: shift all neutral colors to have a slight **warm hue** (60 on the hue wheel = yellow-orange range). This is invisible as "color" but felt as "warmth." The existing 6 accent colors remain untouched.

### Icon System Recommendation

**Lucide React** — the most comprehensive open-source icon library with:
- 1500+ consistent icons (24px stroke-based, matching all hube needs)
- First-class React package (`lucide-react`) with tree-shaking
- MIT license, no attribution
- Perfect for Tauri—no external requests, no CDN
- Consistent stroke-width = cohesive visual system

Migration: Replace all Unicode symbols in the sidebar, replace emoji icons in Launcher, add icons for empty states, buttons, and status indicators.

**Fallback:** If bundle size is a concern, use `@phosphor-icons/react` (similar quality, slightly smaller) or hand-pick SVGs. Lucide's tree-shaking makes this acceptable—each imported icon adds ~1KB gzipped.

### Animation/Motion Approach

**Library:** `motion` (formerly Framer Motion) — v12 is stable, tree-shakable, works with React 19.

But for this redesign, **CSS transitions + Tailwind will cover 90% of needs** without a dependency:

1. **Page transitions:** CSS `@view-transition` / `view-transition-name` for SPA route changes (native browser API, no library needed)
2. **Micro-interactions:** Tailwind `transition-all`, `hover:`, `group-hover:` for button and card hover states
3. **Staggered reveals:** Pure CSS with `animation-delay` on staggered children
4. **Sidebar:** CSS `translate-x` transition for collapse/expand
5. **Modal/overlay:** CSS `opacity` + `scale` transitions

**Add motion library only for:** AI chat message stream-in animations, dashboard widget mount animations, and calendar month transitions. Start without it—add only when needed.

### Technical Approach

#### Tailwind Config Changes

Tailwind v4 uses CSS-first configuration (no `tailwind.config.js`). Changes go in `index.css`:

```css
/* New custom theme tokens */
@theme {
  --font-sans: 'Instrument Sans', 'Inter', sans-serif;
  --font-mono: 'JetBrains Mono', monospace;
  --color-surface-base: oklch(14% 0.015 60);
  --color-surface-elevated: oklch(18% 0.02 55);
  --color-surface-card: oklch(22% 0.025 55);
  --color-surface-warm: oklch(28% 0.025 55);
  --color-text-primary: oklch(92% 0.01 60);
  --color-text-secondary: oklch(65% 0.02 55);
}
```

#### New Dependencies

**Required:**
- `lucide-react` — icon system
- `@fontsource-variable/inter` — body font
- `@fontsource-variable/instrument-sans` — heading font
- `@fontsource/jetbrains-mono` — monospace font

**Optional (add later):**
- `motion` — complex animations (AI chat, dashboard transitions)
- `sonner` or `react-hot-toast` — lightweight toast system

#### New Files to Create

```
src/
  components/
    atoms/
      Input.tsx          ← Form input abstraction
      Select.tsx         ← Form select abstraction
      Textarea.tsx       ← Form textarea abstraction
      Spinner.tsx        ← Loading spinner
      IconButton.tsx     ← Icon-only button (for toolbars)
    molecules/
      EmptyState.tsx     ← Illustrated empty state
      Toast.tsx          ← Toast notification system
      PageHeader.tsx     ← Consistent page header
      Card.tsx           ← Base card component
      StatCard.tsx       ← Dashboard stat card
  styles/
    fonts.css            ← @fontsource imports
    tokens.css           ← Custom CSS variables
```

#### Files to Modify

| File | Changes |
|------|---------|
| `index.css` | Replace `@theme` block, add warm surface tokens, font definitions, remove vestigial rules |
| `App.css` | Delete entirely (vestigial Vite template CSS, unused in production) |
| `index.html` | Add preconnect for fonts if using CDN fallback |
| `App.tsx` | Add layout wrapper with page transition container, pass theme-aware class |
| `contexts/ThemeContext.tsx` | Add accent color CSS custom properties for broader use, surface tone support |
| `components/atoms/Button.tsx` | Redesign with warm colors, add icon slot, ghost/outline variants |
| `components/atoms/Badge.tsx` | Redesign with accent-aware colors |
| `components/atoms/Modal.tsx` | Add entrance animation, warm styling, better backdrop |
| `components/organisms/Sidebar.tsx` | Full redesign: Lucide icons, warm colors, active indicator, collapsible |
| `components/organisms/FinanceWidgets.tsx` | Visual refresh with warm cards |
| `components/organisms/CardTrackerWidget.tsx` | Visual refresh with warm cards |
| `components/molecules/TaskModal.tsx` | Redesign with new form components |
| `components/molecules/EventModal.tsx` | Redesign with new form components |
| `components/molecules/EventChip.tsx` | Visual refresh |
| `pages/Dashboard/index.tsx` | Redesign with StatCard, warm cards, widget layout |
| `pages/Tasks/index.tsx` | Redesign task list, add empty state |
| `pages/Notes/index.tsx` | Redesign note list, editor panes, add empty state |
| `pages/Calendar/index.tsx` | Redesign calendar grid cells, header |
| `pages/AI/index.tsx` | Redesign chat bubbles, input area, add empty state |
| `pages/Launcher/index.tsx` | Redesign app cards |
| `pages/Projects/index.tsx` | Redesign project cards, modal, add empty state |
| `pages/Settings/index.tsx` | Redesign settings sections with new components |
| `pages/Wishlist/index.tsx` | Redesign wishlist items, modal, add empty state |
| `pages/Network/index.tsx` | Minimal visual refresh |
| `pages/Docs/index.tsx` | Refactor into smaller files, visual refresh |
| `package.json` | Add new dependencies |
| `vite.config.ts` | No changes needed (fonts are npm packages, not external) |

### Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Bundle size increase from fonts + icons | Medium | Low | Tree-shaking, font subsetting, variable fonts (single file per family) |
| Light mode quality suffers from dark-first focus | Medium | Medium | Design light mode in parallel—simple cream/paper tones (lower effort) |
| Tauri WebView rendering differences | Low | Medium | Avoid `backdrop-filter: blur()` until V8 renders it acceptably; test all CSS in Tauri during implementation |
| Regressions on existing functionality | Low | High | Each page redesign = isolated, one at a time; use git branches |
| Design consistency across 11 pages | Medium | Medium | Create a component library (atoms first, then molecules) before touching pages |
| No frontend tests to catch regressions | High | High | This is a pre-existing risk. Add visual/integration tests for critical flows if time permits |
| Dependency churn (React 19, Tailwind v4 are new) | Low | Low | Already using latest versions; redesign should stay current |

### Ready for Proposal

Yes. The exploration is comprehensive enough to move to `sdd-propose`. The recommended direction is **"Solaris" (Warm Ambient)** with **Instrument Sans + Inter** typography, **Lucide** icons, **CSS-native motion** (library only if needed), and the existing OKLCH accent system preserved. The change is well-scoped: it's visual/component work with zero backend changes, zero routing changes, and no data model modifications.
