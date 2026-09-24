import { useState, type ReactNode } from 'react'
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { PaymentMethod } from '../../../features/sales/types/pos.types'
import { AdvancedPaymentPanel } from '../../../features/sales/components/modern/AdvancedPaymentPanel'

vi.mock('../../../design-system/components/modal', () => ({
  Modal: ({ isOpen, children }: { isOpen: boolean; children: ReactNode }) =>
    isOpen ? <div role="dialog">{children}</div> : null,
}))

function PaymentPanelHarness({
  total = 150,
  checkoutBlockedReason,
  onCheckout,
}: {
  total?: number
  checkoutBlockedReason?: string
  onCheckout: () => void
}) {
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('cash')
  const [paidAmount, setPaidAmount] = useState('')
  const [installmentMonths, setInstallmentMonths] = useState(3)

  return (
    <AdvancedPaymentPanel
      paymentMethod={paymentMethod}
      setPaymentMethod={setPaymentMethod}
      paidAmount={paidAmount}
      setPaidAmount={setPaidAmount}
      total={total}
      isProcessing={false}
      onCheckout={onCheckout}
      checkoutBlockedReason={checkoutBlockedReason}
      installmentMonths={installmentMonths}
      setInstallmentMonths={setInstallmentMonths}
    />
  )
}

function renderPaymentPanel(props: { total?: number; checkoutBlockedReason?: string } = {}) {
  const onCheckout = vi.fn()
  render(<PaymentPanelHarness {...props} onCheckout={onCheckout} />)
  return { onCheckout }
}

function keypadButton(name: string) {
  return screen.getByRole('button', { name })
}

async function paymentInput() {
  const input = screen.getByRole('textbox', { name: 'المبلغ المدفوع' })
  await waitFor(() => expect(input).toHaveValue('150.00'))
  return input
}

describe('AdvancedPaymentPanel numeric keypad', () => {
  beforeEach(() => vi.clearAllMocks())

  it('replaces the default cash amount with keypad digits and enables exact checkout', async () => {
    renderPaymentPanel()
    const input = await paymentInput()

    fireEvent.click(keypadButton('مسح المبلغ'))
    for (const digit of ['1', '5', '0']) fireEvent.click(keypadButton(digit))

    expect(input).toHaveValue('150')
    expect(screen.getByRole('button', { name: /إتمام البيع/ })).toBeEnabled()
  })

  it('shows the customer change when cash received exceeds the invoice total', async () => {
    renderPaymentPanel()
    await paymentInput()

    fireEvent.click(keypadButton('مسح المبلغ'))
    for (const digit of ['2', '0', '0']) fireEvent.click(keypadButton(digit))

    expect(screen.getByText(/الباقي للعميل \(المردود\): ₪50/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /إتمام البيع/ })).toBeEnabled()
  })

  it('blocks checkout and reports the remaining amount when cash payment is short', async () => {
    renderPaymentPanel()
    await paymentInput()

    fireEvent.click(keypadButton('مسح المبلغ'))
    for (const digit of ['1', '0', '0']) fireEvent.click(keypadButton(digit))

    expect(screen.getByText('المتبقي للدفع: ₪50')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /إتمام البيع/ })).toBeDisabled()
  })

  it('supports 00, decimal entry, backspace, and clearing', async () => {
    renderPaymentPanel()
    const input = await paymentInput()

    fireEvent.click(keypadButton('مسح المبلغ'))
    fireEvent.click(keypadButton('00'))
    expect(input).toHaveValue('0')
    fireEvent.click(keypadButton('5'))
    fireEvent.click(keypadButton('حذف آخر رقم'))
    expect(input).toHaveValue('')

    for (const key of ['1', '2', 'نقطة عشرية', '5']) fireEvent.click(keypadButton(key))
    expect(input).toHaveValue('12.5')
    fireEvent.click(keypadButton('مسح المبلغ'))
    expect(input).toHaveValue('')
  })

  it('sets the exact invoice amount and submits only through the checkout action', async () => {
    const { onCheckout } = renderPaymentPanel()
    const input = await paymentInput()

    fireEvent.click(keypadButton('مسح المبلغ'))
    fireEvent.click(keypadButton('دفع المبلغ كاملًا'))
    expect(input).toHaveValue('150.00')
    fireEvent.click(screen.getByRole('button', { name: /إتمام البيع/ }))
    expect(onCheckout).toHaveBeenCalledTimes(1)
  })

  it('routes physical numpad keys to the active amount without changing barcode/search input', async () => {
    renderPaymentPanel()
    const input = await paymentInput()

    fireEvent.click(keypadButton('مسح المبلغ'))
    const oneKey = keypadButton('1')
    fireEvent.keyDown(oneKey, { key: '1', code: 'Numpad1' })
    fireEvent.keyDown(oneKey, { key: '5', code: 'Numpad5' })
    fireEvent.keyDown(oneKey, { key: '0', code: 'Numpad0' })

    expect(input).toHaveValue('150')
  })

  it('edits the selected split-payment amount with the same keypad', async () => {
    renderPaymentPanel()
    await paymentInput()

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    fireEvent.click(keypadButton('تقسيم الدفع'))
    const splitDialog = screen.getByRole('dialog')
    const splitKeypadButton = (name: string) => within(splitDialog).getByRole('button', { name })

    expect(splitDialog).toBeInTheDocument()
    const firstSplitInput = within(splitDialog).getByRole('textbox', { name: 'مبلغ طريقة الدفع 1' })
    fireEvent.focus(firstSplitInput)
    await waitFor(() => expect(firstSplitInput).toHaveValue('150.00'))

    fireEvent.click(splitKeypadButton('مسح المبلغ'))
    for (const digit of ['1', '0', '0']) fireEvent.click(splitKeypadButton(digit))
    expect(firstSplitInput).toHaveValue('100')
    expect(within(splitDialog).getByText('المبلغ المتبقي').parentElement).toHaveTextContent('₪50')

    fireEvent.click(within(splitDialog).getByRole('button', { name: 'إضافة طريقة دفع' }))
    const secondSplitInput = within(splitDialog).getByRole('textbox', { name: 'مبلغ طريقة الدفع 2' })
    expect(secondSplitInput).toHaveValue('50.00')
    fireEvent.click(splitKeypadButton('مسح المبلغ'))
    for (const digit of ['5', '0']) fireEvent.click(splitKeypadButton(digit))
    expect(within(splitDialog).getByText('المبلغ الموزع').parentElement).toHaveTextContent('₪150')
    fireEvent.click(within(splitDialog).getByRole('button', { name: 'تأكيد التقسيم' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(screen.getByLabelText('ملخص تقسيم الدفع')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /إتمام البيع/ })).toBeEnabled()
  })

  it('keeps the keypad out of payment methods that do not use a cash amount field', async () => {
    renderPaymentPanel()
    await paymentInput()

    fireEvent.click(keypadButton('تغيير الطريقة'))
    fireEvent.click(keypadButton('بطاقة'))

    expect(screen.queryByRole('textbox', { name: 'المبلغ المدفوع' })).not.toBeInTheDocument()
    expect(screen.queryByLabelText('لوحة الأرقام للمبلغ المدفوع')).not.toBeInTheDocument()
  })

  it('leaves shift validation in control of checkout availability', async () => {
    renderPaymentPanel({ checkoutBlockedReason: 'افتح الوردية أولًا قبل إتمام البيع.' })
    await paymentInput()

    expect(screen.getByRole('button', { name: /إتمام البيع/ })).toBeDisabled()
    expect(screen.getByText('افتح الوردية أولًا قبل إتمام البيع.')).toBeInTheDocument()
  })
})
