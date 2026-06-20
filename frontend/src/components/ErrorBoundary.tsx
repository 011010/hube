import { Component, type ReactNode } from 'react'

interface Props { children: ReactNode }
interface State { error: Error | null }

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  render() {
    if (this.state.error) {
      return (
        <div className="flex flex-col items-center justify-center h-full p-8 text-center">
          <p className="text-red-400 text-sm font-medium mb-2">Something went wrong</p>
          <p className="text-gray-500 text-xs font-mono">{this.state.error.message}</p>
          <button
            onClick={() => this.setState({ error: null })}
            className="mt-4 text-xs text-indigo-400 hover:text-indigo-300"
          >
            Try again
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
