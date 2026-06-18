import { useState, useRef, useEffect } from 'react'
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
      const response = await fetch('/api/v1/ai/chat', {
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
    <div className="flex flex-col h-full bg-gray-50 dark:bg-gray-950">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 shrink-0">
        <div className="flex items-center gap-3">
          <span className="text-xl">✦</span>
          <div>
            <h1 className="text-base font-semibold text-gray-900 dark:text-white">IA Playground</h1>
            <p className="text-xs text-gray-400 dark:text-gray-500">
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
              <div className="text-5xl mb-4">✦</div>
              <p className="text-gray-500 dark:text-gray-400 text-sm mb-8">
                Ask anything about your hub. I can read and update your data.
              </p>
              <div className="flex flex-wrap justify-center gap-2">
                {SUGGESTIONS.map(s => (
                  <button
                    key={s}
                    onClick={() => sendMessage(s)}
                    className="text-xs px-3 py-1.5 rounded-full bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:border-(--color-accent) hover:text-gray-900 dark:hover:text-white transition-colors"
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
                  <span className="text-white text-xs">✦</span>
                </div>
              )}

              <div className={`max-w-[80%] ${msg.role === 'user' ? 'order-first' : ''}`}>
                {/* Tool calls */}
                {msg.toolCalls && msg.toolCalls.length > 0 && (
                  <div className="mb-2 space-y-1">
                    {msg.toolCalls.map((tc, i) => (
                      <div key={i} className="flex items-center gap-2 text-xs text-gray-400 dark:text-gray-500">
                        <span className={`w-3 h-3 rounded-full ${tc.done ? 'bg-green-500' : 'bg-(--color-accent) animate-pulse'}`} />
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
                  <div className="text-sm text-gray-800 dark:text-gray-200 prose prose-sm dark:prose-invert max-w-none">
                    {msg.content ? (
                      <MarkdownContent content={msg.content} />
                    ) : msg.streaming ? (
                      <span className="inline-block w-2 h-4 bg-(--color-accent) animate-pulse rounded-sm" />
                    ) : null}
                  </div>
                )}
              </div>

              {msg.role === 'user' && (
                <div className="w-7 h-7 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center shrink-0 mt-0.5">
                  <span className="text-gray-600 dark:text-gray-300 text-xs">U</span>
                </div>
              )}
            </div>
          ))}

          <div ref={bottomRef} />
        </div>
      </div>

      {/* Input */}
      <div className="shrink-0 px-4 py-4 border-t border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
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
                    : 'text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300'
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
              className="flex-1 resize-none bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl px-4 py-3 text-sm text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:border-(--color-accent) transition-colors"
              style={{ maxHeight: '140px', overflowY: 'auto' }}
            />
            <button
              onClick={() => sendMessage(input)}
              disabled={!input.trim() || loading}
              className="px-4 py-3 rounded-xl bg-(--color-accent) hover:bg-(--color-accent-hover) disabled:opacity-40 text-white text-sm font-medium transition-colors shrink-0"
            >
              {loading ? '…' : '↑'}
            </button>
          </div>
          <p className="text-xs text-gray-400 dark:text-gray-600 mt-2 text-center">
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
            <code className="bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded text-pink-500 dark:text-pink-400 text-xs" {...props}>
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
