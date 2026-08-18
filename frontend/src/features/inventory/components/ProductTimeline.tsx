import { clsx } from 'clsx'

interface TimelineEvent {
  id: string
  type: 'purchase' | 'sale' | 'inspection' | 'price_change' | 'stock_adjustment' | 'transfer'
  date: string
  description: string
  details?: string
  user?: string
}

interface ProductTimelineProps {
  events: TimelineEvent[]
  className?: string
}

export function ProductTimeline({ events, className }: ProductTimelineProps) {
  const typeIcons = {
    purchase: '📦',
    sale: '💰',
    inspection: '🔍',
    price_change: '💲',
    stock_adjustment: '📊',
    transfer: '🔄',
  }

  const typeColors = {
    purchase: 'bg-success-10 border-success-30 text-success',
    sale: 'bg-primary-10 border-primary-30 text-primary',
    inspection: 'bg-warning-10 border-warning-30 text-warning',
    price_change: 'bg-info-10 border-info-30 text-info',
    stock_adjustment: 'bg-muted-10 border-muted-30 text-muted',
    transfer: 'bg-purple-10 border-purple-30 text-purple',
  }

  if (events.length === 0) {
    return (
      <div className={clsx('text-center py-8 text-muted', className)}>
        لا يوجد سجل للمنتج
      </div>
    )
  }

  return (
    <div className={clsx('space-y-4', className)}>
      {events.map((event, index) => (
        <div key={event.id} className="flex gap-4">
          <div className="flex flex-col items-center">
            <div
              className={clsx(
                'w-10 h-10 rounded-full flex items-center justify-center text-lg border-2',
                typeColors[event.type]
              )}
            >
              {typeIcons[event.type]}
            </div>
            {index < events.length - 1 && (
              <div className="w-0.5 h-full bg-border mt-2" />
            )}
          </div>
          <div className="flex-1 pb-4">
            <div className="flex items-center justify-between mb-1">
              <span className="font-medium text-text">{event.description}</span>
              <span className="text-sm text-muted">{event.date}</span>
            </div>
            {event.details && (
              <p className="text-sm text-muted mb-1">{event.details}</p>
            )}
            {event.user && (
              <p className="text-xs text-muted">بواسطة: {event.user}</p>
            )}
          </div>
        </div>
      ))}
    </div>
  )
}
