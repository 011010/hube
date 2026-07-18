import { createContext, useContext, useState } from 'react'
import { http, TOKEN_KEY } from '../services/api'

interface AuthContextValue {
  token: string | null
  isAuthenticated: boolean
  login: (token: string) => Promise<boolean>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

// eslint-disable-next-line react-refresh/only-export-components -- hook intentionally co-located with its provider
export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within an AuthProvider')
  return ctx
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem(TOKEN_KEY))

  // Validate the candidate token against a protected endpoint; only persist
  // it when the server accepts it so the gate can show an error otherwise.
  async function login(nextToken: string): Promise<boolean> {
    const trimmed = nextToken.trim()
    if (!trimmed) return false
    try {
      await http.get('/settings', { headers: { Authorization: `Bearer ${trimmed}` } })
      localStorage.setItem(TOKEN_KEY, trimmed)
      setToken(trimmed)
      return true
    } catch {
      return false
    }
  }

  function logout() {
    localStorage.removeItem(TOKEN_KEY)
    setToken(null)
  }

  return (
    <AuthContext.Provider value={{ token, isAuthenticated: token !== null, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}
