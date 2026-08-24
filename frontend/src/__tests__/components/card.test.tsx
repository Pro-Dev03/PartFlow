import { describe, it, expect } from 'vitest'
import { render, screen } from '../../test/utils'
import { Card, CardContent } from '../../components/ui/card'

describe('Card Component', () => {
  it('should render card with content', () => {
    render(
      <Card>
        <CardContent>Test Content</CardContent>
      </Card>
    )
    
    expect(screen.getByText('Test Content')).toBeInTheDocument()
  })

  it('should render with different variants', () => {
    const { rerender } = render(
      <Card variant="default">
        <CardContent>Default</CardContent>
      </Card>
    )
    expect(screen.getByText('Default')).toBeInTheDocument()
    
    rerender(
      <Card variant="featured">
        <CardContent>Featured</CardContent>
      </Card>
    )
    expect(screen.getByText('Featured')).toBeInTheDocument()
  })
})