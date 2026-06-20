import { useEffect, useRef, useState } from 'react'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface NavSection {
  id: string
  label: string
}

// ---------------------------------------------------------------------------
// Sidebar nav sections
// ---------------------------------------------------------------------------

const SECTIONS: NavSection[] = [
  { id: 'overview', label: 'Overview' },
  { id: 'modules',  label: 'Modules' },
  { id: 'ai',       label: 'AI' },
  { id: 'api',      label: 'API Reference' },
  { id: 'config',   label: 'Configuration' },
  { id: 'running',  label: 'Running' },
  { id: 'backup',   label: 'Backup & Export' },
  { id: 'security', label: 'Security' },
]

// ---------------------------------------------------------------------------
// Small reusable components (no external dependencies)
// ---------------------------------------------------------------------------

function SectionHeading({ id, children }: { id: string; children: React.ReactNode }) {
  return (
    <h2
      id={id}
      className="text-xl font-semibold text-gray-900 dark:text-white mb-4 pt-2 scroll-mt-6"
    >
      {children}
    </h2>
  )
}

function SubHeading({ children }: { children: React.ReactNode }) {
  return (
    <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wider mt-6 mb-3">
      {children}
    </h3>
  )
}

function P({ children }: { children: React.ReactNode }) {
  return (
    <p className="text-sm text-gray-600 dark:text-gray-400 leading-relaxed mb-3">
      {children}
    </p>
  )
}

function Code({ children }: { children: React.ReactNode }) {
  return (
    <code className="bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-200 px-1.5 py-0.5 rounded text-[12px] font-mono">
      {children}
    </code>
  )
}

function CodeBlock({ children }: { children: string }) {
  return (
    <pre className="bg-gray-900 dark:bg-gray-950 border border-gray-700 dark:border-gray-800 rounded-xl p-4 text-[12px] font-mono text-gray-200 overflow-x-auto mb-4 leading-relaxed">
      <code>{children}</code>
    </pre>
  )
}

function Table({
  headers,
  rows,
}: {
  headers: string[]
  rows: (string | React.ReactNode)[][]
}) {
  return (
    <div className="overflow-x-auto mb-4">
      <table className="w-full text-sm border-collapse">
        <thead>
          <tr className="border-b border-gray-200 dark:border-gray-700">
            {headers.map((h) => (
              <th
                key={h}
                className="text-left py-2 pr-6 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider whitespace-nowrap"
              >
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr
              key={i}
              className="border-b border-gray-100 dark:border-gray-800 last:border-0"
            >
              {row.map((cell, j) => (
                <td
                  key={j}
                  className="py-2 pr-6 text-gray-700 dark:text-gray-300 align-top"
                >
                  {cell}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function Divider() {
  return <hr className="border-gray-200 dark:border-gray-800 my-8" />
}

// ---------------------------------------------------------------------------
// Section content
// ---------------------------------------------------------------------------

function OverviewSection() {
  return (
    <section>
      <SectionHeading id="overview">Overview</SectionHeading>
      <P>
        hube is a self-hosted personal hub: tasks, notes, calendar, apps launcher, wishlist,
        network diagrams, AI playground, and finance widgets — all in one place.
      </P>
      <SubHeading>Tech stack</SubHeading>
      <Table
        headers={['Layer', 'Stack']}
        rows={[
          ['Backend', 'Go · Chi router · SQLite · sqlx'],
          ['Frontend', 'React · Vite · TanStack Query'],
          ['Database', 'SQLite (FTS5 + WAL mode)'],
        ]}
      />
      <SubHeading>Architecture</SubHeading>
      <P>
        The backend follows{' '}
        <strong className="text-gray-900 dark:text-white">
          Hexagonal / Clean Architecture
        </strong>
        : domain → application → infrastructure layers. No business logic lives in HTTP
        handlers. Each module owns its repository interface; the infrastructure layer
        provides the SQLite implementation.
      </P>
    </section>
  )
}

function ModulesSection() {
  const rows: (string | React.ReactNode)[][] = [
    [
      <strong className="text-gray-900 dark:text-white">Tasks</strong>,
      'title, description, priority, status, due_date, recurrence, project_id',
      'Recurring tasks auto-recreate via scheduler every 5 min. Priority: low / medium / high. Status: todo / done.',
    ],
    [
      <strong className="text-gray-900 dark:text-white">Notes</strong>,
      'title, content (Markdown), tags[], folder_id',
      'FTS5 full-text search + RAG semantic search via embeddings.',
    ],
    [
      <strong className="text-gray-900 dark:text-white">Calendar</strong>,
      'title, description, start_at, end_at, all_day, color',
      'Filter by date range with ?from=&to=',
    ],
    [
      <strong className="text-gray-900 dark:text-white">Projects</strong>,
      'name, description, status',
      'Status: active / paused / done. Tasks link via project_id.',
    ],
    [
      <strong className="text-gray-900 dark:text-white">Apps</strong>,
      'name, url, icon, description, color, sort_order, active',
      'Launcher shows active apps as cards. Inactive apps are hidden.',
    ],
    [
      <strong className="text-gray-900 dark:text-white">Wishlist</strong>,
      'name, url, price, currency, priority, purchased',
      'Price tracking with purchase status toggle.',
    ],
    [
      <strong className="text-gray-900 dark:text-white">Diagrams</strong>,
      'name, nodes (JSON), edges (JSON)',
      'Visual network diagram editor via React Flow.',
    ],
    [
      <strong className="text-gray-900 dark:text-white">Settings</strong>,
      'key-value store',
      'Integrations: Money Monkey, PayPinga, Claude, OpenAI, SMTP.',
    ],
  ]
  return (
    <section>
      <SectionHeading id="modules">Modules</SectionHeading>
      <Table headers={['Module', 'Key fields', 'Notes']} rows={rows} />
    </section>
  )
}

function AISection() {
  return (
    <section>
      <SectionHeading id="ai">AI</SectionHeading>
      <SubHeading>Claude agent</SubHeading>
      <P>
        A Claude agent has full tool access to all hub modules: list tasks, create notes,
        query events, search apps, and more. Configure it by setting{' '}
        <Code>ANTHROPIC_API_KEY</Code>.
      </P>
      <SubHeading>OpenAI-compatible providers</SubHeading>
      <P>
        Set <Code>OPENAI_BASE_URL</Code> to point at any OpenAI-compatible provider —
        OpenRouter, Ollama, local OpenCode instances, etc. — without changing application code.
      </P>
      <SubHeading>RAG — semantic search</SubHeading>
      <P>
        Notes are indexed with embeddings. Cosine similarity search is available via{' '}
        <Code>POST /api/v1/notes/semantic-search</Code>. Set{' '}
        <Code>OPENAI_EMBEDDING_MODEL</Code> to choose the embedding model.
      </P>
    </section>
  )
}

function APISection() {
  const crudRows = (resource: string): (string | React.ReactNode)[][] => [
    ['GET', `${resource}`, 'List all items'],
    ['POST', `${resource}`, 'Create item'],
    ['GET', `${resource}/{id}`, 'Get item by ID'],
    ['PUT', `${resource}/{id}`, 'Update item (partial — unset fields preserved)'],
    ['DELETE', `${resource}/{id}`, 'Delete item'],
  ]

  const specialRows: (string | React.ReactNode)[][] = [
    ['GET', '/api/v1/notes/search?q=', 'FTS5 full-text search'],
    ['POST', '/api/v1/notes/semantic-search', 'RAG semantic search — body: { q, top_k }'],
    ['GET', '/api/v1/settings', 'Get all settings (API keys masked)'],
    ['PUT', '/api/v1/settings', 'Update settings'],
    ['POST', '/api/v1/email/digest', 'Send task digest email'],
    ['GET', '/api/v1/finance/summary', 'Money Monkey balance + transactions'],
    ['GET', '/api/v1/cards/summary', 'PayPinga card tracker'],
    ['POST', '/api/v1/ai/chat', 'AI chat — body: { messages[], model? }'],
    ['POST', '/api/v1/backup', 'Trigger immediate DB backup'],
    ['GET', '/api/v1/export', 'Download ZIP export (notes, events, JSON)'],
    ['GET', '/health', 'Health check'],
  ]

  const modules: [string, string][] = [
    ['Tasks', '/api/v1/tasks'],
    ['Notes', '/api/v1/notes'],
    ['Events', '/api/v1/events'],
    ['Apps', '/api/v1/apps'],
    ['Wishlist', '/api/v1/wishlist'],
    ['Projects', '/api/v1/projects'],
    ['Diagrams', '/api/v1/diagrams'],
  ]

  return (
    <section>
      <SectionHeading id="api">API Reference</SectionHeading>
      <P>
        Base path: <Code>/api/v1</Code>. All endpoints accept and return JSON.
      </P>

      {modules.map(([name, path]) => (
        <div key={path}>
          <SubHeading>{name} — <Code>{path}</Code></SubHeading>
          <Table
            headers={['Method', 'Path', 'Description']}
            rows={crudRows(path)}
          />
        </div>
      ))}

      <SubHeading>Special endpoints</SubHeading>
      <Table headers={['Method', 'Path', 'Description']} rows={specialRows} />
    </section>
  )
}

function ConfigSection() {
  const rows: (string | React.ReactNode)[][] = [
    ['PORT', '8080', 'HTTP server port'],
    ['DB_PATH', './hube.db', 'SQLite database path'],
    ['BACKUP_DIR', './backups', 'Backup output directory'],
    ['BACKUP_RETAIN', '7', 'Number of backup files to keep'],
    ['ANTHROPIC_API_KEY', '—', 'Claude API key for AI Playground'],
    ['OPENAI_API_KEY', '—', 'OpenAI-compatible API key'],
    ['OPENAI_BASE_URL', '—', <>Custom base URL (e.g. <Code>https://api.openrouter.ai/v1</Code>)</>],
    ['OPENAI_MODEL', '—', 'Model name for OpenAI-compatible provider'],
    ['OPENAI_EMBEDDING_MODEL', '—', 'Model for RAG embeddings'],
    ['SMTP_HOST', '—', 'SMTP server hostname'],
    ['SMTP_PORT', '587', 'SMTP port (587 = STARTTLS, 465 = TLS)'],
    ['SMTP_USER', '—', 'SMTP username'],
    ['SMTP_PASS', '—', 'SMTP password'],
    ['SMTP_FROM', '—', 'From address for digest emails'],
    ['MONKEYAPI_URL', '—', 'Money Monkey base URL'],
    ['MONKEYAPI_KEY', '—', 'Money Monkey API key'],
    ['PAYPINGA_URL', '—', 'PayPinga base URL'],
    ['PAYPINGA_KEY', '—', 'PayPinga API key'],
    ['HUBE_DOMAIN', '—', <>Public domain for CORS (e.g. <Code>hub.example.com</Code>)</>],
  ]
  return (
    <section>
      <SectionHeading id="config">Configuration</SectionHeading>
      <P>
        All configuration is done through environment variables. The server reads them at
        startup — a restart is required when changing credentials.
      </P>
      <Table headers={['Variable', 'Default', 'Description']} rows={rows} />
    </section>
  )
}

function RunningSection() {
  return (
    <section>
      <SectionHeading id="running">Running</SectionHeading>
      <SubHeading>Local development</SubHeading>
      <CodeBlock>{`# Backend
cd backend
go run ./cmd/api

# Frontend (in a separate terminal)
cd frontend
pnpm install
pnpm dev`}</CodeBlock>

      <SubHeading>Docker</SubHeading>
      <CodeBlock>{`docker compose up`}</CodeBlock>
    </section>
  )
}

function BackupSection() {
  return (
    <section>
      <SectionHeading id="backup">Backup & Export</SectionHeading>
      <SubHeading>Automatic backup</SubHeading>
      <P>
        SQLite is a single file — zero setup. Automatic backup runs every 12 hours via{' '}
        <Code>VACUUM INTO</Code>, which is safe under concurrent writes. The number of
        retained files is controlled by <Code>BACKUP_RETAIN</Code>.
      </P>
      <SubHeading>Manual backup</SubHeading>
      <P>
        Trigger an immediate backup at any time:
      </P>
      <CodeBlock>{`POST /api/v1/backup`}</CodeBlock>
      <SubHeading>Export all data</SubHeading>
      <P>
        Download a ZIP archive containing all your data:{' '}
        <Code>GET /api/v1/export</Code>
      </P>
      <Table
        headers={['Content', 'Format']}
        rows={[
          ['Notes', 'Markdown (.md)'],
          ['Events', 'iCalendar (.ics)'],
          ['Everything else', 'JSON'],
        ]}
      />
    </section>
  )
}

function SecuritySection() {
  const items = [
    'All PUT endpoints use fetch-then-merge — partial updates never erase existing fields.',
    'Request body capped at 2 MB.',
    'HTTP timeouts: read 15 s, write 60 s, idle 120 s.',
    'Internal errors return a generic message — SQL details never leak to the client.',
    'API keys are masked in the settings GET response.',
    'href URLs are validated against an allowlist (http / https / mailto only).',
    'No authentication yet — intended for local or VPN use.',
  ]
  return (
    <section>
      <SectionHeading id="security">Security</SectionHeading>
      <ul className="space-y-2 mb-4">
        {items.map((item) => (
          <li key={item} className="flex gap-2 text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
            <span className="text-gray-400 dark:text-gray-600 shrink-0 mt-0.5">–</span>
            {item}
          </li>
        ))}
      </ul>
    </section>
  )
}

// ---------------------------------------------------------------------------
// Main page
// ---------------------------------------------------------------------------

export function DocsPage() {
  const [activeId, setActiveId] = useState<string>(SECTIONS[0].id)
  const contentRef = useRef<HTMLDivElement>(null)

  // Scrollspy: observe which section is in the viewport
  useEffect(() => {
    const sectionEls = SECTIONS.map((s) => document.getElementById(s.id)).filter(Boolean) as HTMLElement[]

    if (sectionEls.length === 0) return

    const observer = new IntersectionObserver(
      (entries) => {
        // Find the topmost intersecting section
        const visible = entries
          .filter((e) => e.isIntersecting)
          .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)

        if (visible.length > 0) {
          setActiveId(visible[0].target.id)
        }
      },
      {
        root: contentRef.current,
        rootMargin: '-10% 0px -60% 0px',
        threshold: 0,
      },
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
  }

  return (
    <div className="flex h-full">
      {/* Docs sidebar */}
      <aside className="w-48 shrink-0 border-r border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 sticky top-0 h-screen overflow-y-auto py-6 px-3">
        <p className="text-[10px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider px-2 mb-3">
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
                  : 'text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-800'
              }`}
            >
              {s.label}
            </button>
          ))}
        </nav>
      </aside>

      {/* Scrollable content */}
      <div ref={contentRef} className="flex-1 overflow-y-auto">
        <div className="max-w-3xl mx-auto px-8 py-8">
          <div className="mb-8">
            <h1 className="text-3xl font-semibold text-gray-900 dark:text-white mb-2">
              Documentation
            </h1>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              hube — self-hosted personal hub
            </p>
          </div>

          <OverviewSection />
          <Divider />
          <ModulesSection />
          <Divider />
          <AISection />
          <Divider />
          <APISection />
          <Divider />
          <ConfigSection />
          <Divider />
          <RunningSection />
          <Divider />
          <BackupSection />
          <Divider />
          <SecuritySection />
        </div>
      </div>
    </div>
  )
}
