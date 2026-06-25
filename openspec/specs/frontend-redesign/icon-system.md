# Icon System Specification

## Purpose

Define Lucide React integration, complete migration from Unicode symbols to first-class React icon components, and establish consistent icon usage patterns across all components and pages.

## Requirements

### Requirement: Library Integration

The system MUST use `lucide-react` as the sole icon library. Imports MUST use named per-icon imports (`import { Plus } from 'lucide-react'`) to enable tree-shaking. No barrel imports or dynamic icon resolution SHALL be used.

#### Scenario: Tree-shaking verification

- GIVEN the production build completes
- WHEN bundle analysis runs
- THEN ONLY explicitly imported Lucide icons MUST appear in the bundle
- AND unused icons MUST be tree-shaken away

### Requirement: Sidebar Icon Mapping

All 11 sidebar navigation items MUST use specific Lucide icons, replacing the current Unicode symbols.

| Route | Label | Unicode (old) | Lucide (new) |
|-------|-------|---------------|--------------|
| `/` | Dashboard | `⊞` | `LayoutDashboard` |
| `/launcher` | Apps | `◈` | `Grid3X3` |
| `/tasks` | Tasks | `✓` | `ListChecks` |
| `/calendar` | Calendar | `◷` | `Calendar` |
| `/notes` | Notes | `✎` | `FileText` |
| `/projects` | Projects | `◫` | `FolderKanban` |
| `/ai` | IA | `✦` | `Brain` |
| `/wishlist` | Wishlist | `♡` | `Heart` |
| `/network` | Network | `⊛` | `Network` |
| `/docs` | Docs | `☰` | `BookOpen` |
| `/settings` | Settings | `⚙` | `Settings` |

#### Scenario: Sidebar renders Lucide icons

- GIVEN the Sidebar component mounts
- WHEN it renders all nav items
- THEN each item MUST use a Lucide React component (e.g., `<LayoutDashboard size={18} />`)
- AND zero Unicode symbols (`⊞`, `◈`, `✓`, `◷`, `✎`, `◫`, `✦`, `♡`, `⊛`, `☰`, `⚙`) SHALL appear in the rendered HTML

### Requirement: Unicode Removal Completeness

Every Unicode icon character in the codebase MUST be replaced with a Lucide React equivalent. This includes sidebar items, button decorations, modal close buttons (`✕` → `<X />`), empty state illustrations, status indicators, and any decorative symbols in page components.

#### Scenario: Modal close button

- GIVEN a Modal renders with `open={true}`
- WHEN the close button renders
- THEN it MUST use `<X size={16} />` from `lucide-react`
- AND the Unicode `✕` character MUST NOT be present

#### Scenario: Status indicators

- GIVEN a success/warning/error message renders
- WHEN a status indicator displays
- THEN it MUST use `CheckCircle`, `AlertTriangle`, or `XCircle` from Lucide
- AND emoji or Unicode characters MUST NOT be used for status

#### Scenario: Empty state icons

- GIVEN an EmptyState component renders (e.g., "No tasks yet")
- WHEN the icon slot is populated
- THEN it MUST render a Lucide icon at 40-48px size with reduced opacity

### Requirement: Consistent Icon Sizing

Icons SHALL use consistent sizes throughout the application: sidebar navigation 18px, button icons 16px, empty state 40-48px, status indicators 16-18px.

#### Scenario: Button icon size consistency

- GIVEN any Button with an `icon` prop renders
- WHEN the icon renders
- THEN the Lucide icon MUST receive `size={16}` for `size="sm"` or `size={16}` for `size="md"`
- AND all icons of the same button size MUST be the same pixel size

## Acceptance Criteria

- [ ] Zero Unicode icon characters remain in `frontend/src/` (confirmed via `grep -r '[\u2200-\u22FF\u25A0-\u25FF\u2600-\u27BF]' frontend/src/`)
- [ ] Bundle analysis confirms `lucide-react` is tree-shaken (only used icons in bundle)
- [ ] Sidebar renders 11 distinct Lucide icons matching the mapping table
- [ ] All Modal close buttons use Lucide `X` icon
- [ ] EmptyState components across Tasks, Notes, Wishlist, Projects, AI use Lucide icons
- [ ] Icon sizes are consistent: sidebar 18px, buttons 16px, empty state 40-48px
