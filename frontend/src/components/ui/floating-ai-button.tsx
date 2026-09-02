import { useRef, useState } from 'react'
import NeonAIBot from './neon-ai-bot'

interface FloatingAIButtonProps {
  onClick: (position: { x: number; y: number }) => void
}

const DRAG_THRESHOLD = 10

export default function FloatingAIButton({ onClick }: FloatingAIButtonProps) {
  const [position, setPosition] = useState({ x: -1, y: -1 })
  const [isDragging, setIsDragging] = useState(false)
  const [isHovered, setIsHovered] = useState(false)
  const pointerIdRef = useRef<number | null>(null)
  const startPointRef = useRef({ x: 0, y: 0 })
  const dragOffsetRef = useRef({ x: 0, y: 0 })
  const dragStartedRef = useRef(false)
  const suppressClickRef = useRef(false)

  const resetDragState = () => {
    pointerIdRef.current = null
    dragStartedRef.current = false
    setIsDragging(false)
  }

  const handlePointerDown = (event: React.PointerEvent<HTMLButtonElement>) => {
    pointerIdRef.current = event.pointerId
    dragStartedRef.current = false
    suppressClickRef.current = false
    setIsDragging(false)

    startPointRef.current = { x: event.clientX, y: event.clientY }
    const rect = event.currentTarget.getBoundingClientRect()
    dragOffsetRef.current = {
      x: event.clientX - rect.left,
      y: event.clientY - rect.top,
    }

    // Convert from bottom-left positioning to left-top for dragging
    if (isDefaultPosition) {
      setPosition({
        x: 20,
        y: window.innerHeight - 80 - 20,
      })
    }

    if (event.currentTarget.setPointerCapture) {
      event.currentTarget.setPointerCapture(event.pointerId)
    }
  }

  const handlePointerMove = (event: React.PointerEvent<HTMLButtonElement>) => {
    if (pointerIdRef.current !== event.pointerId) return

    const distance = Math.hypot(
      event.clientX - startPointRef.current.x,
      event.clientY - startPointRef.current.y,
    )

    if (distance > DRAG_THRESHOLD || dragStartedRef.current) {
      dragStartedRef.current = true
      setIsDragging(true)
      event.preventDefault()

      const nextX = event.clientX - dragOffsetRef.current.x
      const nextY = event.clientY - dragOffsetRef.current.y

      setPosition({
        x: Math.max(0, Math.min(nextX, window.innerWidth - 80)),
        y: Math.max(0, Math.min(nextY, window.innerHeight - 80)),
      })
    }
  }

  const handlePointerUp = (event: React.PointerEvent<HTMLButtonElement>) => {
    if (pointerIdRef.current !== event.pointerId) return

    if (dragStartedRef.current) {
      suppressClickRef.current = true
    }

    if (event.currentTarget.hasPointerCapture && event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId)
    }

    resetDragState()
  }

  const handlePointerLeave = (event: React.PointerEvent<HTMLButtonElement>) => {
    if (pointerIdRef.current !== event.pointerId) return

    if (dragStartedRef.current) {
      suppressClickRef.current = true
    }

    resetDragState()
  }

  const handleClick = () => {
    if (suppressClickRef.current) {
      suppressClickRef.current = false
      return
    }

    onClick(position)
  }

  const isDefaultPosition = position.x === -1 && position.y === -1

  return (
    <button
      type="button"
      aria-label="فتح مساعد PartFlow"
      className="fixed border-0 bg-transparent p-0 outline-none"
      style={
        isDefaultPosition
          ? {
              left: 20,
              bottom: 20,
              zIndex: 2147483647,
              cursor: 'pointer',
              touchAction: 'none',
              userSelect: 'none',
              pointerEvents: 'auto',
            }
          : {
              left: position.x,
              top: position.y,
              zIndex: 2147483647,
              cursor: isDragging ? 'grabbing' : 'pointer',
              touchAction: 'none',
              userSelect: 'none',
              pointerEvents: 'auto',
            }
      }
      onPointerDown={handlePointerDown}
      onPointerMove={handlePointerMove}
      onPointerUp={handlePointerUp}
      onPointerLeave={(e) => {
        setIsHovered(false)
        handlePointerLeave(e)
      }}
      onPointerCancel={handlePointerLeave}
      onPointerEnter={() => setIsHovered(true)}
      onClick={handleClick}
    >
      <div className="relative bot-hover-effect" style={{ width: '80px', height: '80px' }}>
        <NeonAIBot size={80} />
      </div>
      <style>{`
        .bot-hover-effect:hover .sparkle-container {
          opacity: 1 !important;
        }
      `}</style>
    </button>
  )
}
