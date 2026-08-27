import { describe, it, expect } from 'vitest'
import { render, screen } from '../../test/utils'
import Modal from '../../components/ui/modal'

describe('Modal Component', () => {
  it('should render when isOpen is true', () => {
    render(
      <Modal isOpen={true} onClose={() => {}}>
        <div>Test Content</div>
      </Modal>
    )
    
    expect(screen.getByText('Test Content')).toBeInTheDocument()
  })

  it('should not render when isOpen is false', () => {
    render(
      <Modal isOpen={false} onClose={() => {}}>
        <div>Test Content</div>
      </Modal>
    )
    
    expect(screen.queryByText('Test Content')).not.toBeInTheDocument()
  })
})