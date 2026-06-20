import { createContext, useContext, useEffect, useState } from 'react'

export type ThemeMode = 'dark' | 'light' | 'system'
export type AccentColor = 'indigo' | 'violet' | 'rose' | 'emerald' | 'sky' | 'orange'

interface ThemeContextValue {
  mode: ThemeMode
  accent: AccentColor
  setMode: (m: ThemeMode) => void
  setAccent: (a: AccentColor) => void
}

const ThemeContext = createContext<ThemeContextValue>({
  mode: 'dark',
  accent: 'indigo',
  setMode: () => {},
  setAccent: () => {},
})

export function useTheme() {
  return useContext(ThemeContext)
}

const ACCENT_CLASS: Record<AccentColor, string> = {
  indigo: '',
  violet: 'theme-violet',
  rose: 'theme-rose',
  emerald: 'theme-emerald',
  sky: 'theme-sky',
  orange: 'theme-orange',
}

function resolvedDark(mode: ThemeMode): boolean {
  if (mode === 'dark') return true
  if (mode === 'light') return false
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

function applyTheme(mode: ThemeMode, accent: AccentColor) {
  const html = document.documentElement
  const dark = resolvedDark(mode)

  html.classList.toggle('dark', dark)

  // Remove all accent classes, add the new one
  Object.values(ACCENT_CLASS).forEach(c => c && html.classList.remove(c))
  const cls = ACCENT_CLASS[accent]
  if (cls) html.classList.add(cls)
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [mode, setModeState] = useState<ThemeMode>(() => {
    const v = localStorage.getItem('theme-mode')
    return (v === 'dark' || v === 'light' || v === 'system') ? v : 'dark'
  })
  const [accent, setAccentState] = useState<AccentColor>(() => {
    const v = localStorage.getItem('theme-accent')
    return (v === 'indigo' || v === 'violet' || v === 'rose' || v === 'emerald' || v === 'sky' || v === 'orange') ? v : 'indigo'
  })

  useEffect(() => {
    applyTheme(mode, accent)
  }, [mode, accent])

  // React to system preference changes when mode === 'system'
  useEffect(() => {
    if (mode !== 'system') return
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const handler = () => applyTheme('system', accent)
    mq.addEventListener('change', handler)
    return () => mq.removeEventListener('change', handler)
  }, [mode, accent])

  function setMode(m: ThemeMode) {
    localStorage.setItem('theme-mode', m)
    setModeState(m)
  }

  function setAccent(a: AccentColor) {
    localStorage.setItem('theme-accent', a)
    setAccentState(a)
  }

  return (
    <ThemeContext.Provider value={{ mode, accent, setMode, setAccent }}>
      {children}
    </ThemeContext.Provider>
  )
}
