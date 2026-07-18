import { createContext, useContext, useEffect, useState } from 'react'

export type ThemeMode = 'dark' | 'light' | 'system'

export type ThemeName =
  | 'solaris'
  | 'apple-inspired'
  | 'minimal-tech'
  | 'glassmorphism'
  | 'neobrutalism'
  | 'synthwave'
  | 'cyberpunk'
  | 'solarpunk'
  | 'frutiger-aero'
  | 'japandi'
  | 'dark-academia'
  | 'luxury-black-gold'
  | 'neo-tokyo'
  | 'tulum-boho'
  | 'retro-terminal'
  | 'space-opera'

interface ThemeContextValue {
  mode: ThemeMode
  theme: ThemeName
  setMode: (m: ThemeMode) => void
  setTheme: (t: ThemeName) => void
}

const ThemeContext = createContext<ThemeContextValue>({
  mode: 'dark',
  theme: 'solaris',
  setMode: () => {},
  setTheme: () => {},
})

// eslint-disable-next-line react-refresh/only-export-components -- hook intentionally co-located with its provider
export function useTheme() {
  return useContext(ThemeContext)
}

const ALL_THEMES: ThemeName[] = [
  'solaris', 'apple-inspired', 'minimal-tech', 'glassmorphism',
  'neobrutalism', 'synthwave', 'cyberpunk', 'solarpunk',
  'frutiger-aero', 'japandi', 'dark-academia', 'luxury-black-gold',
  'neo-tokyo', 'tulum-boho', 'retro-terminal', 'space-opera',
]

function isThemeName(v: string): v is ThemeName {
  return (ALL_THEMES as readonly string[]).includes(v)
}

function resolvedDark(mode: ThemeMode): boolean {
  if (mode === 'dark') return true
  if (mode === 'light') return false
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

function applyTheme(mode: ThemeMode, theme: ThemeName) {
  const html = document.documentElement
  const dark = resolvedDark(mode)

  html.classList.toggle('dark', dark)

  // Remove all theme classes, add the new one
  for (const t of ALL_THEMES) {
    html.classList.remove(`theme-${t}`)
  }
  html.classList.add(`theme-${theme}`)
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [mode, setModeState] = useState<ThemeMode>(() => {
    const v = localStorage.getItem('theme-mode')
    return (v === 'dark' || v === 'light' || v === 'system') ? v : 'dark'
  })
  const [theme, setThemeState] = useState<ThemeName>(() => {
    const raw = localStorage.getItem('theme-name')
    if (raw && isThemeName(raw)) return raw
    return 'solaris'
  })

  useEffect(() => {
    applyTheme(mode, theme)
  }, [mode, theme])

  // React to system preference changes when mode === 'system'
  useEffect(() => {
    if (mode !== 'system') return
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const handler = () => applyTheme('system', theme)
    mq.addEventListener('change', handler)
    return () => mq.removeEventListener('change', handler)
  }, [mode, theme])

  function setMode(m: ThemeMode) {
    localStorage.setItem('theme-mode', m)
    setModeState(m)
  }

  function setTheme(t: ThemeName) {
    localStorage.setItem('theme-name', t)
    setThemeState(t)
  }

  return (
    <ThemeContext.Provider value={{ mode, theme, setMode, setTheme }}>
      {children}
    </ThemeContext.Provider>
  )
}
