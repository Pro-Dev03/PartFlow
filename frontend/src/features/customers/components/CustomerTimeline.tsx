import { clsx } from 'clsx'
import type { CustomerTimelineEvent } from '../types/customer'

interface CustomerTimelineProps {
  events: CustomerTimelineEvent[]
  className?: string
}

export function CustomerTimeline({ events, className }: CustomerTimelineProps) {
  const typeIcons = {
    purchase: '💰',
    payment: '💳',
    debt: '📝',
    note: '📋',
    contact: '📞',
  }

  const typeColors = {
    purchase: 'bg-success-10 border-success-30 text-success',
    payment: 'bg-primary-10 border-primary-30 text-primary',
    debt: 'bg-danger-10 border-danger-30 text-danger',
    note: 'bg-muted-10 border-muted-30 text-muted',
    contact: 'bg-info-10 border-info-30 text-info',
  }

  if (events.length === 0) {
    return (
      <div className={clsx('text-center py-8 text-muted', className)}>
        لا يوجد سجل للعميل
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
            {event.amount !== undefined && (
              <p className="text-sm font-medium text-text">
                المبلغ: {event.amount.toFixed(2)}
              </p>
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
