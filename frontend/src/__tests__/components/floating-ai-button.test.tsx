import { describe, it, expect, vi } from 'vitest'
import { fireEvent, render, screen } from '../../test/utils'
import FloatingAIButton from '../../components/ui/floating-ai-button'

describe('FloatingAIButton', () => {
  it('calls onClick when clicked without dragging', () => {
    const onClick = vi.fn()
    render(<FloatingAIButton onClick={onClick} />)

    const button = screen.getByRole('button', { name: /فتح مساعد partflow/i })
    fireEvent.click(button)

    expect(onClick).toHaveBeenCalledTimes(1)
  })

})
