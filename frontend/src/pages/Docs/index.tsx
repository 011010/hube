import { useEffect, useRef, useState } from 'react'
import { PanelLeft } from 'lucide-react'
import { IconButton } from '../../components/atoms/IconButton'


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

function SectionHeading({ id, children }: { id: string; children: React.ReactNode }) {
  return (
    <h2 id={id} className="text-xl font-semibold text-text-primary mb-4 pt-2 scroll-mt-6">
      {children}
    </h2>
  )
}

function SubHeading({ children }: { children: React.ReactNode }) {
  return (
    <h3 className="text-sm font-semibold text-text-secondary uppercase tracking-wider mt-6 mb-3">
      {children}
    </h3>
  )
}

function P({ children }: { children: React.ReactNode }) {
  return (
    <p className="text-sm text-text-secondary leading-relaxed mb-3">
      {children}
    </p>
  )
}

function Code({ children }: { children: React.ReactNode }) {
  return (
    <code className="bg-surface-elevated text-text-primary px-1.5 py-0.5 rounded text-[12px] font-mono">
      {children}
    </code>
  )
}

function CodeBlock({ children }: { children: string }) {
  return (
    <pre className="bg-surface-base border border-border rounded-xl p-4 text-[12px] font-mono text-text-secondary overflow-x-auto mb-4 leading-relaxed">
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
          <tr className="border-b border-border">
            {headers.map((h) => (
              <th
                key={h}
                className="text-left py-2 pr-6 text-xs font-semibold text-text-muted uppercase tracking-wider whitespace-nowrap"
              >
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i} className="border-b border-border/50 last:border-0">
              {row.map((cell, j) => (
                <td key={j} className="py-2 pr-6 text-text-secondary align-top">
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
  return <hr className="border-border my-8" />
}

function BulletList({ items }: { items: (string | React.ReactNode)[] }) {
  return (
    <ul className="space-y-1.5 mb-4">
      {items.map((item, i) => (
        <li key={i} className="flex gap-2 text-sm text-text-secondary leading-relaxed">
          <span className="text-text-muted shrink-0 mt-0.5">–</span>
          {item}
        </li>
      ))}
    </ul>
  )
}

// ---------------------------------------------------------------------------
// Sections
// ---------------------------------------------------------------------------

function OverviewSection() {
  return (
    <section>
      <SectionHeading id="overview">Overview</SectionHeading>
      <P>
        hube is a self-hosted personal hub: tasks, notes, calendar, app launcher, wishlist,
        network diagrams, AI playground, and finance widgets — all backed by a single SQLite file.
      </P>
      <SubHeading>Tech stack</SubHeading>
      <Table
        headers={['Layer', 'Stack']}
        rows={[
          ['Backend', 'Go · Chi router · SQLite · sqlx'],
          ['Frontend', 'React · Vite · TanStack Query'],
          ['Database', 'SQLite (FTS5 + WAL mode)'],
          ['AI', 'Claude (Anthropic) · OpenAI-compatible providers'],
        ]}
      />
      <SubHeading>Architecture</SubHeading>
      <P>
        The backend follows{' '}
        <strong className="text-text-primary">Hexagonal / Clean Architecture</strong>:
        domain → application → infrastructure layers. No business logic lives in HTTP handlers.
        Each module owns its repository interface; the SQLite layer provides the implementation.
        All <Code>PUT</Code> endpoints use fetch-then-merge so partial updates never erase existing fields.
      </P>
    </section>
  )
}

function TasksSection() {
  return (
    <section>
      <SectionHeading id="tasks">Tasks</SectionHeading>
      <P>
        Full task management with priorities, statuses, due dates, project linkage, and automatic
        recurring task generation.
      </P>
      <SubHeading>Fields</SubHeading>
      <Table
        headers={['Field', 'Type', 'Values / Notes']}
        rows={[
          ['title', 'string', 'Required'],
          ['description', 'string', 'Optional free text'],
          ['priority', 'enum', <><Code>low</Code> · <Code>medium</Code> · <Code>high</Code></>],
          ['status', 'enum', <><Code>todo</Code> · <Code>in_progress</Code> · <Code>done</Code></>],
          ['due_date', 'ISO 8601', 'Optional — e.g. 2025-12-31T00:00:00Z'],
          ['project_id', 'string', 'Optional — links to a project'],
          ['recurrence', 'enum', <><Code>""</Code> · <Code>daily</Code> · <Code>weekly</Code> · <Code>monthly</Code></>],
        ]}
      />
      <SubHeading>Capabilities</SubHeading>
      <BulletList items={[
        <>Filter by project: <Code>GET /api/v1/tasks?project_id=&#123;id&#125;</Code></>,
        'Recurring tasks auto-recreate via background scheduler every 5 minutes. When a recurring task\'s period elapses, a new copy is created with status todo.',
        'Partial PUT: send only the fields you want to change — existing fields are preserved.',
      ]} />
      <SubHeading>Examples</SubHeading>
      <CodeBlock>{`# Create a weekly recurring task linked to a project
POST /api/v1/tasks
{
  "title": "Weekly review",
  "priority": "high",
  "recurrence": "weekly",
  "project_id": "proj-uuid"
}

# Mark a task in_progress
PUT /api/v1/tasks/{id}
{ "status": "in_progress" }

# List tasks for a specific project
GET /api/v1/tasks?project_id={id}`}</CodeBlock>
    </section>
  )
}

function NotesSection() {
  return (
    <section>
      <SectionHeading id="notes">Notes</SectionHeading>
      <P>
        Markdown notes with tags, folder organisation, full-text search (FTS5), and optional
        semantic search via embeddings.
      </P>
      <SubHeading>Fields</SubHeading>
      <Table
        headers={['Field', 'Type', 'Notes']}
        rows={[
          ['title', 'string', 'Required'],
          ['content', 'string', 'Markdown'],
          ['folder_id', 'string', 'Optional — links to a folder'],
          ['tags', 'string[]', 'Array of tag strings'],
        ]}
      />
      <SubHeading>Capabilities</SubHeading>
      <BulletList items={[
        <>Filter by folder: <Code>GET /api/v1/notes?folder_id=&#123;id&#125;</Code></>,
        <>Full-text search (FTS5): <Code>GET /api/v1/notes/search?q=keyword</Code> — searches title + content.</>,
        <>Semantic search (RAG): <Code>POST /api/v1/notes/semantic-search</Code> — requires <Code>OPENAI_API_KEY</Code> and <Code>OPENAI_EMBEDDING_MODEL</Code>. Returns the most semantically similar notes.</>,
        'When a note is created or updated and RAG is configured, embedding indexing runs in the background — it doesn\'t block the response.',
      ]} />
      <SubHeading>Examples</SubHeading>
      <CodeBlock>{`# Create a note with tags and folder
POST /api/v1/notes
{
  "title": "Meeting notes",
  "content": "## Agenda\\n- Item 1\\n- Item 2",
  "folder_id": "folder-uuid",
  "tags": ["meetings", "q4"]
}

# Full-text search
GET /api/v1/notes/search?q=architecture

# Semantic search (requires embeddings configured)
POST /api/v1/notes/semantic-search
{ "q": "ideas about product design", "top_k": 5 }`}</CodeBlock>
    </section>
  )
}

function FoldersSection() {
  return (
    <section>
      <SectionHeading id="folders">Folders</SectionHeading>
      <P>
        Folders organise notes into a hierarchy. Each folder can have a parent, enabling nested
        structures.
      </P>
      <SubHeading>Fields</SubHeading>
      <Table
        headers={['Field', 'Type', 'Notes']}
        rows={[
          ['name', 'string', 'Required'],
          ['parent_id', 'string | null', 'Optional — enables nesting'],
        ]}
      />
      <SubHeading>Endpoints</SubHeading>
      <Table
        headers={['Method', 'Path', 'Description']}
        rows={[
          ['GET', '/api/v1/folders', 'List all folders'],
          ['POST', '/api/v1/folders', 'Create folder'],
          ['PUT', '/api/v1/folders/{id}', 'Rename or reparent'],
          ['DELETE', '/api/v1/folders/{id}', 'Delete folder'],
        ]}
      />
      <SubHeading>Examples</SubHeading>
      <CodeBlock>{`# Create a root folder
POST /api/v1/folders
{ "name": "Work" }

# Create a nested folder
POST /api/v1/folders
{ "name": "Q4 Planning", "parent_id": "work-folder-uuid" }

# Then filter notes by folder
GET /api/v1/notes?folder_id={id}`}</CodeBlock>
    </section>
  )
}

function CalendarSection() {
  return (
    <section>
      <SectionHeading id="calendar">Calendar</SectionHeading>
      <P>
        Calendar events with full-day support, color coding, and date-range filtering.
      </P>
      <SubHeading>Fields</SubHeading>
      <Table
        headers={['Field', 'Type', 'Notes']}
        rows={[
          ['title', 'string', 'Required'],
          ['description', 'string', 'Optional'],
          ['start_at', 'RFC3339', 'e.g. 2025-12-01T09:00:00Z'],
          ['end_at', 'RFC3339', 'Must be after start_at'],
          ['all_day', 'boolean', 'Hides time display when true'],
          ['color', 'string', 'Hex color — e.g. #6366f1'],
        ]}
      />
      <SubHeading>Capabilities</SubHeading>
      <BulletList items={[
        <>Date-range filter: <Code>GET /api/v1/events?from=2025-01-01T00:00:00Z&to=2025-01-31T23:59:59Z</Code> — both params required for filtering, otherwise returns all events.</>,
        'All-day events set all_day: true — the calendar renders them without a time slot.',
      ]} />
      <SubHeading>Examples</SubHeading>
      <CodeBlock>{`# Create an all-day event
POST /api/v1/events
{
  "title": "Team offsite",
  "start_at": "2025-11-10T00:00:00Z",
  "end_at": "2025-11-11T00:00:00Z",
  "all_day": true,
  "color": "#f59e0b"
}

# Get events for a week
GET /api/v1/events?from=2025-11-10T00:00:00Z&to=2025-11-17T00:00:00Z`}</CodeBlock>
    </section>
  )
}

function ProjectsSection() {
  return (
    <section>
      <SectionHeading id="projects">Projects</SectionHeading>
      <P>
        Projects group tasks together and track their lifecycle through three statuses.
      </P>
      <SubHeading>Fields</SubHeading>
      <Table
        headers={['Field', 'Type', 'Values']}
        rows={[
          ['name', 'string', 'Required'],
          ['description', 'string', 'Optional'],
          ['status', 'enum', <><Code>active</Code> · <Code>paused</Code> · <Code>done</Code></>],
        ]}
      />
      <SubHeading>Capabilities</SubHeading>
      <BulletList items={[
        'Tasks link to a project via project_id — one task belongs to at most one project.',
        <>Filter tasks by project: <Code>GET /api/v1/tasks?project_id=&#123;id&#125;</Code></>,
        'The AI agent\'s list_projects tool returns project details including task counts and progress.',
      ]} />
      <SubHeading>Examples</SubHeading>
      <CodeBlock>{`# Create a project
POST /api/v1/projects
{ "name": "Product launch", "status": "active" }

# Mark done
PUT /api/v1/projects/{id}
{ "status": "done" }

# Get all tasks for a project
GET /api/v1/tasks?project_id={id}`}</CodeBlock>
    </section>
  )
}

function AppsSection() {
  return (
    <section>
      <SectionHeading id="apps">Apps</SectionHeading>
      <P>
        The app launcher holds shortcuts to any URL. Active apps are shown as cards; inactive
        ones are hidden from the launcher without being deleted.
      </P>
      <SubHeading>Fields</SubHeading>
      <Table
        headers={['Field', 'Type', 'Notes']}
        rows={[
          ['name', 'string', 'Required — display label'],
          ['url', 'string', 'Required — target URL'],
          ['description', 'string', 'Optional subtitle'],
          ['icon', 'string', 'Emoji or URL for the card icon'],
          ['color', 'string', 'Hex accent color for the card'],
          ['sort_order', 'number', 'Lower = first in the grid'],
          ['active', 'boolean', 'false hides without deleting'],
        ]}
      />
      <SubHeading>Examples</SubHeading>
      <CodeBlock>{`# Add a new shortcut
POST /api/v1/apps
{
  "name": "Grafana",
  "url": "https://grafana.internal",
  "icon": "📊",
  "color": "#f97316",
  "sort_order": 1,
  "active": true
}

# Hide without deleting
PUT /api/v1/apps/{id}
{ "active": false }

# Reorder
PUT /api/v1/apps/{id}
{ "sort_order": 5 }`}</CodeBlock>
    </section>
  )
}

function WishlistSection() {
  return (
    <section>
      <SectionHeading id="wishlist">Wishlist</SectionHeading>
      <P>
        Track items you want to buy with target vs current price tracking, store references,
        and purchase status.
      </P>
      <SubHeading>Fields</SubHeading>
      <Table
        headers={['Field', 'Type', 'Notes']}
        rows={[
          ['name', 'string', 'Required'],
          ['description', 'string', 'Optional details'],
          ['url', 'string', 'Product link (http/https only)'],
          ['store', 'string', 'Store name — e.g. Amazon'],
          ['target_price', 'number', 'Price you\'re willing to pay'],
          ['current_price', 'number', 'Current listed price'],
          ['currency', 'string', 'e.g. USD, EUR, MXN'],
          ['priority', 'enum', <><Code>low</Code> · <Code>medium</Code> · <Code>high</Code></>],
          ['status', 'enum', <><Code>pending</Code> · <Code>purchased</Code></>],
          ['notes', 'string', 'Free-text notes'],
        ]}
      />
      <SubHeading>Examples</SubHeading>
      <CodeBlock>{`# Add an item with price tracking
POST /api/v1/wishlist
{
  "name": "Keychron Q1",
  "url": "https://keychron.com/products/q1",
  "store": "Keychron",
  "target_price": 150,
  "current_price": 169,
  "currency": "USD",
  "priority": "medium"
}

# Mark as purchased
PUT /api/v1/wishlist/{id}
{ "status": "purchased" }

# Update current price after checking
PUT /api/v1/wishlist/{id}
{ "current_price": 145 }`}</CodeBlock>
    </section>
  )
}

function DiagramsSection() {
  return (
    <section>
      <SectionHeading id="diagrams">Diagrams</SectionHeading>
      <P>
        Visual network diagrams built with React Flow. Nodes and edges are stored as JSON blobs —
        the UI handles the canvas; the API handles persistence.
      </P>
      <SubHeading>Fields</SubHeading>
      <Table
        headers={['Field', 'Type', 'Notes']}
        rows={[
          ['name', 'string', 'Required — diagram label'],
          ['nodes', 'JSON', 'React Flow node array'],
          ['edges', 'JSON', 'React Flow edge array'],
        ]}
      />
      <SubHeading>Examples</SubHeading>
      <CodeBlock>{`# Create an empty diagram
POST /api/v1/diagrams
{ "name": "Infrastructure", "nodes": [], "edges": [] }

# Save diagram state (full replace)
PUT /api/v1/diagrams/{id}
{
  "name": "Infrastructure",
  "nodes": [{ "id": "1", "data": { "label": "API" }, "position": { "x": 100, "y": 100 } }],
  "edges": []
}`}</CodeBlock>
    </section>
  )
}

function FinanceSection() {
  return (
    <section>
      <SectionHeading id="finance">Finance</SectionHeading>
      <P>
        Read-only financial summary pulled from a{' '}
        <strong className="text-text-primary">Money Monkey</strong> instance.
        Configure with <Code>MONKEYAPI_URL</Code> and <Code>MONKEYAPI_KEY</Code>.
        Returns <Code>&#123;"configured": false&#125;</Code> if not set up.
      </P>
      <SubHeading>Response shape</SubHeading>
      <Table
        headers={['Field', 'Type', 'Description']}
        rows={[
          ['balance', 'number', 'Current account balance'],
          ['month_income', 'number', 'Total income this month'],
          ['month_expenses', 'number', 'Total expenses this month'],
          ['recent_transactions', 'array', 'Last N transactions'],
        ]}
      />
      <SubHeading>Transaction fields</SubHeading>
      <Table
        headers={['Field', 'Type', 'Notes']}
        rows={[
          ['id', 'string', ''],
          ['amount', 'number', 'Positive = income, negative = expense'],
          ['type', 'string', 'Transaction type'],
          ['category', 'string', 'e.g. Food, Transport'],
          ['description', 'string', ''],
          ['date', 'string', 'ISO 8601 date'],
        ]}
      />
      <SubHeading>Example</SubHeading>
      <CodeBlock>{`GET /api/v1/finance/summary

# Response
{
  "balance": 4250.00,
  "month_income": 5000.00,
  "month_expenses": 1820.50,
  "recent_transactions": [
    { "amount": -45.00, "category": "Food", "description": "Grocery run", "date": "2025-11-08" }
  ]
}`}</CodeBlock>
    </section>
  )
}

function CardsSection() {
  return (
    <section>
      <SectionHeading id="cards">Card Tracker</SectionHeading>
      <P>
        Read-only credit card summary pulled from a{' '}
        <strong className="text-text-primary">PayPinga</strong> instance.
        Configure with <Code>PAYPINGA_URL</Code> and <Code>PAYPINGA_KEY</Code>.
        Returns <Code>&#123;"configured": false&#125;</Code> if not set up.
      </P>
      <SubHeading>Response shape</SubHeading>
      <Table
        headers={['Field', 'Type', 'Description']}
        rows={[
          ['total_debt', 'number', 'Sum of all card balances'],
          ['total_payment', 'number', 'Total estimated payment due'],
          ['next_pay_date', 'string | null', 'Earliest upcoming payment date'],
          ['cards', 'array', 'Per-card breakdown'],
        ]}
      />
      <SubHeading>Card fields</SubHeading>
      <Table
        headers={['Field', 'Type', 'Notes']}
        rows={[
          ['alias', 'string', 'Friendly name — e.g. "Chase Sapphire"'],
          ['bank', 'string', 'Issuing bank'],
          ['last4', 'string', 'Last 4 digits'],
          ['pay_date', 'string', 'Payment due date'],
          ['limit', 'number', 'Credit limit'],
          ['estimated_payment', 'number', 'Minimum / estimated payment'],
          ['balance', 'number', 'Current balance'],
        ]}
      />
      <SubHeading>Example</SubHeading>
      <CodeBlock>{`GET /api/v1/cards/summary

# Response
{
  "total_debt": 1240.00,
  "total_payment": 310.00,
  "next_pay_date": "2025-11-15",
  "cards": [
    {
      "alias": "Chase Sapphire",
      "bank": "Chase",
      "last4": "4242",
      "pay_date": "2025-11-15",
      "limit": 5000,
      "estimated_payment": 310.00,
      "balance": 1240.00
    }
  ]
}`}</CodeBlock>
    </section>
  )
}

function AISection() {
  return (
    <section>
      <SectionHeading id="ai">AI Agent</SectionHeading>
      <P>
        The AI playground exposes a chat interface backed by Claude or any OpenAI-compatible
        provider. The agent has access to hub tools — it can read and write your data using natural
        language.
      </P>
      <SubHeading>Configuration</SubHeading>
      <Table
        headers={['Provider', 'Required vars']}
        rows={[
          ['Claude (Anthropic)', <><Code>ANTHROPIC_API_KEY</Code></>],
          ['OpenAI / OpenRouter / Ollama', <><Code>OPENAI_API_KEY</Code> · <Code>OPENAI_BASE_URL</Code> · <Code>OPENAI_MODEL</Code></>],
        ]}
      />
      <SubHeading>Available tools</SubHeading>
      <Table
        headers={['Tool', 'What it does', 'Key params']}
        rows={[
          ['list_tasks', 'List tasks, optionally filtered by status', 'status: todo | in_progress | done'],
          ['create_task', 'Create a new task', 'title (req), description, priority, due_date, project_id'],
          ['update_task', 'Update status, title, or priority of a task', 'id (req), title, status, priority'],
          ['list_notes', 'List notes, optionally filtered by folder', 'folder_id'],
          ['create_note', 'Create a new Markdown note', 'title (req), content'],
          ['search_notes', 'Full-text search across all notes', 'query (req)'],
          ['list_projects', 'List all projects with task counts', '—'],
          ['list_events', 'List calendar events in a date range', 'from, to (YYYY-MM-DD, defaults: today + 30d)'],
          ['list_apps', 'List all launcher apps', '—'],
        ]}
      />
      <SubHeading>Chat endpoint</SubHeading>
      <CodeBlock>{`POST /api/v1/ai/chat
{
  "messages": [
    { "role": "user", "content": "What tasks are in progress?" }
  ]
}

# Multi-turn conversation — include prior messages
POST /api/v1/ai/chat
{
  "messages": [
    { "role": "user",      "content": "Show my tasks" },
    { "role": "assistant", "content": "You have 3 tasks: ..." },
    { "role": "user",      "content": "Mark the first one done" }
  ]
}`}</CodeBlock>
      <SubHeading>RAG — semantic search on notes</SubHeading>
      <P>
        When <Code>OPENAI_API_KEY</Code> and <Code>OPENAI_EMBEDDING_MODEL</Code> are set, notes
        are automatically indexed with embeddings on create/update. You can then run cosine
        similarity search independently of the AI chat:
      </P>
      <CodeBlock>{`POST /api/v1/notes/semantic-search
{ "q": "ideas about system design", "top_k": 5 }
# top_k defaults to 5, max 50`}</CodeBlock>
    </section>
  )
}

function APISection() {
  const crudRows = (resource: string): (string | React.ReactNode)[][] => [
    ['GET',    `${resource}`,      'List all items'],
    ['POST',   `${resource}`,      'Create item — returns 201'],
    ['GET',    `${resource}/{id}`, 'Get by ID'],
    ['PUT',    `${resource}/{id}`, 'Partial update — unset fields preserved'],
    ['DELETE', `${resource}/{id}`, 'Delete — returns 204'],
  ]

  const specialRows: (string | React.ReactNode)[][] = [
    ['GET',  '/api/v1/notes/search?q=',              'FTS5 full-text search'],
    ['POST', '/api/v1/notes/semantic-search',         'RAG search — body: { q, top_k }'],
    ['GET',  '/api/v1/events?from=&to=',              'Filter events by RFC3339 range'],
    ['GET',  '/api/v1/tasks?project_id=',             'Filter tasks by project'],
    ['GET',  '/api/v1/notes?folder_id=',              'Filter notes by folder'],
    ['GET',  '/api/v1/settings',                      'Get all settings (API keys masked)'],
    ['PUT',  '/api/v1/settings',                      'Update settings'],
    ['POST', '/api/v1/email/digest',                  'Send task digest email via SMTP'],
    ['GET',  '/api/v1/finance/summary',               'Money Monkey balance + transactions'],
    ['GET',  '/api/v1/cards/summary',                 'PayPinga card tracker summary'],
    ['POST', '/api/v1/ai/chat',                       'AI chat — body: { messages[] }'],
    ['POST', '/api/v1/backup',                        'Trigger immediate DB backup'],
    ['GET',  '/api/v1/export',                        'Download ZIP export'],
    ['GET',  '/health',                               'Health check — returns 200'],
  ]

  const modules: [string, string][] = [
    ['Tasks',    '/api/v1/tasks'],
    ['Notes',    '/api/v1/notes'],
    ['Folders',  '/api/v1/folders'],
    ['Events',   '/api/v1/events'],
    ['Apps',     '/api/v1/apps'],
    ['Wishlist', '/api/v1/wishlist'],
    ['Projects', '/api/v1/projects'],
    ['Diagrams', '/api/v1/diagrams'],
  ]

  return (
    <section>
      <SectionHeading id="api">API Reference</SectionHeading>
      <P>
        Base path: <Code>/api/v1</Code>. All endpoints accept and return JSON.
        Max request body: <Code>2 MB</Code>.
      </P>

      {modules.map(([name, path]) => (
        <div key={path}>
          <SubHeading>{name} — <Code>{path}</Code></SubHeading>
          <Table headers={['Method', 'Path', 'Description']} rows={crudRows(path)} />
        </div>
      ))}

      <SubHeading>Special endpoints</SubHeading>
      <Table headers={['Method', 'Path', 'Description']} rows={specialRows} />
    </section>
  )
}

function ConfigSection() {
  const rows: (string | React.ReactNode)[][] = [
    ['PORT',                  '8080',      'HTTP server port'],
    ['DB_PATH',               './hube.db', 'SQLite database path'],
    ['BACKUP_DIR',            './backups', 'Backup output directory'],
    ['BACKUP_RETAIN',         '7',         'Number of backup files to keep'],
    ['ANTHROPIC_API_KEY',     '—',         'Claude API key for AI playground'],
    ['OPENAI_API_KEY',        '—',         'OpenAI-compatible API key'],
    ['OPENAI_BASE_URL',       '—',         <>Custom base URL (e.g. <Code>https://api.openrouter.ai/v1</Code>)</>],
    ['OPENAI_MODEL',          '—',         'Model name for chat completions'],
    ['OPENAI_EMBEDDING_MODEL','—',         'Model for RAG embeddings'],
    ['SMTP_HOST',             '—',         'SMTP server hostname'],
    ['SMTP_PORT',             '587',       'SMTP port (587 = STARTTLS, 465 = TLS)'],
    ['SMTP_USER',             '—',         'SMTP username'],
    ['SMTP_PASS',             '—',         'SMTP password'],
    ['SMTP_FROM',             '—',         'From address for digest emails'],
    ['MONKEYAPI_URL',         '—',         'Money Monkey base URL'],
    ['MONKEYAPI_KEY',         '—',         'Money Monkey API key'],
    ['PAYPINGA_URL',          '—',         'PayPinga base URL'],
    ['PAYPINGA_KEY',          '—',         'PayPinga API key'],
    ['HUBE_DOMAIN',           '—',         <>Public domain for CORS (e.g. <Code>hub.example.com</Code>)</>],
  ]
  return (
    <section>
      <SectionHeading id="config">Configuration</SectionHeading>
      <P>
        All configuration is done through environment variables. The server reads them at
        startup — a restart is required when changing credentials. hube auto-loads a{' '}
        <Code>.env</Code> file at the project root if present.
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

# Frontend (separate terminal)
cd frontend
pnpm install
pnpm dev`}</CodeBlock>

      <SubHeading>Docker (development)</SubHeading>
      <CodeBlock>{`docker compose up`}</CodeBlock>

      <SubHeading>Docker (production)</SubHeading>
      <CodeBlock>{`docker compose -f docker-compose.prod.yml up -d`}</CodeBlock>

      <SubHeading>Reverse proxy with Caddy</SubHeading>
      <P>
        The repo includes a <Code>Caddyfile</Code> for TLS termination and proxying.
        Set <Code>HUBE_DOMAIN</Code> to your public domain so CORS is configured correctly.
      </P>
      <CodeBlock>{`# Caddyfile snippet
hub.example.com {
  reverse_proxy localhost:8080
}`}</CodeBlock>
    </section>
  )
}

function BackupSection() {
  return (
    <section>
      <SectionHeading id="backup">Backup & Export</SectionHeading>
      <SubHeading>Automatic backup</SubHeading>
      <P>
        SQLite is a single file — zero setup. Automatic backup runs every{' '}
        <strong className="text-text-primary">12 hours</strong> via{' '}
        <Code>VACUUM INTO</Code>, which is safe under concurrent writes. The{' '}
        <Code>BACKUP_RETAIN</Code> env var controls how many files to keep (default 7).
      </P>
      <SubHeading>Manual backup</SubHeading>
      <CodeBlock>{`POST /api/v1/backup
# Triggers an immediate VACUUM INTO backup — returns 200 on success`}</CodeBlock>
      <SubHeading>Export all data</SubHeading>
      <CodeBlock>{`GET /api/v1/export
# Downloads a ZIP archive`}</CodeBlock>
      <Table
        headers={['Content', 'Format']}
        rows={[
          ['Notes',          'Markdown (.md)'],
          ['Events',         'iCalendar (.ics)'],
          ['Everything else','JSON'],
        ]}
      />
    </section>
  )
}

function SecuritySection() {
  const items: (string | React.ReactNode)[] = [
    <>All <Code>PUT</Code> endpoints use fetch-then-merge — partial updates never erase existing fields.</>,
    <>Request body capped at <Code>2 MB</Code>.</>,
    <>HTTP timeouts: read 15 s, write 60 s, idle 120 s.</>,
    'Internal errors return a generic message — SQL details never leak to the client.',
    <>API keys are masked in the <Code>GET /api/v1/settings</Code> response.</>,
    <>href URLs are validated against an allowlist (<Code>http</Code> / <Code>https</Code> / <Code>mailto</Code> only).</>,
    'No authentication yet — intended for local or VPN use. Protect with Caddy basic_auth or a VPN for external exposure.',
  ]
  return (
    <section>
      <SectionHeading id="security">Security</SectionHeading>
      <BulletList items={items} />
    </section>
  )
}

// ---------------------------------------------------------------------------
// Main page
// ---------------------------------------------------------------------------

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

          <OverviewSection />
          <Divider />
          <TasksSection />
          <Divider />
          <NotesSection />
          <Divider />
          <FoldersSection />
          <Divider />
          <CalendarSection />
          <Divider />
          <ProjectsSection />
          <Divider />
          <AppsSection />
          <Divider />
          <WishlistSection />
          <Divider />
          <DiagramsSection />
          <Divider />
          <FinanceSection />
          <Divider />
          <CardsSection />
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
