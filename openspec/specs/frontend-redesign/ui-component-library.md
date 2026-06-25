# UI Component Library Specification

## Purpose

Define redesigned atom, molecule, and organism components using the Solaris warm theme tokens. Standardize form input components, introduce missing patterns (EmptyState, Toast, Spinner), and refresh all existing components for visual consistency across 11 pages.

## Requirements

### Requirement: Button (Redesigned)

`Button` MUST support `primary`, `ghost`, `danger`, and `outline` variants. `primary` MUST use `--color-accent` as background. All buttons MUST accept an optional `icon` prop (ReactNode) placed before children.

#### Scenario: Primary button with icon

- GIVEN a Button with `variant="primary"` and `icon={<Plus size={16} />}`
- WHEN it renders
- THEN background MUST be `var(--color-accent)` with white text
- AND the Plus icon MUST render before the button label

#### Scenario: Outline variant

- GIVEN a Button with `variant="outline"`
- WHEN it renders
- THEN it MUST have a `1px` border using `var(--color-border)` and transparent background
- AND text MUST use `var(--color-text-secondary)`

### Requirement: Badge (Redesigned)

`Badge` MUST use accent-aware warm colors instead of hardcoded `bg-gray-700` / `text-gray-300`. Variants: `default`, `success`, `warning`, `danger` — each with OKLCH values matching the warm palette.

#### Scenario: Success badge

- GIVEN a Badge with `variant="success"` and `label="Done"`
- WHEN it renders
- THEN it MUST show a warm green background and light green text
- AND SHOULD NOT use hardcoded `bg-emerald-900` / `text-emerald-300`

### Requirement: Modal (Redesigned)

`Modal` MUST animate entrance with opacity + scale transition (200ms CSS transition). The close button MUST use Lucide `X` icon. The backdrop MUST use the warm-tinted `surface-base` overlay.

#### Scenario: Modal open animation

- GIVEN the user triggers a modal (`open={true}`)
- WHEN the modal renders
- THEN backdrop MUST fade in from opacity 0 to 0.6 over 200ms
- AND content panel MUST scale from 0.95 to 1.0

### Requirement: New Form Atoms

`Input`, `Select`, and `Textarea` atoms MUST provide consistent warm styling with `--color-accent` focus ring and an error state (red border + error message slot).

#### Scenario: Input with error

- GIVEN an Input with `error="Title is required"`
- WHEN it renders
- THEN the input border MUST be red (`oklch(55% 0.2 25)`)
- AND the error message MUST display below the input in red text

### Requirement: New Error/Loading Molecules

`Spinner` MUST animate with CSS-only rotation and accept `size` (`sm`/`md`/`lg`). `EmptyState` MUST accept `icon` (Lucide), `title`, `description`, and optional `action` (ReactNode). `Toast` MUST auto-dismiss after 4 seconds and support `success`/`error`/`info`/`warning` variants.

#### Scenario: Toast auto-dismiss

- GIVEN a Toast with `variant="success"` fires
- WHEN 4 seconds elapse
- THEN the Toast MUST automatically dismiss with a fade-out transition

### Requirement: Sidebar (Redesigned)

`Sidebar` MUST use Lucide icons, warm surface colors, an accent-colored active indicator on the current nav item, and support future collapse/expand behavior.

#### Scenario: Active navigation item

- GIVEN the user navigates to `/tasks`
- WHEN the Tasks NavLink renders
- THEN the item MUST show `var(--color-accent)` background with white text and a Lucide icon
- AND other nav items MUST use `var(--color-text-secondary)` with warm hover state

### Requirement: Card Pattern (New + Migrated)

`Card` and `StatCard` molecules MUST use warm surface tokens. `FinanceWidgets` and `CardTrackerWidget` MUST migrate from hardcoded `bg-gray-900 border-gray-800` to use `surface-elevated` / `border` tokens.

#### Scenario: Finances use Card tokens

- GIVEN the FinanceWidgets component renders summary cards
- WHEN a summary card renders
- THEN it MUST use `var(--color-surface-elevated)` as background and `var(--color-border)` as border

## Acceptance Criteria

- [ ] Button, Badge, Modal, Spinner render with warm theme tokens
- [ ] TaskModal and EventModal use Input/Select/Textarea atoms for all form fields
- [ ] Sidebar renders all 11 nav items with Lucide icons and active indicator
- [ ] Empty states render on Tasks, Notes, Wishlist, Projects, AI pages
- [ ] Toast system is available for use across pages
- [ ] Zero hardcoded `bg-gray-*` surface values remain in component code
