import { useState, useEffect, useRef } from 'react'
import axios from 'axios'
import { useSettings, useUpdateSettings } from '../../hooks/useSettings'
import type { Settings } from '../../hooks/useSettings'
import { useTheme } from '../../contexts/ThemeContext'
import { Check, Moon, Sun, Monitor } from 'lucide-react'
import { PageHeader } from '../../components/molecules/PageHeader'
import { Input } from '../../components/atoms/Input'
import type { ThemeMode, ThemeName } from '../../contexts/ThemeContext'

function StatusDot({ active }: { active: boolean }) {
  return (
    <span className={`inline-block w-2 h-2 rounded-full ${active ? 'bg-emerald-400' : 'bg-text-muted'}`} />
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-surface-elevated border border-border rounded-xl p-6 mb-4">
      <h2 className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-5">{title}</h2>
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
      <label className="block text-sm text-text-secondary mb-1">{label}</label>
      {hint && <p className="text-xs text-text-muted mb-1.5">{hint}</p>}
      {children}
    </div>
  )
}

import type { ReactNode } from 'react'

const MODES: { value: ThemeMode; label: string; icon: ReactNode }[] = [
  { value: 'dark', label: 'Dark', icon: <Moon size={16} /> },
  { value: 'light', label: 'Light', icon: <Sun size={16} /> },
  { value: 'system', label: 'System', icon: <Monitor size={16} /> },
]

const THEMES: { value: ThemeName; label: string; surface: string; accent: string }[] = [
  { value: 'solaris',          label: 'Solaris',         surface: 'oklch(22% 0.025 55)',   accent: 'oklch(58.75% 0.265 296.96)' },
  { value: 'apple-inspired',   label: 'Apple Inspired',  surface: 'oklch(16% 0.005 250)',  accent: 'oklch(55% 0.15 250)' },
  { value: 'minimal-tech',     label: 'Minimal Tech',    surface: 'oklch(20% 0.01 260)',   accent: 'oklch(60% 0.18 240)' },
  { value: 'glassmorphism',    label: 'Glassmorphism',   surface: 'oklch(23% 0.02 280)',   accent: 'oklch(70% 0.15 280)' },
  { value: 'neobrutalism',     label: 'Neobrutalism',    surface: 'oklch(20% 0.005 0)',    accent: 'oklch(80% 0.15 90)' },
  { value: 'synthwave',        label: 'Synthwave',       surface: 'oklch(18% 0.04 300)',   accent: 'oklch(65% 0.25 330)' },
  { value: 'cyberpunk',        label: 'Cyberpunk',       surface: 'oklch(16% 0.02 160)',   accent: 'oklch(75% 0.2 140)' },
  { value: 'solarpunk',        label: 'Solarpunk',       surface: 'oklch(23% 0.03 80)',    accent: 'oklch(70% 0.2 70)' },
  { value: 'frutiger-aero',    label: 'Frutiger Aero',   surface: 'oklch(20% 0.02 220)',   accent: 'oklch(65% 0.15 200)' },
  { value: 'japandi',          label: 'Japandi',         surface: 'oklch(24% 0.01 70)',    accent: 'oklch(55% 0.12 40)' },
  { value: 'dark-academia',    label: 'Dark Academia',    surface: 'oklch(20% 0.02 50)',    accent: 'oklch(65% 0.12 80)' },
  { value: 'luxury-black-gold',label: 'Luxury Black Gold',surface: 'oklch(14% 0.005 0)',    accent: 'oklch(75% 0.15 85)' },
  { value: 'neo-tokyo',        label: 'Neo Tokyo',       surface: 'oklch(18% 0.02 280)',   accent: 'oklch(65% 0.15 210)' },
  { value: 'tulum-boho',       label: 'Tulum Boho',      surface: 'oklch(24% 0.02 60)',    accent: 'oklch(65% 0.15 30)' },
  { value: 'retro-terminal',   label: 'Retro Terminal',  surface: 'oklch(13% 0.01 150)',   accent: 'oklch(75% 0.2 140)' },
  { value: 'space-opera',      label: 'Space Opera',     surface: 'oklch(13% 0.03 260)',   accent: 'oklch(65% 0.2 290)' },
]

const api = axios.create({ baseURL: '/api/v1' })

function EmailSection() {
  const toRef = useRef<HTMLInputElement>(null)
  const [status, setStatus] = useState<'idle' | 'sending' | 'ok' | 'err'>('idle')
  const [errMsg, setErrMsg] = useState('')

  async function sendDigest() {
    const to = toRef.current?.value.trim()
    if (!to) return
    setStatus('sending')
    try {
      await api.post('/email/digest', { to: [to] })
      setStatus('ok')
      setTimeout(() => setStatus('idle'), 3000)
    } catch (e: unknown) {
      const msg = axios.isAxiosError(e) ? e.response?.data?.error ?? e.message : String(e)
      setErrMsg(msg)
      setStatus('err')
      setTimeout(() => setStatus('idle'), 4000)
    }
  }

  return (
    <Section title="Email digest">
      <p className="text-xs text-text-muted mb-4">
        Sends a plain-text task digest to the specified address. Configure SMTP via{' '}
        <code className="bg-surface-base px-1 rounded text-[11px]">SMTP_HOST / SMTP_USER / SMTP_PASS / SMTP_FROM</code>{' '}
        environment variables on the server.
      </p>
      <div className="flex gap-2">
        <Input
          ref={toRef}
          type="email"
          placeholder="you@example.com"
          className="flex-1"
        />
        <button
          onClick={sendDigest}
          disabled={status === 'sending'}
          className="px-4 py-2 bg-(--color-accent) hover:bg-(--color-accent-hover) disabled:opacity-50 text-white text-sm rounded-lg transition-colors whitespace-nowrap"
        >
          {status === 'sending' ? 'Sending…' : status === 'ok' ? <><Check size={14} className="inline" /> Sent</> : 'Send digest'}
        </button>
      </div>
      {status === 'err' && (
        <p className="text-xs text-red-400 mt-2">{errMsg}</p>
      )}
    </Section>
  )
}

function ExportSection() {
  const [status, setStatus] = useState<'idle' | 'exporting' | 'err'>('idle')
  const [errMsg, setErrMsg] = useState('')

  async function handleExport() {
    setStatus('exporting')
    try {
      const res = await fetch('/api/v1/export')
      if (!res.ok) {
        const body = await res.json().catch(() => ({ error: res.statusText }))
        throw new Error(body?.error ?? res.statusText)
      }
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `hube_export_${new Date().toISOString().slice(0, 10)}.zip`
      a.click()
      URL.revokeObjectURL(url)
      setStatus('idle')
    } catch (e: unknown) {
      setErrMsg(e instanceof Error ? e.message : String(e))
      setStatus('err')
      setTimeout(() => setStatus('idle'), 4000)
    }
  }

  return (
    <Section title="Export data">
      <p className="text-xs text-text-muted mb-4">
        Download all your data as a ZIP archive — notes (Markdown), tasks, events (iCal), apps,
        wishlist, projects, and diagrams (JSON).
      </p>
      <button
        onClick={handleExport}
        disabled={status === 'exporting'}
        className="px-4 py-2 bg-(--color-accent) hover:bg-(--color-accent-hover) disabled:opacity-50 text-white text-sm rounded-lg transition-colors"
      >
        {status === 'exporting' ? 'Exporting…' : 'Export data'}
      </button>
      {status === 'err' && (
        <p className="text-xs text-red-400 mt-2">{errMsg}</p>
      )}
    </Section>
  )
}

export function SettingsPage() {
  const { data, isLoading } = useSettings()
  const update = useUpdateSettings()
  const { mode, theme, setMode, setTheme } = useTheme()
  const [form, setForm] = useState<Settings | null>(null)
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    if (data && !form) setForm(data)
  }, [data])

  if (isLoading || !form) return <div className="p-6 text-text-muted text-sm">Loading…</div>

  function setGeneral(patch: Partial<Settings['general']>) {
    setForm(f => f ? { ...f, general: { ...f.general, ...patch } } : f)
  }

  function setIntegration(patch: Partial<Settings['integrations']>) {
    setForm(f => f ? { ...f, integrations: { ...f.integrations, ...patch } } : f)
  }

  function handleSave() {
    if (!form) return
    // Re-sync view_preferences from the latest react-query cache before submitting:
    // useViewPreference performs its own PUT whenever a view is toggled, and `form`
    // was only seeded once at mount, so it can otherwise revert a fresher preference.
    const payload: Settings = data
      ? { ...form, general: { ...form.general, view_preferences: data.general.view_preferences } }
      : form
    update.mutate(payload, {
      onSuccess: () => {
        setSaved(true)
        setTimeout(() => setSaved(false), 2000)
      },
    })
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <PageHeader
        title="Settings"
        actions={
          <button
            onClick={handleSave}
            disabled={update.isPending}
            className="px-4 py-2 bg-(--color-accent) hover:bg-(--color-accent-hover) disabled:opacity-50 text-white text-sm rounded-lg transition-colors"
          >
            {saved ? <><Check size={14} className="inline" /> Saved</> : update.isPending ? 'Saving…' : 'Save changes'}
          </button>
        }
      />

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
                    ? 'border-(--color-accent) bg-(--color-accent)/10 text-text-primary'
                    : 'border-border text-text-muted hover:border-border-subtle'
                }`}
              >
                {m.icon}
                {m.label}
              </button>
            ))}
          </div>
        </Field>
        <Field label="Theme" hint="Choose a visual theme for the entire interface.">
          <div className="grid grid-cols-4 gap-2">
            {THEMES.map(t => (
              <button
                key={t.value}
                onClick={() => setTheme(t.value)}
                className={`rounded-xl border p-3 text-left transition-all ${
                  theme === t.value
                    ? 'border-(--color-accent) ring-1 ring-(--color-accent) bg-(--color-accent)/5'
                    : 'border-border hover:border-border-subtle bg-surface-base/50'
                }`}
              >
                {/* Color preview */}
                <div className="flex gap-1.5 mb-2">
                  <span
                    className="w-5 h-5 rounded-full border border-white/10"
                    style={{ backgroundColor: t.surface }}
                  />
                  <span
                    className="w-5 h-5 rounded-full border border-white/10"
                    style={{ backgroundColor: t.accent }}
                  />
                </div>
                {/* Theme name */}
                <span className={`text-xs font-medium leading-tight ${
                  theme === t.value ? 'text-text-primary' : 'text-text-secondary'
                }`}>
                  {t.label}
                </span>
              </button>
            ))}
          </div>
        </Field>
      </Section>

      <Section title="General">
        <Field label="Display name" hint="Shown in the sidebar header.">
          <Input
            value={form.general.display_name}
            onChange={e => setGeneral({ display_name: (e.target as HTMLInputElement).value })}
            placeholder="e.g. Iosif"
          />
        </Field>
      </Section>

      <Section title="Integrations">
        <p className="text-xs text-text-muted mb-5">
          Changes are stored in the database. Restart the server to apply new credentials to active clients.
        </p>

        {/* Money Monkey */}
        <div className="mb-6 pb-6 border-b border-border">
          <div className="flex items-center gap-2 mb-3">
            <StatusDot active={form.integrations.monkeyapi_enabled} />
            <span className="text-sm font-medium text-text-primary">Money Monkey</span>
            <span className={`text-xs ml-1 ${form.integrations.monkeyapi_enabled ? 'text-emerald-500' : 'text-text-muted'}`}>
              {form.integrations.monkeyapi_enabled ? 'Connected' : 'Not configured'}
            </span>
          </div>
          <Field label="API URL">
            <Input
              value={form.integrations.monkeyapi_url}
              onChange={e => setIntegration({ monkeyapi_url: (e.target as HTMLInputElement).value })}
              placeholder="https://your-app.vercel.app"
            />
          </Field>
          <Field label="API Key">
            <Input
              type="password"
              value={form.integrations.monkeyapi_key}
              onChange={e => setIntegration({ monkeyapi_key: (e.target as HTMLInputElement).value })}
              placeholder="Leave unchanged to keep current key"
            />
          </Field>
        </div>

        {/* PayPinga */}
        <div>
          <div className="flex items-center gap-2 mb-3">
            <StatusDot active={form.integrations.paypinga_enabled} />
            <span className="text-sm font-medium text-text-primary">PayPinga</span>
            <span className={`text-xs ml-1 ${form.integrations.paypinga_enabled ? 'text-emerald-500' : 'text-text-muted'}`}>
              {form.integrations.paypinga_enabled ? 'Connected' : 'Not configured'}
            </span>
          </div>
          <Field label="API URL">
            <Input
              value={form.integrations.paypinga_url}
              onChange={e => setIntegration({ paypinga_url: (e.target as HTMLInputElement).value })}
              placeholder="https://your-app.vercel.app"
            />
          </Field>
          <Field label="API Key">
            <Input
              type="password"
              value={form.integrations.paypinga_key}
              onChange={e => setIntegration({ paypinga_key: (e.target as HTMLInputElement).value })}
              placeholder="Leave unchanged to keep current key"
            />
          </Field>
        </div>
      </Section>

      <EmailSection />

      <ExportSection />

      <Section title="System">
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <p className="text-text-muted text-xs mb-0.5">Backend</p>
            <p className="text-text-secondary">Go + chi + SQLite</p>
          </div>
          <div>
            <p className="text-text-muted text-xs mb-0.5">Frontend</p>
            <p className="text-text-secondary">React + Vite + TanStack Query</p>
          </div>
          <div>
            <p className="text-text-muted text-xs mb-0.5">Database</p>
            <p className="text-text-secondary">SQLite (FTS5 + WAL)</p>
          </div>
          <div>
            <p className="text-text-muted text-xs mb-0.5">Version</p>
            <p className="text-text-secondary">0.1.0</p>
          </div>
        </div>
      </Section>
    </div>
  )
}
