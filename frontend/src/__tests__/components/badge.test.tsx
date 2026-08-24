import { describe, it, expect } from 'vitest'
import { render, screen } from '../../test/utils'
import { Badge } from '../../components/ui/badge'

describe('Badge Component', () => {
  it('should render badge with text', () => {
    render(<Badge>Test Badge</Badge>)
    expect(screen.getByText('Test Badge')).toBeInTheDocument()
  })

  it('should render with different variants', () => {
    const { rerender } = render(<Badge variant="success">Success</Badge>)
    expect(screen.getByText('Success')).toBeInTheDocument()
    
    rerender(<Badge variant="danger">Danger</Badge>)
    expect(screen.getByText('Danger')).toBeInTheDocument()
  })
})