import { useState } from 'react'
import { useAuth } from '../../contexts/AuthContext'
import { HubeLogo } from '../atoms/HubeLogo'
import { Input } from '../atoms/Input'
import { Button } from '../atoms/Button'

export function AuthGate() {
  const { login } = useAuth()
  const [token, setToken] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!token.trim() || loading) return
    setLoading(true)
    setError(null)
    const ok = await login(token)
    setLoading(false)
    if (!ok) setError('Invalid token or server unreachable')
  }

  return (
    <div className="flex h-screen items-center justify-center bg-surface-base px-4">
      <div className="w-full max-w-sm bg-surface-elevated border border-border rounded-2xl p-8">
        <div className="flex flex-col items-center mb-8">
          <HubeLogo size={40} />
          <h1 className="text-xl font-semibold tracking-tight text-text-primary mt-4">hube</h1>
          <p className="text-sm text-text-muted mt-1 text-center">
            Enter your access token to continue
          </p>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            type="password"
            value={token}
            onChange={e => setToken(e.target.value)}
            placeholder="Access token"
            autoFocus
            error={error ?? undefined}
          />
          <Button
            type="submit"
            disabled={!token.trim() || loading}
            className="w-full justify-center"
          >
            {loading ? 'Verifying…' : 'Sign in'}
          </Button>
        </form>
      </div>
    </div>
  )
}
