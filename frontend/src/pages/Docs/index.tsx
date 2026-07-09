import { Fragment, useEffect, useRef, useState } from 'react'
import { PanelLeft } from 'lucide-react'
import { IconButton } from '../../components/atoms/IconButton'
import { DocSection, Divider } from './DocSection'
import { docSections } from './content'

interface NavSection {
  id: string
  label: string
}

const SECTIONS: NavSection[] = [
  { id: 'overview',  label: 'Overview' },
  { id: 'tasks',     label: 'Tasks' },
  { id: 'notes',     label: 'Notes' },
  { id: 'folders',   label: 'Folders' },
  { id: 'calendar',  label: 'Calendar' },
  { id: 'projects',  label: 'Projects' },
  { id: 'apps',      label: 'Apps' },
  { id: 'wishlist',  label: 'Wishlist' },
  { id: 'diagrams',  label: 'Diagrams' },
  { id: 'finance',   label: 'Finance' },
  { id: 'cards',     label: 'Card Tracker' },
  { id: 'ai',        label: 'AI Agent' },
  { id: 'api',       label: 'API Reference' },
  { id: 'config',    label: 'Configuration' },
  { id: 'running',   label: 'Running' },
  { id: 'backup',    label: 'Backup & Export' },
  { id: 'security',  label: 'Security' },
]

export function DocsPage() {
  const [activeId, setActiveId] = useState<string>(SECTIONS[0].id)
  const [panelOpen, setPanelOpen] = useState(false)
  const contentRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const sectionEls = SECTIONS.map((s) => document.getElementById(s.id)).filter(Boolean) as HTMLElement[]
    if (sectionEls.length === 0) return

    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((e) => e.isIntersecting)
          .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)
        if (visible.length > 0) setActiveId(visible[0].target.id)
      },
      { root: contentRef.current, rootMargin: '-10% 0px -60% 0px', threshold: 0 },
    )

    sectionEls.forEach((el) => observer.observe(el))
    return () => observer.disconnect()
  }, [])

  function scrollTo(id: string) {
    const el = document.getElementById(id)
    if (!el) return
    const container = contentRef.current
    if (!container) return
    container.scrollTo({ top: el.offsetTop - 24, behavior: 'smooth' })
    setActiveId(id)
    setPanelOpen(false)
  }

  return (
    <div className="flex h-full">
      {/* TOC panel backdrop (mobile only) */}
      {panelOpen && (
        <div
          className="fixed inset-0 z-20 bg-black/50 md:hidden"
          onClick={() => setPanelOpen(false)}
          aria-hidden="true"
        />
      )}

      <aside
        className={`fixed inset-y-0 left-0 z-30 w-48 shrink-0 border-r border-border bg-surface-base overflow-y-auto py-6 px-3 transition-transform duration-200 ease-out md:static md:z-auto md:translate-x-0 md:sticky md:top-0 md:h-screen ${
          panelOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <p className="text-[10px] font-semibold text-text-muted uppercase tracking-wider px-2 mb-3">
          On this page
        </p>
        <nav className="space-y-0.5">
          {SECTIONS.map((s) => (
            <button
              key={s.id}
              onClick={() => scrollTo(s.id)}
              className={`w-full text-left px-2 py-1.5 rounded-lg text-sm transition-colors ${
                activeId === s.id
                  ? 'text-[color:var(--color-accent)] font-medium bg-[color:var(--color-accent)]/10'
                  : 'text-text-muted hover:text-text-primary hover:bg-surface-elevated'
              }`}
            >
              {s.label}
            </button>
          ))}
        </nav>
      </aside>

      <div ref={contentRef} className="flex-1 overflow-y-auto">
        <div className="max-w-3xl mx-auto px-8 py-8">
          <div className="mb-8 flex items-start justify-between gap-3">
            <div>
              <h1 className="text-3xl font-semibold text-text-primary mb-2">
                Documentation
              </h1>
              <p className="text-sm text-text-muted">
                hube — self-hosted personal hub
              </p>
            </div>
            <IconButton
              icon={<PanelLeft size={18} />}
              aria-label="Toggle table of contents"
              variant="ghost"
              className="md:hidden shrink-0"
              onClick={() => setPanelOpen((v) => !v)}
            />
          </div>

          {docSections.map((section, i) => (
            <Fragment key={section.id}>
              {i > 0 && <Divider />}
              <DocSection data={section} />
            </Fragment>
          ))}
        </div>
      </div>
    </div>
  )
}
