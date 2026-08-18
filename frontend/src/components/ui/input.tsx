import * as React from "react"

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  error?: string
  label?: string
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, error, label, ...props }, ref) => {
    return (
      <div className="flex flex-col space-y-1">
        {label && (
          <label className="text-sm font-medium text-text-dim">
            {label}
          </label>
        )}
        <input
          ref={ref}
          className={`px-4 py-2 rounded-[5px] border border-border bg-surface text-text placeholder:text-text-faint focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent transition-all ${
            error ? 'border-danger focus:ring-danger' : ''
          } ${className || ''}`}
          {...props}
        />
        {error && (
          <span className="text-xs text-danger">{error}</span>
        )}
      </div>
    )
  }
)
Input.displayName = "Input"

export { Input }