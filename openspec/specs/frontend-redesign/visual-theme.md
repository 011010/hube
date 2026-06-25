# Visual Theme Specification

## Purpose

Define the warm ambient (Solaris) color token system, typography scale with Instrument Sans + Inter + JetBrains Mono, and dark/light mode integration that gives hube a distinctive visual identity away from generic Tailwind dashboard aesthetics.

## Requirements

### Requirement: Warm Surface Tokens

The system MUST define OKLCH-based warm surface tokens in its `@theme` block, replacing all hardcoded `bg-gray-*` background utilities used for page and component surfaces.

| Token | Dark Value | Light Value |
|-------|-----------|-------------|
| `surface-base` | `oklch(14% 0.015 60)` | `oklch(97% 0.01 60)` |
| `surface-elevated` | `oklch(18% 0.02 55)` | `oklch(94% 0.015 55)` |
| `surface-card` | `oklch(22% 0.025 55)` | `oklch(92% 0.015 55)` |
| `text-primary` | `oklch(92% 0.01 60)` | `oklch(18% 0.015 60)` |
| `text-secondary` | `oklch(65% 0.02 55)` | `oklch(45% 0.02 55)` |
| `border` | `oklch(28% 0.025 55)` | `oklch(85% 0.015 55)` |

#### Scenario: Dark mode page background

- GIVEN the user has dark mode active (`.dark` class on `<html>`)
- WHEN any page renders
- THEN the body background MUST be `var(--color-surface-base)` (warm charcoal)
- AND all card surfaces MUST use `var(--color-surface-card)`

#### Scenario: Light mode page background

- GIVEN the user has light mode active
- WHEN any page renders
- THEN the body background MUST be the warm cream value `oklch(97% 0.01 60)`

### Requirement: Accent Color Preservation

The 6 existing OKLCH accent classes MUST remain unchanged and functional: `theme-violet`, `theme-rose`, `theme-emerald`, `theme-sky`, `theme-orange`, plus default `indigo`.

#### Scenario: Accent switch

- GIVEN the user changes accent color to rose
- WHEN any component uses `var(--color-accent)`
- THEN the applied color MUST be `oklch(59.38% 0.24 12.25)`
- AND the hover variant MUST be `oklch(65% 0.22 12)`

### Requirement: Typography Scale

The system MUST load `@fontsource-variable/instrument-sans` (headings), `@fontsource-variable/inter` (body text), and `@fontsource/jetbrains-mono` (code) as npm packages. No external CDN calls SHALL be made for fonts.

#### Scenario: Font loading

- GIVEN the application loads
- WHEN CSS applies font-family
- THEN `--font-sans` MUST resolve to `'Instrument Sans', 'Inter', sans-serif`
- AND `--font-mono` MUST resolve to `'JetBrains Mono', monospace`
- AND no FOUT SHALL occur — fallback fonts MUST render immediately

### Requirement: Theme Token Integration

The system MUST apply warm surface tokens to all page backgrounds and component surfaces, and SHALL remove references to hardcoded `bg-gray-950`, `bg-gray-900`, `bg-gray-800` as surface values in components.

#### Scenario: Surface migration

- GIVEN a component currently using `bg-gray-900 border border-gray-800`
- WHEN that component renders
- THEN the background MUST be `var(--color-surface-elevated)`
- AND the border MUST be `var(--color-border)`

## Acceptance Criteria

- [ ] All `@theme` tokens in `index.css` use OKLCH with warm hue (55-60)
- [ ] 6 accent classes unchanged and functional with warm base
- [ ] `pnpm run build` succeeds with all 3 fontsource packages
- [ ] Zero `bg-gray-950`, `bg-gray-900`, `bg-gray-800` surface references remain in components
- [ ] No external font CDN requests appear in DevTools network tab
- [ ] Both dark and light modes render with warm undertones
