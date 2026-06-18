import { useState, useEffect } from 'react'
import { useSettings, useUpdateSettings } from '../../hooks/useSettings'
import type { Settings } from '../../hooks/useSettings'
import { useTheme } from '../../contexts/ThemeContext'
import type { ThemeMode, AccentColor } from '../../contexts/ThemeContext'

function StatusDot({ active }: { active: boolean }) {
  return (
    <span className={`inline-block w-2 h-2 rounded-full ${active ? 'bg-green-400' : 'bg-gray-600'}`} />
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-6 mb-4">
      <h2 className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-5">{title}</h2>
      {children}
    </div>
  )
}

function Field({
  label,
  hint,
  children,
}: {
  label: string
  hint?: string
  children: React.ReactNode
}) {
  return (
    <div className="mb-4 last:mb-0">
      <label className="block text-sm text-gray-700 dark:text-gray-300 mb-1">{label}</label>
      {hint && <p className="text-xs text-gray-400 dark:text-gray-600 mb-1.5">{hint}</p>}
      {children}
    </div>
  )
}

const INPUT = 'w-full bg-gray-100 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-900 dark:text-white focus:outline-none focus:border-(--color-accent) placeholder-gray-400 dark:placeholder-gray-600'

const MODES: { value: ThemeMode; label: string; icon: string }[] = [
  { value: 'dark', label: 'Dark', icon: '🌙' },
  { value: 'light', label: 'Light', icon: '☀️' },
  { value: 'system', label: 'System', icon: '💻' },
]

const ACCENTS: { value: AccentColor; color: string }[] = [
  { value: 'indigo',  color: 'oklch(58.75% 0.265 296.96)' },
  { value: 'violet',  color: 'oklch(55.67% 0.265 303.01)' },
  { value: 'rose',    color: 'oklch(59.38% 0.24  12.25)'  },
  { value: 'emerald', color: 'oklch(60.59% 0.185 152.63)' },
  { value: 'sky',     color: 'oklch(62.72% 0.198 237.08)' },
  { value: 'orange',  color: 'oklch(65.87% 0.216  47.40)' },
]

export function SettingsPage() {
  const { data, isLoading } = useSettings()
  const update = useUpdateSettings()
  const { mode, accent, setMode, setAccent } = useTheme()
  const [form, setForm] = useState<Settings | null>(null)
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    if (data && !form) setForm(data)
  }, [data])

  if (isLoading || !form) return <div className="p-6 text-gray-500 text-sm">Loading…</div>

  function setGeneral(patch: Partial<Settings['general']>) {
    setForm(f => f ? { ...f, general: { ...f.general, ...patch } } : f)
  }

  function setIntegration(patch: Partial<Settings['integrations']>) {
    setForm(f => f ? { ...f, integrations: { ...f.integrations, ...patch } } : f)
  }

  function handleSave() {
    if (!form) return
    update.mutate(form, {
      onSuccess: () => {
        setSaved(true)
        setTimeout(() => setSaved(false), 2000)
      },
    })
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">Settings</h1>
        <button
          onClick={handleSave}
          disabled={update.isPending}
          className="px-4 py-2 bg-(--color-accent) hover:bg-(--color-accent-hover) disabled:opacity-50 text-white text-sm rounded-lg transition-colors"
        >
          {saved ? '✓ Saved' : update.isPending ? 'Saving…' : 'Save changes'}
        </button>
      </div>

      {/* Appearance */}
      <Section title="Appearance">
        <Field label="Mode">
          <div className="flex gap-2">
            {MODES.map(m => (
              <button
                key={m.value}
                onClick={() => setMode(m.value)}
                className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm border transition-colors ${
                  mode === m.value
                    ? 'border-(--color-accent) bg-(--color-accent)/10 text-gray-900 dark:text-white'
                    : 'border-gray-200 dark:border-gray-700 text-gray-500 dark:text-gray-400 hover:border-gray-300 dark:hover:border-gray-600'
                }`}
              >
                <span>{m.icon}</span>
                {m.label}
              </button>
            ))}
          </div>
        </Field>
        <Field label="Accent color">
          <div className="flex gap-2">
            {ACCENTS.map(a => (
              <button
                key={a.value}
                onClick={() => setAccent(a.value)}
                title={a.value}
                className={`w-8 h-8 rounded-full border-2 transition-transform ${
                  accent === a.value ? 'border-white dark:border-white scale-110' : 'border-transparent'
                }`}
                style={{ backgroundColor: a.color }}
              />
            ))}
          </div>
        </Field>
      </Section>

      <Section title="General">
        <Field label="Display name" hint="Shown in the sidebar header.">
          <input
            className={INPUT}
            value={form.general.display_name}
            onChange={e => setGeneral({ display_name: e.target.value })}
            placeholder="e.g. Iosif"
          />
        </Field>
      </Section>

      <Section title="Integrations">
        <p className="text-xs text-gray-400 dark:text-gray-600 mb-5">
          Changes are stored in the database. Restart the server to apply new credentials to active clients.
        </p>

        {/* Money Monkey */}
        <div className="mb-6 pb-6 border-b border-gray-100 dark:border-gray-800">
          <div className="flex items-center gap-2 mb-3">
            <StatusDot active={form.integrations.monkeyapi_enabled} />
            <span className="text-sm font-medium text-gray-800 dark:text-gray-200">Money Monkey</span>
            <span className={`text-xs ml-1 ${form.integrations.monkeyapi_enabled ? 'text-green-500' : 'text-gray-600'}`}>
              {form.integrations.monkeyapi_enabled ? 'Connected' : 'Not configured'}
            </span>
          </div>
          <Field label="API URL">
            <input
              className={INPUT}
              value={form.integrations.monkeyapi_url}
              onChange={e => setIntegration({ monkeyapi_url: e.target.value })}
              placeholder="https://your-app.vercel.app"
            />
          </Field>
          <Field label="API Key">
            <input
              className={INPUT}
              type="password"
              value={form.integrations.monkeyapi_key}
              onChange={e => setIntegration({ monkeyapi_key: e.target.value })}
              placeholder="Leave unchanged to keep current key"
            />
          </Field>
        </div>

        {/* PayPinga */}
        <div>
          <div className="flex items-center gap-2 mb-3">
            <StatusDot active={form.integrations.paypinga_enabled} />
            <span className="text-sm font-medium text-gray-800 dark:text-gray-200">PayPinga</span>
            <span className={`text-xs ml-1 ${form.integrations.paypinga_enabled ? 'text-green-500' : 'text-gray-600'}`}>
              {form.integrations.paypinga_enabled ? 'Connected' : 'Not configured'}
            </span>
          </div>
          <Field label="API URL">
            <input
              className={INPUT}
              value={form.integrations.paypinga_url}
              onChange={e => setIntegration({ paypinga_url: e.target.value })}
              placeholder="https://your-app.vercel.app"
            />
          </Field>
          <Field label="API Key">
            <input
              className={INPUT}
              type="password"
              value={form.integrations.paypinga_key}
              onChange={e => setIntegration({ paypinga_key: e.target.value })}
              placeholder="Leave unchanged to keep current key"
            />
          </Field>
        </div>
      </Section>

      <Section title="System">
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <p className="text-gray-600 text-xs mb-0.5">Backend</p>
            <p className="text-gray-700 dark:text-gray-300">Go + chi + SQLite</p>
          </div>
          <div>
            <p className="text-gray-400 dark:text-gray-600 text-xs mb-0.5">Frontend</p>
            <p className="text-gray-700 dark:text-gray-300">React + Vite + TanStack Query</p>
          </div>
          <div>
            <p className="text-gray-400 dark:text-gray-600 text-xs mb-0.5">Database</p>
            <p className="text-gray-700 dark:text-gray-300">SQLite (FTS5 + WAL)</p>
          </div>
          <div>
            <p className="text-gray-400 dark:text-gray-600 text-xs mb-0.5">Version</p>
            <p className="text-gray-700 dark:text-gray-300">0.1.0</p>
          </div>
        </div>
      </Section>
    </div>
  )
}
