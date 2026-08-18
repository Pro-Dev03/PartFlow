import * as React from "react"

interface TextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  error?: string
  label?: string
}

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, error, label, ...props }, ref) => {
    return (
      <div className="flex flex-col space-y-1">
        {label && (
          <label className="text-sm font-medium text-ink-soft dark:text-gray-300">
            {label}
          </label>
        )}
        <textarea
          ref={ref}
          className={`px-4 py-2 rounded-[8px] border border-rule bg-white text-ink placeholder:text-ink-muted focus:outline-none focus:ring-2 focus:ring-seal focus:border-transparent transition-all resize-y dark:bg-paper-dark dark:border-rule-dark dark:text-white dark:placeholder:text-gray-500 dark:focus:ring-seal-dark ${
            error ? 'border-debt focus:ring-debt' : ''
          } ${className || ''}`}
          {...props}
        />
        {error && (
          <span className="text-xs text-debt">{error}</span>
        )}
      </div>
    )
  }
)
Textarea.displayName = "Textarea"

export { Textarea }