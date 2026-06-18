import { useCardTrackerSummary } from '../../hooks/useCardTracker'

function fmt(amount: number) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(amount)
}

function daysUntil(dateStr: string) {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const target = new Date(dateStr + 'T00:00:00')
  return Math.ceil((target.getTime() - today.getTime()) / (1000 * 60 * 60 * 24))
}

export function CardTrackerWidget() {
  const { data, isLoading, isError } = useCardTrackerSummary()

  if (isLoading) return <Skeleton />

  if (isError || !data || data.configured === false) {
    return (
      <div className="bg-gray-900 border border-gray-800 rounded-xl px-5 py-4 text-sm text-gray-500">
        PayPinga not connected.{' '}
        <span className="text-gray-600">Set PAYPINGA_URL and PAYPINGA_KEY to enable.</span>
      </div>
    )
  }

  const activeCards = data.cards.filter(c => c.estimated_payment > 0)

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="bg-gray-900 border border-gray-800 rounded-xl px-5 py-4">
          <p className="text-xs text-gray-500 mb-1">Total debt</p>
          <p className="text-2xl font-semibold tabular-nums text-red-400">{fmt(data.total_debt)}</p>
        </div>
        <div className="bg-gray-900 border border-gray-800 rounded-xl px-5 py-4">
          <p className="text-xs text-gray-500 mb-1">This month's payment</p>
          <p className="text-2xl font-semibold tabular-nums text-amber-400">{fmt(data.total_payment)}</p>
          {data.next_pay_date && (
            <p className="text-xs text-gray-600 mt-1">
              Next due in {daysUntil(data.next_pay_date)}d · {data.next_pay_date}
            </p>
          )}
        </div>
      </div>

      {activeCards.length > 0 && (
        <div className="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
          <div className="px-5 py-3 border-b border-gray-800">
            <span className="text-xs font-medium text-gray-500 uppercase tracking-wider">Cards</span>
          </div>
          <ul className="divide-y divide-gray-800">
            {data.cards.map(card => {
              const days = daysUntil(card.pay_date)
              const urgent = days <= 3
              return (
                <li key={card.id} className="flex items-center justify-between px-5 py-3">
                  <div>
                    <p className="text-sm text-gray-200">
                      {card.alias}
                      {card.last4 && <span className="text-gray-500 ml-1">···{card.last4}</span>}
                    </p>
                    <p className={`text-xs mt-0.5 ${urgent ? 'text-red-400' : 'text-gray-500'}`}>
                      {card.bank} · pays {card.pay_date} ({days}d)
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-medium tabular-nums text-amber-400">{fmt(card.estimated_payment)}</p>
                    <p className="text-xs text-gray-600 tabular-nums">{fmt(card.balance)} remaining</p>
                  </div>
                </li>
              )
            })}
          </ul>
          <div className="px-5 py-3 border-t border-gray-800">
            <a
              href="https://paypinga.vercel.app"
              target="_blank"
              rel="noreferrer"
              className="text-xs text-indigo-400 hover:text-indigo-300 transition-colors"
            >
              Open PayPinga →
            </a>
          </div>
        </div>
      )}
    </div>
  )
}

function Skeleton() {
  return (
    <div className="grid grid-cols-2 gap-4 animate-pulse">
      {[0, 1].map(i => (
        <div key={i} className="bg-gray-900 border border-gray-800 rounded-xl px-5 py-4 h-20" />
      ))}
    </div>
  )
}
