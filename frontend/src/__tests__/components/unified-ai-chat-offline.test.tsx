import { describe, it, expect } from 'vitest'
import { fireEvent, render, screen } from '../../test/utils'
import UnifiedAIChat from '../../components/ui/unified-ai-chat'
import { generateAssistantReply } from '../../lib/assistant-response'

describe('UnifiedAIChat offline fallback', () => {
  it('uses local store-aware fallback responses when offline', async () => {
    render(
      <UnifiedAIChat
        isOffline
        assistantContext={{
          lowStockCount: 2,
          overdueDebtsCount: 1,
          salesToday: 1500,
          salesYesterday: 1200,
          lowStockItems: [
            { name: 'بطارية', quantity: 3, minQuantity: 5 },
            { name: 'زيت', quantity: 2, minQuantity: 4 },
          ],
          overdueDebts: [
            { customerName: 'أحمد', amount: 250, days: 7 },
          ],
        }}
      />,
    )

    const input = screen.getByPlaceholderText('اكتب رسالتك...')
    fireEvent.change(input, { target: { value: 'ما الذي يحدث الآن؟' } })
    fireEvent.click(screen.getByRole('button', { name: 'إرسال' }))

    expect(await screen.findByText(/أولويات اليوم واضحة/i)).toBeInTheDocument()
  })

  it('uses product names from the backend payload when low-stock alerts are loaded', () => {
    const response = generateAssistantReply('شو عم يصير؟', {
      lowStockCount: 2,
      overdueDebtsCount: 0,
      salesToday: 0,
      salesYesterday: 0,
      lowStockItems: [
        { product_name: 'بطارية', quantity: 2, minQuantity: 5 },
        { product_name: 'زيت', quantity: 1, minQuantity: 4 },
      ],
      overdueDebts: [],
    })

    expect(response).toContain('بطارية')
    expect(response).toContain('زيت')
  })

  it('introduces itself like a human store assistant when asked who it is', async () => {
    render(
      <UnifiedAIChat
        isOffline
        assistantContext={{
          lowStockCount: 0,
          overdueDebtsCount: 0,
          salesToday: 0,
          salesYesterday: 0,
          lowStockItems: [],
          overdueDebts: [],
        }}
      />,
    )

    const input = screen.getByPlaceholderText('اكتب رسالتك...')
    fireEvent.change(input, { target: { value: 'من أنت؟' } })
    fireEvent.click(screen.getByRole('button', { name: 'إرسال' }))

    expect(await screen.findByText(/أنا مساعدك الشخصي|أنا مساعدك في PartFlow/i)).toBeInTheDocument()
  })

  it('answers in a local Palestinian Arabic tone', async () => {
    render(
      <UnifiedAIChat
        isOffline
        assistantContext={{
          lowStockCount: 1,
          overdueDebtsCount: 0,
          salesToday: 300,
          salesYesterday: 200,
          lowStockItems: [{ name: 'بطارية', quantity: 2, minQuantity: 5 }],
          overdueDebts: [],
        }}
      />,
    )

    const input = screen.getByPlaceholderText('اكتب رسالتك...')
    fireEvent.change(input, { target: { value: 'شو عم يصير؟' } })
    fireEvent.click(screen.getByRole('button', { name: 'إرسال' }))

    const matches = await screen.findAllByText(/أولويات اليوم|محتاج|شو عم/i)
    expect(matches.length).toBeGreaterThan(0)
  })

  it('responds politely when asked how the assistant is doing', async () => {
    render(
      <UnifiedAIChat
        isOffline
        assistantContext={{
          lowStockCount: 0,
          overdueDebtsCount: 0,
          salesToday: 0,
          salesYesterday: 0,
          lowStockItems: [],
          overdueDebts: [],
        }}
      />,
    )

    const input = screen.getByPlaceholderText('اكتب رسالتك...')
    fireEvent.change(input, { target: { value: 'كيف حالك؟' } })
    fireEvent.click(screen.getByRole('button', { name: 'إرسال' }))

    expect(await screen.findByText(/أنا بخير|جاهز لمساعدتك|أستطيع متابعة المخزون/i)).toBeInTheDocument()
  })

  it('switches to an urgent tone for urgent store issues', async () => {
    render(
      <UnifiedAIChat
        isOffline
        assistantContext={{
          lowStockCount: 3,
          overdueDebtsCount: 2,
          salesToday: 400,
          salesYesterday: 500,
          lowStockItems: [
            { name: 'بطارية', quantity: 1, minQuantity: 5 },
            { name: 'زيت', quantity: 2, minQuantity: 6 },
          ],
          overdueDebts: [{ customerName: 'سامي', amount: 600, days: 12 }],
        }}
      />,
    )

    const input = screen.getByPlaceholderText('اكتب رسالتك...')
    fireEvent.change(input, { target: { value: 'مستعجل، شو لازم أعمل؟' } })
    fireEvent.click(screen.getByRole('button', { name: 'إرسال' }))

    expect(await screen.findByText(/مستعجل|عاجل|حالاً|طوارئ/i)).toBeInTheDocument()
  })

  it('shows the assistant name and connection status in the chat header', () => {
    render(
      <UnifiedAIChat
        isOffline
        assistantContext={{
          lowStockCount: 0,
          overdueDebtsCount: 0,
          salesToday: 0,
          salesYesterday: 0,
          lowStockItems: [],
          overdueDebts: [],
        }}
      />,
    )

    expect(screen.getByText('أمان')).toBeInTheDocument()
    expect(screen.getByText(/أوفلاين|غير متصل/i)).toBeInTheDocument()
  })
})
