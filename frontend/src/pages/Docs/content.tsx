import { Code, type DocSectionData } from './DocSection'

// ---------------------------------------------------------------------------
// Static documentation content.
//
// This is pure data — no rendering logic lives here. Each entry matches
// `DocSectionData` from ./DocSection and is rendered by the single generic
// `DocSection` component. Keeping content and rendering separate means this
// file can grow arbitrarily (it holds all docs copy) without the rendering
// code or the page shell growing with it.
// ---------------------------------------------------------------------------

const apiCrudRows = (resource: string): (string | React.ReactNode)[][] => [
  ['GET', `${resource}`, 'List all items'],
  ['POST', `${resource}`, 'Create item — returns 201'],
  ['GET', `${resource}/{id}`, 'Get by ID'],
  ['PUT', `${resource}/{id}`, 'Partial update — unset fields preserved'],
  ['DELETE', `${resource}/{id}`, 'Delete — returns 204'],
]

const apiModules: [string, string][] = [
  ['Tasks', '/api/v1/tasks'],
  ['Notes', '/api/v1/notes'],
  ['Folders', '/api/v1/folders'],
  ['Events', '/api/v1/events'],
  ['Apps', '/api/v1/apps'],
  ['Wishlist', '/api/v1/wishlist'],
  ['Projects', '/api/v1/projects'],
  ['Diagrams', '/api/v1/diagrams'],
]

const apiSpecialRows: (string | React.ReactNode)[][] = [
  ['GET', '/api/v1/notes/search?q=', 'FTS5 full-text search'],
  ['POST', '/api/v1/notes/semantic-search', 'RAG search — body: { q, top_k }'],
  ['GET', '/api/v1/events?from=&to=', 'Filter events by RFC3339 range'],
  ['GET', '/api/v1/tasks?project_id=', 'Filter tasks by project'],
  ['GET', '/api/v1/notes?folder_id=', 'Filter notes by folder'],
  ['GET', '/api/v1/settings', 'Get all settings (API keys masked)'],
  ['PUT', '/api/v1/settings', 'Update settings'],
  ['POST', '/api/v1/email/digest', 'Send task digest email via SMTP'],
  ['GET', '/api/v1/finance/summary', 'Money Monkey balance + transactions'],
  ['GET', '/api/v1/cards/summary', 'PayPinga card tracker summary'],
  ['POST', '/api/v1/ai/chat', 'AI chat — body: { messages[] }'],
  ['POST', '/api/v1/backup', 'Trigger immediate DB backup'],
  ['GET', '/api/v1/export', 'Download ZIP export'],
  ['GET', '/health', 'Health check — returns 200'],
]

const configRows: (string | React.ReactNode)[][] = [
  ['PORT', '8080', 'HTTP server port'],
  ['DB_PATH', './hube.db', 'SQLite database path'],
  ['BACKUP_DIR', './backups', 'Backup output directory'],
  ['BACKUP_RETAIN', '7', 'Number of backup files to keep'],
  ['ANTHROPIC_API_KEY', '—', 'Claude API key for AI playground'],
  ['OPENAI_API_KEY', '—', 'OpenAI-compatible API key'],
  ['OPENAI_BASE_URL', '—', <>Custom base URL (e.g. <Code>https://api.openrouter.ai/v1</Code>)</>],
  ['OPENAI_MODEL', '—', 'Model name for chat completions'],
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

const securityItems: (string | React.ReactNode)[] = [
  <>All <Code>PUT</Code> endpoints use fetch-then-merge — partial updates never erase existing fields.</>,
  <>Request body capped at <Code>2 MB</Code>.</>,
  <>HTTP timeouts: read 15 s, write 60 s, idle 120 s.</>,
  'Internal errors return a generic message — SQL details never leak to the client.',
  <>API keys are masked in the <Code>GET /api/v1/settings</Code> response.</>,
  <>href URLs are validated against an allowlist (<Code>http</Code> / <Code>https</Code> / <Code>mailto</Code> only).</>,
  'No authentication yet — intended for local or VPN use. Protect with Caddy basic_auth or a VPN for external exposure.',
]

export const docSections: DocSectionData[] = [
  {
    id: 'overview',
    title: 'Overview',
    subsections: [
      {
        blocks: [
          {
            body: [
              'hube is a self-hosted personal hub: tasks, notes, calendar, app launcher, wishlist, network diagrams, AI playground, and finance widgets — all backed by a single SQLite file.',
            ],
          },
        ],
      },
      {
        heading: 'Tech stack',
        blocks: [
          {
            table: {
              headers: ['Layer', 'Stack'],
              rows: [
                ['Backend', 'Go · Chi router · SQLite · sqlx'],
                ['Frontend', 'React · Vite · TanStack Query'],
                ['Database', 'SQLite (FTS5 + WAL mode)'],
                ['AI', 'Claude (Anthropic) · OpenAI-compatible providers'],
              ],
            },
          },
        ],
      },
      {
        heading: 'Architecture',
        blocks: [
          {
            body: [
              <>
                The backend follows{' '}
                <strong className="text-text-primary">Hexagonal / Clean Architecture</strong>:
                domain → application → infrastructure layers. No business logic lives in HTTP handlers.
                Each module owns its repository interface; the SQLite layer provides the implementation.
                All <Code>PUT</Code> endpoints use fetch-then-merge so partial updates never erase existing fields.
              </>,
            ],
          },
        ],
      },
    ],
  },
  {
    id: 'tasks',
    title: 'Tasks',
    subsections: [
      {
        blocks: [
          {
            body: [
              'Full task management with priorities, statuses, due dates, project linkage, and automatic recurring task generation.',
            ],
          },
        ],
      },
      {
        heading: 'Fields',
        blocks: [
          {
            table: {
              headers: ['Field', 'Type', 'Values / Notes'],
              rows: [
                ['title', 'string', 'Required'],
                ['description', 'string', 'Optional free text'],
                ['priority', 'enum', <><Code>low</Code> · <Code>medium</Code> · <Code>high</Code></>],
                ['status', 'enum', <><Code>todo</Code> · <Code>in_progress</Code> · <Code>done</Code></>],
                ['due_date', 'ISO 8601', 'Optional — e.g. 2025-12-31T00:00:00Z'],
                ['project_id', 'string', 'Optional — links to a project'],
                ['recurrence', 'enum', <><Code>""</Code> · <Code>daily</Code> · <Code>weekly</Code> · <Code>monthly</Code></>],
              ],
            },
          },
        ],
      },
      {
        heading: 'Capabilities',
        blocks: [
          {
            bullets: [
              <>Filter by project: <Code>GET /api/v1/tasks?project_id=&#123;id&#125;</Code></>,
              'Recurring tasks auto-recreate via background scheduler every 5 minutes. When a recurring task\'s period elapses, a new copy is created with status todo.',
              'Partial PUT: send only the fields you want to change — existing fields are preserved.',
            ],
          },
        ],
      },
      {
        heading: 'Examples',
        blocks: [
          {
            code: `# Create a weekly recurring task linked to a project
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
GET /api/v1/tasks?project_id={id}`,
          },
        ],
      },
    ],
  },
  {
    id: 'notes',
    title: 'Notes',
    subsections: [
      {
        blocks: [
          {
            body: [
              'Markdown notes with tags, folder organisation, full-text search (FTS5), and optional semantic search via embeddings.',
            ],
          },
        ],
      },
      {
        heading: 'Fields',
        blocks: [
          {
            table: {
              headers: ['Field', 'Type', 'Notes'],
              rows: [
                ['title', 'string', 'Required'],
                ['content', 'string', 'Markdown'],
                ['folder_id', 'string', 'Optional — links to a folder'],
                ['tags', 'string[]', 'Array of tag strings'],
              ],
            },
          },
        ],
      },
      {
        heading: 'Capabilities',
        blocks: [
          {
            bullets: [
              <>Filter by folder: <Code>GET /api/v1/notes?folder_id=&#123;id&#125;</Code></>,
              <>Full-text search (FTS5): <Code>GET /api/v1/notes/search?q=keyword</Code> — searches title + content.</>,
              <>Semantic search (RAG): <Code>POST /api/v1/notes/semantic-search</Code> — requires <Code>OPENAI_API_KEY</Code> and <Code>OPENAI_EMBEDDING_MODEL</Code>. Returns the most semantically similar notes.</>,
              'When a note is created or updated and RAG is configured, embedding indexing runs in the background — it doesn\'t block the response.',
            ],
          },
        ],
      },
      {
        heading: 'Examples',
        blocks: [
          {
            code: `# Create a note with tags and folder
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
{ "q": "ideas about product design", "top_k": 5 }`,
          },
        ],
      },
    ],
  },
  {
    id: 'folders',
    title: 'Folders',
    subsections: [
      {
        blocks: [
          {
            body: [
              'Folders organise notes into a hierarchy. Each folder can have a parent, enabling nested structures.',
            ],
          },
        ],
      },
      {
        heading: 'Fields',
        blocks: [
          {
            table: {
              headers: ['Field', 'Type', 'Notes'],
              rows: [
                ['name', 'string', 'Required'],
                ['parent_id', 'string | null', 'Optional — enables nesting'],
              ],
            },
          },
        ],
      },
      {
        heading: 'Endpoints',
        blocks: [
          {
            table: {
              headers: ['Method', 'Path', 'Description'],
              rows: [
                ['GET', '/api/v1/folders', 'List all folders'],
                ['POST', '/api/v1/folders', 'Create folder'],
                ['PUT', '/api/v1/folders/{id}', 'Rename or reparent'],
                ['DELETE', '/api/v1/folders/{id}', 'Delete folder'],
              ],
            },
          },
        ],
      },
      {
        heading: 'Examples',
        blocks: [
          {
            code: `# Create a root folder
POST /api/v1/folders
{ "name": "Work" }

# Create a nested folder
POST /api/v1/folders
{ "name": "Q4 Planning", "parent_id": "work-folder-uuid" }

# Then filter notes by folder
GET /api/v1/notes?folder_id={id}`,
          },
        ],
      },
    ],
  },
  {
    id: 'calendar',
    title: 'Calendar',
    subsections: [
      {
        blocks: [
          {
            body: ['Calendar events with full-day support, color coding, and date-range filtering.'],
          },
        ],
      },
      {
        heading: 'Fields',
        blocks: [
          {
            table: {
              headers: ['Field', 'Type', 'Notes'],
              rows: [
                ['title', 'string', 'Required'],
                ['description', 'string', 'Optional'],
                ['start_at', 'RFC3339', 'e.g. 2025-12-01T09:00:00Z'],
                ['end_at', 'RFC3339', 'Must be after start_at'],
                ['all_day', 'boolean', 'Hides time display when true'],
                ['color', 'string', 'Hex color — e.g. #6366f1'],
              ],
            },
          },
        ],
      },
      {
        heading: 'Capabilities',
        blocks: [
          {
            bullets: [
              <>Date-range filter: <Code>GET /api/v1/events?from=2025-01-01T00:00:00Z&to=2025-01-31T23:59:59Z</Code> — both params required for filtering, otherwise returns all events.</>,
              'All-day events set all_day: true — the calendar renders them without a time slot.',
            ],
          },
        ],
      },
      {
        heading: 'Examples',
        blocks: [
          {
            code: `# Create an all-day event
POST /api/v1/events
{
  "title": "Team offsite",
  "start_at": "2025-11-10T00:00:00Z",
  "end_at": "2025-11-11T00:00:00Z",
  "all_day": true,
  "color": "#f59e0b"
}

# Get events for a week
GET /api/v1/events?from=2025-11-10T00:00:00Z&to=2025-11-17T00:00:00Z`,
          },
        ],
      },
    ],
  },
  {
    id: 'projects',
    title: 'Projects',
    subsections: [
      {
        blocks: [
          {
            body: ['Projects group tasks together and track their lifecycle through three statuses.'],
          },
        ],
      },
      {
        heading: 'Fields',
        blocks: [
          {
            table: {
              headers: ['Field', 'Type', 'Values'],
              rows: [
                ['name', 'string', 'Required'],
                ['description', 'string', 'Optional'],
                ['status', 'enum', <><Code>active</Code> · <Code>paused</Code> · <Code>done</Code></>],
              ],
            },
          },
        ],
      },
      {
        heading: 'Capabilities',
        blocks: [
          {
            bullets: [
              'Tasks link to a project via project_id — one task belongs to at most one project.',
              <>Filter tasks by project: <Code>GET /api/v1/tasks?project_id=&#123;id&#125;</Code></>,
              'The AI agent\'s list_projects tool returns project details including task counts and progress.',
            ],
          },
        ],
      },
      {
        heading: 'Examples',
        blocks: [
          {
            code: `# Create a project
POST /api/v1/projects
{ "name": "Product launch", "status": "active" }

# Mark done
PUT /api/v1/projects/{id}
{ "status": "done" }

# Get all tasks for a project
GET /api/v1/tasks?project_id={id}`,
          },
        ],
      },
    ],
  },
  {
    id: 'apps',
    title: 'Apps',
    subsections: [
      {
        blocks: [
          {
            body: [
              'The app launcher holds shortcuts to any URL. Active apps are shown as cards; inactive ones are hidden from the launcher without being deleted.',
            ],
          },
        ],
      },
      {
        heading: 'Fields',
        blocks: [
          {
            table: {
              headers: ['Field', 'Type', 'Notes'],
              rows: [
                ['name', 'string', 'Required — display label'],
                ['url', 'string', 'Required — target URL'],
                ['description', 'string', 'Optional subtitle'],
                ['icon', 'string', 'Emoji or URL for the card icon'],
                ['color', 'string', 'Hex accent color for the card'],
                ['sort_order', 'number', 'Lower = first in the grid'],
                ['active', 'boolean', 'false hides without deleting'],
              ],
            },
          },
        ],
      },
      {
        heading: 'Examples',
        blocks: [
          {
            code: `# Add a new shortcut
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
{ "sort_order": 5 }`,
          },
        ],
      },
    ],
  },
  {
    id: 'wishlist',
    title: 'Wishlist',
    subsections: [
      {
        blocks: [
          {
            body: [
              'Track items you want to buy with target vs current price tracking, store references, and purchase status.',
            ],
          },
        ],
      },
      {
        heading: 'Fields',
        blocks: [
          {
            table: {
              headers: ['Field', 'Type', 'Notes'],
              rows: [
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
              ],
            },
          },
        ],
      },
      {
        heading: 'Examples',
        blocks: [
          {
            code: `# Add an item with price tracking
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
{ "current_price": 145 }`,
          },
        ],
      },
    ],
  },
  {
    id: 'diagrams',
    title: 'Diagrams',
    subsections: [
      {
        blocks: [
          {
            body: [
              'Visual network diagrams built with React Flow. Nodes and edges are stored as JSON blobs — the UI handles the canvas; the API handles persistence.',
            ],
          },
        ],
      },
      {
        heading: 'Fields',
        blocks: [
          {
            table: {
              headers: ['Field', 'Type', 'Notes'],
              rows: [
                ['name', 'string', 'Required — diagram label'],
                ['nodes', 'JSON', 'React Flow node array'],
                ['edges', 'JSON', 'React Flow edge array'],
              ],
            },
          },
        ],
      },
      {
        heading: 'Examples',
        blocks: [
          {
            code: `# Create an empty diagram
POST /api/v1/diagrams
{ "name": "Infrastructure", "nodes": [], "edges": [] }

# Save diagram state (full replace)
PUT /api/v1/diagrams/{id}
{
  "name": "Infrastructure",
  "nodes": [{ "id": "1", "data": { "label": "API" }, "position": { "x": 100, "y": 100 } }],
  "edges": []
}`,
          },
        ],
      },
    ],
  },
  {
    id: 'finance',
    title: 'Finance',
    subsections: [
      {
        blocks: [
          {
            body: [
              <>
                Read-only financial summary pulled from a{' '}
                <strong className="text-text-primary">Money Monkey</strong> instance.
                Configure with <Code>MONKEYAPI_URL</Code> and <Code>MONKEYAPI_KEY</Code>.
                Returns <Code>&#123;"configured": false&#125;</Code> if not set up.
              </>,
            ],
          },
        ],
      },
      {
        heading: 'Response shape',
        blocks: [
          {
            table: {
              headers: ['Field', 'Type', 'Description'],
              rows: [
                ['balance', 'number', 'Current account balance'],
                ['month_income', 'number', 'Total income this month'],
                ['month_expenses', 'number', 'Total expenses this month'],
                ['recent_transactions', 'array', 'Last N transactions'],
              ],
            },
          },
        ],
      },
      {
        heading: 'Transaction fields',
        blocks: [
          {
            table: {
              headers: ['Field', 'Type', 'Notes'],
              rows: [
                ['id', 'string', ''],
                ['amount', 'number', 'Positive = income, negative = expense'],
                ['type', 'string', 'Transaction type'],
                ['category', 'string', 'e.g. Food, Transport'],
                ['description', 'string', ''],
                ['date', 'string', 'ISO 8601 date'],
              ],
            },
          },
        ],
      },
      {
        heading: 'Example',
        blocks: [
          {
            code: `GET /api/v1/finance/summary

# Response
{
  "balance": 4250.00,
  "month_income": 5000.00,
  "month_expenses": 1820.50,
  "recent_transactions": [
    { "amount": -45.00, "category": "Food", "description": "Grocery run", "date": "2025-11-08" }
  ]
}`,
          },
        ],
      },
    ],
  },
  {
    id: 'cards',
    title: 'Card Tracker',
    subsections: [
      {
        blocks: [
          {
            body: [
              <>
                Read-only credit card summary pulled from a{' '}
                <strong className="text-text-primary">PayPinga</strong> instance.
                Configure with <Code>PAYPINGA_URL</Code> and <Code>PAYPINGA_KEY</Code>.
                Returns <Code>&#123;"configured": false&#125;</Code> if not set up.
              </>,
            ],
          },
        ],
      },
      {
        heading: 'Response shape',
        blocks: [
          {
            table: {
              headers: ['Field', 'Type', 'Description'],
              rows: [
                ['total_debt', 'number', 'Sum of all card balances'],
                ['total_payment', 'number', 'Total estimated payment due'],
                ['next_pay_date', 'string | null', 'Earliest upcoming payment date'],
                ['cards', 'array', 'Per-card breakdown'],
              ],
            },
          },
        ],
      },
      {
        heading: 'Card fields',
        blocks: [
          {
            table: {
              headers: ['Field', 'Type', 'Notes'],
              rows: [
                ['alias', 'string', 'Friendly name — e.g. "Chase Sapphire"'],
                ['bank', 'string', 'Issuing bank'],
                ['last4', 'string', 'Last 4 digits'],
                ['pay_date', 'string', 'Payment due date'],
                ['limit', 'number', 'Credit limit'],
                ['estimated_payment', 'number', 'Minimum / estimated payment'],
                ['balance', 'number', 'Current balance'],
              ],
            },
          },
        ],
      },
      {
        heading: 'Example',
        blocks: [
          {
            code: `GET /api/v1/cards/summary

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
}`,
          },
        ],
      },
    ],
  },
  {
    id: 'ai',
    title: 'AI Agent',
    subsections: [
      {
        blocks: [
          {
            body: [
              'The AI playground exposes a chat interface backed by Claude or any OpenAI-compatible provider. The agent has access to hub tools — it can read and write your data using natural language.',
            ],
          },
        ],
      },
      {
        heading: 'Configuration',
        blocks: [
          {
            table: {
              headers: ['Provider', 'Required vars'],
              rows: [
                ['Claude (Anthropic)', <><Code>ANTHROPIC_API_KEY</Code></>],
                ['OpenAI / OpenRouter / Ollama', <><Code>OPENAI_API_KEY</Code> · <Code>OPENAI_BASE_URL</Code> · <Code>OPENAI_MODEL</Code></>],
              ],
            },
          },
        ],
      },
      {
        heading: 'Available tools',
        blocks: [
          {
            table: {
              headers: ['Tool', 'What it does', 'Key params'],
              rows: [
                ['list_tasks', 'List tasks, optionally filtered by status', 'status: todo | in_progress | done'],
                ['create_task', 'Create a new task', 'title (req), description, priority, due_date, project_id'],
                ['update_task', 'Update status, title, or priority of a task', 'id (req), title, status, priority'],
                ['list_notes', 'List notes, optionally filtered by folder', 'folder_id'],
                ['create_note', 'Create a new Markdown note', 'title (req), content'],
                ['search_notes', 'Full-text search across all notes', 'query (req)'],
                ['list_projects', 'List all projects with task counts', '—'],
                ['list_events', 'List calendar events in a date range', 'from, to (YYYY-MM-DD, defaults: today + 30d)'],
                ['list_apps', 'List all launcher apps', '—'],
              ],
            },
          },
        ],
      },
      {
        heading: 'Chat endpoint',
        blocks: [
          {
            code: `POST /api/v1/ai/chat
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
}`,
          },
        ],
      },
      {
        heading: 'RAG — semantic search on notes',
        blocks: [
          {
            body: [
              <>
                When <Code>OPENAI_API_KEY</Code> and <Code>OPENAI_EMBEDDING_MODEL</Code> are set, notes
                are automatically indexed with embeddings on create/update. You can then run cosine
                similarity search independently of the AI chat:
              </>,
            ],
            code: `POST /api/v1/notes/semantic-search
{ "q": "ideas about system design", "top_k": 5 }
# top_k defaults to 5, max 50`,
          },
        ],
      },
    ],
  },
  {
    id: 'api',
    title: 'API Reference',
    subsections: [
      {
        blocks: [
          {
            body: [
              <>
                Base path: <Code>/api/v1</Code>. All endpoints accept and return JSON.
                Max request body: <Code>2 MB</Code>.
              </>,
            ],
          },
        ],
      },
      ...apiModules.map(([name, path]) => ({
        heading: <>{name} — <Code>{path}</Code></>,
        blocks: [{ table: { headers: ['Method', 'Path', 'Description'], rows: apiCrudRows(path) } }],
      })),
      {
        heading: 'Special endpoints',
        blocks: [
          {
            table: { headers: ['Method', 'Path', 'Description'], rows: apiSpecialRows },
          },
        ],
      },
    ],
  },
  {
    id: 'config',
    title: 'Configuration',
    subsections: [
      {
        blocks: [
          {
            body: [
              <>
                All configuration is done through environment variables. The server reads them at
                startup — a restart is required when changing credentials. hube auto-loads a{' '}
                <Code>.env</Code> file at the project root if present.
              </>,
            ],
            table: { headers: ['Variable', 'Default', 'Description'], rows: configRows },
          },
        ],
      },
    ],
  },
  {
    id: 'running',
    title: 'Running',
    subsections: [
      {
        heading: 'Local development',
        blocks: [
          {
            code: `# Backend
cd backend
go run ./cmd/api

# Frontend (separate terminal)
cd frontend
pnpm install
pnpm dev`,
          },
        ],
      },
      {
        heading: 'Docker (development)',
        blocks: [{ code: `docker compose up` }],
      },
      {
        heading: 'Docker (production)',
        blocks: [{ code: `docker compose -f docker-compose.prod.yml up -d` }],
      },
      {
        heading: 'Reverse proxy with Caddy',
        blocks: [
          {
            body: [
              <>
                The repo includes a <Code>Caddyfile</Code> for TLS termination and proxying.
                Set <Code>HUBE_DOMAIN</Code> to your public domain so CORS is configured correctly.
              </>,
            ],
            code: `# Caddyfile snippet
hub.example.com {
  reverse_proxy localhost:8080
}`,
          },
        ],
      },
    ],
  },
  {
    id: 'backup',
    title: 'Backup & Export',
    subsections: [
      {
        heading: 'Automatic backup',
        blocks: [
          {
            body: [
              <>
                SQLite is a single file — zero setup. Automatic backup runs every{' '}
                <strong className="text-text-primary">12 hours</strong> via{' '}
                <Code>VACUUM INTO</Code>, which is safe under concurrent writes. The{' '}
                <Code>BACKUP_RETAIN</Code> env var controls how many files to keep (default 7).
              </>,
            ],
          },
        ],
      },
      {
        heading: 'Manual backup',
        blocks: [
          {
            code: `POST /api/v1/backup
# Triggers an immediate VACUUM INTO backup — returns 200 on success`,
          },
        ],
      },
      {
        heading: 'Export all data',
        blocks: [
          {
            code: `GET /api/v1/export
# Downloads a ZIP archive`,
          },
          {
            table: {
              headers: ['Content', 'Format'],
              rows: [
                ['Notes', 'Markdown (.md)'],
                ['Events', 'iCalendar (.ics)'],
                ['Everything else', 'JSON'],
              ],
            },
          },
        ],
      },
    ],
  },
  {
    id: 'security',
    title: 'Security',
    subsections: [
      {
        blocks: [{ bullets: securityItems }],
      },
    ],
  },
]
