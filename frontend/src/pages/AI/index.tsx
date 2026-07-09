import { useState, useRef, useEffect } from 'react'
import { Sparkles, Send, Square } from 'lucide-react'
import { API_BASE } from '../../services/api'
import ReactMarkdown from 'react-markdown'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'

type Role = 'user' | 'assistant'

interface Message {
  id: string
  role: Role
  content: string
  toolCalls?: { tool: string; done: boolean }[]
  streaming?: boolean
}

interface SSEEvent {
  type: 'text' | 'tool_use' | 'tool_result' | 'done' | 'error'
  content?: string
  tool?: string
  error?: string
}

const TOOL_LABELS: Record<string, string> = {
  list_tasks: 'Reading tasks',
  create_task: 'Creating task',
  update_task: 'Updating task',
  search_notes: 'Searching notes',
  list_notes: 'Reading notes',
  create_note: 'Creating note',
  list_projects: 'Reading projects',
  list_events: 'Reading calendar',
  list_apps: 'Reading apps',
}

const SUGGESTIONS = [
  'What are my pending tasks?',
  'Search my notes for "architecture"',
  'Show me my projects and their progress',
  "What's on my calendar this week?",
  'Create a task: Review pull requests',
]

type Provider = 'auto' | 'anthropic' | 'openai'

const PROVIDERS: { value: Provider; label: string }[] = [
  { value: 'auto', label: 'Auto' },
  { value: 'anthropic', label: 'Claude' },
  { value: 'openai', label: 'OpenAI-compatible' },
]

export function AIPage() {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [provider, setProvider] = useState<Provider>('auto')
  const bottomRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  async function sendMessage(text: string) {
    if (!text.trim() || loading) return

    const userMsg: Message = {
      id: crypto.randomUUID(),
      role: 'user',
      content: text.trim(),
    }
    const assistantId = crypto.randomUUID()
    const assistantMsg: Message = {
      id: assistantId,
      role: 'assistant',
      content: '',
      toolCalls: [],
      streaming: true,
    }

    setMessages(prev => [...prev, userMsg, assistantMsg])
    setInput('')
    setLoading(true)

    // Build history for the request (exclude the new streaming message)
    const history = [...messages, userMsg].map(m => ({
      role: m.role,
      content: m.content,
    }))

    try {
      const response = await fetch(`${API_BASE}/ai/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ messages: history, provider: provider === 'auto' ? '' : provider }),
      })

      if (!response.ok) {
        const errText = await response.text()
        throw new Error(errText || `HTTP ${response.status}`)
      }

      const reader = response.body!.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() ?? ''

        for (const line of lines) {
          if (!line.trim()) continue
          try {
            const event: SSEEvent = JSON.parse(line)
            handleEvent(assistantId, event)
          } catch {
            // ignore malformed lines
          }
        }
      }
    } catch (err) {
      setMessages(prev =>
        prev.map(m =>
          m.id === assistantId
            ? { ...m, content: `Error: ${err instanceof Error ? err.message : 'Unknown error'}`, streaming: false }
            : m
        )
      )
    } finally {
      setLoading(false)
      setMessages(prev =>
        prev.map(m => (m.id === assistantId ? { ...m, streaming: false } : m))
      )
    }
  }

  function handleEvent(id: string, event: SSEEvent) {
    setMessages(prev =>
      prev.map(m => {
        if (m.id !== id) return m
        switch (event.type) {
          case 'text':
            return { ...m, content: m.content + (event.content ?? '') }
          case 'tool_use':
            return {
              ...m,
              toolCalls: [...(m.toolCalls ?? []), { tool: event.tool!, done: false }],
            }
          case 'tool_result':
            return {
              ...m,
              toolCalls: (m.toolCalls ?? []).map(tc =>
                tc.tool === event.tool && !tc.done ? { ...tc, done: true } : tc
              ),
            }
          case 'error':
            return { ...m, content: `Error: ${event.error ?? 'Unknown error'}`, streaming: false }
          default:
            return m
        }
      })
    )
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendMessage(input)
    }
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="px-6 py-4 border-b border-border bg-surface-elevated shrink-0">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-(--color-accent)/10 flex items-center justify-center shrink-0">
            <Sparkles size={18} className="text-(--color-accent)" />
          </div>
          <div>
            <h1 className="text-base font-semibold text-text-primary">IA Playground</h1>
            <p className="text-xs text-text-muted">
              Acceso completo al hub — tareas, notas, proyectos, calendario
            </p>
          </div>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-4 py-6">
        <div className="max-w-3xl mx-auto space-y-6">
          {messages.length === 0 && (
            <div className="text-center py-16">
              <div className="mb-4 flex justify-center">
                <div className="w-14 h-14 rounded-2xl bg-(--color-accent)/10 flex items-center justify-center">
                  <Sparkles size={28} className="text-(--color-accent)" />
                </div>
              </div>
              <p className="text-text-muted text-sm mb-8">
                Ask anything about your hub. I can read and update your data.
              </p>
              <div className="flex flex-wrap justify-center gap-2">
                {SUGGESTIONS.map(s => (
                  <button
                    key={s}
                    onClick={() => sendMessage(s)}
                    className="text-xs px-3 py-1.5 rounded-full bg-surface-elevated border border-border text-text-secondary hover:border-(--color-accent) hover:text-text-primary transition-colors"
                  >
                    {s}
                  </button>
                ))}
              </div>
            </div>
          )}

          {messages.map(msg => (
            <div key={msg.id} className={`flex gap-3 ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
              {msg.role === 'assistant' && (
                <div className="w-7 h-7 rounded-full bg-(--color-accent) flex items-center justify-center shrink-0 mt-0.5">
                  <Sparkles size={14} className="text-white" />
                </div>
              )}

              <div className={`max-w-[80%] ${msg.role === 'user' ? 'order-first' : ''}`}>
                {/* Tool calls */}
                {msg.toolCalls && msg.toolCalls.length > 0 && (
                  <div className="mb-2 space-y-1">
                    {msg.toolCalls.map((tc, i) => (
                      <div key={i} className="flex items-center gap-2 text-xs text-text-muted">
                        <span className={`w-3 h-3 rounded-full ${tc.done ? 'bg-emerald-500' : 'bg-(--color-accent) animate-pulse'}`} />
                        {TOOL_LABELS[tc.tool] ?? tc.tool}
                      </div>
                    ))}
                  </div>
                )}

                {/* Bubble */}
                {msg.role === 'user' ? (
                  <div className="px-4 py-2.5 rounded-2xl rounded-tr-sm bg-(--color-accent) text-white text-sm">
                    {msg.content}
                  </div>
                ) : (
                  <div className="text-sm text-text-secondary prose prose-sm dark:prose-invert max-w-none">
                    {msg.content ? (
                      <MarkdownContent content={msg.content} />
                    ) : msg.streaming ? (
                      <span className="inline-block w-2 h-4 bg-(--color-accent) animate-pulse rounded-sm" />
                    ) : null}
                  </div>
                )}
              </div>

              {msg.role === 'user' && (
                <div className="w-7 h-7 rounded-full bg-surface-elevated flex items-center justify-center shrink-0 mt-0.5">
                  <span className="text-text-muted text-xs">U</span>
                </div>
              )}
            </div>
          ))}

          <div ref={bottomRef} />
        </div>
      </div>

      {/* Input */}
      <div className="shrink-0 px-4 py-4 border-t border-border bg-surface-elevated">
        <div className="max-w-3xl mx-auto">
          {/* Provider selector */}
          <div className="flex gap-1 mb-2">
            {PROVIDERS.map(p => (
              <button
                key={p.value}
                onClick={() => setProvider(p.value)}
                className={`px-2.5 py-1 rounded-md text-xs transition-colors ${
                  provider === p.value
                    ? 'bg-(--color-accent) text-white'
                    : 'text-text-muted hover:text-text-secondary'
                }`}
              >
                {p.label}
              </button>
            ))}
          </div>
          <div className="flex gap-3 items-end">
            <textarea
              ref={inputRef}
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Ask anything… (Enter to send, Shift+Enter for newline)"
              rows={1}
              className="flex-1 resize-none bg-surface-base border border-border rounded-xl px-4 py-3 text-sm text-text-primary placeholder-text-muted focus:outline-none focus:border-(--color-accent) transition-colors"
              style={{ maxHeight: '140px', overflowY: 'auto' }}
            />
            <button
              onClick={() => sendMessage(input)}
              disabled={!input.trim() || loading}
              className="w-10 h-10 rounded-xl bg-(--color-accent) hover:bg-(--color-accent-hover) disabled:opacity-40 text-white flex items-center justify-center transition-colors shrink-0"
              aria-label="Send message"
            >
              {loading ? <Square size={14} /> : <Send size={14} />}
            </button>
          </div>
          <p className="text-xs text-text-muted mt-2 text-center">
            tools: tasks, notes, projects, calendar, apps
          </p>
        </div>
      </div>
    </div>
  )
}

function MarkdownContent({ content }: { content: string }) {
  return (
    <ReactMarkdown
      components={{
        code({ className, children, ...props }) {
          const match = /language-(\w+)/.exec(className || '')
          const isBlock = match !== null
          return isBlock ? (
            <SyntaxHighlighter style={oneDark} language={match[1]} PreTag="div">
              {String(children).replace(/\n$/, '')}
            </SyntaxHighlighter>
          ) : (
            <code className="bg-surface-elevated px-1 py-0.5 rounded text-pink-400 text-xs" {...props}>
              {children}
            </code>
          )
        },
      }}
    >
      {content}
    </ReactMarkdown>
  )
}
