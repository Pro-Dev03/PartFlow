import { useState } from 'react'
import { clsx } from 'clsx'

type BulkAction = 'delete' | 'update_price' | 'update_stock' | 'change_category' | 'export'

interface BulkActionsProps {
  selectedCount: number
  onAction: (action: BulkAction, data?: any) => void
  onClearSelection: () => void
  disabled?: boolean
  className?: string
}

export function BulkActions({
  selectedCount,
  onAction,
  onClearSelection,
  disabled = false,
  className,
}: BulkActionsProps) {
  const [showMenu, setShowMenu] = useState(false)
  const [activeAction, setActiveAction] = useState<BulkAction | null>(null)

  const handleAction = (action: BulkAction) => {
    setActiveAction(action)
    setShowMenu(false)
    
    // For simple actions, execute directly
    if (action === 'delete' || action === 'export') {
      onAction(action)
      setActiveAction(null)
    }
  }

  const handleConfirmAction = (data?: any) => {
    if (activeAction) {
      onAction(activeAction, data)
      setActiveAction(null)
    }
  }

  if (selectedCount === 0) {
    return null
  }

  return (
    <div className={clsx('bg-primary-10 border border-primary-30 rounded-lg p-4', className)}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="font-medium text-primary">
            {selectedCount} عنصر محدد
          </span>
          <button
            onClick={onClearSelection}
            className="text-sm text-muted hover:text-text underline"
          >
            إلغاء التحديد
          </button>
        </div>

        <div className="relative">
          <button
            onClick={() => setShowMenu(!showMenu)}
            disabled={disabled}
            className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            إجراءات جماعية
          </button>

          {showMenu && (
            <div className="absolute left-0 mt-2 w-56 bg-surface border border-border rounded-lg shadow-lg z-10">
              <div className="py-1">
                <button
                  onClick={() => handleAction('update_price')}
                  className="w-full px-4 py-2 text-right hover:bg-muted-10 transition-colors flex items-center gap-2"
                >
                  <span>💲</span>
                  <span>تحديث الأسعار</span>
                </button>
                <button
                  onClick={() => handleAction('update_stock')}
                  className="w-full px-4 py-2 text-right hover:bg-muted-10 transition-colors flex items-center gap-2"
                >
                  <span>📊</span>
                  <span>تحديث المخزون</span>
                </button>
                <button
                  onClick={() => handleAction('change_category')}
                  className="w-full px-4 py-2 text-right hover:bg-muted-10 transition-colors flex items-center gap-2"
                >
                  <span>📁</span>
                  <span>تغيير التصنيف</span>
                </button>
                <button
                  onClick={() => handleAction('export')}
                  className="w-full px-4 py-2 text-right hover:bg-muted-10 transition-colors flex items-center gap-2"
                >
                  <span>📤</span>
                  <span>تصدير</span>
                </button>
                <hr className="my-1 border-border" />
                <button
                  onClick={() => handleAction('delete')}
                  className="w-full px-4 py-2 text-right hover:bg-danger-10 text-danger transition-colors flex items-center gap-2"
                >
                  <span>🗑️</span>
                  <span>حذف</span>
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Action Modals */}
      {activeAction === 'update_price' && (
        <UpdatePriceModal
          onSubmit={handleConfirmAction}
          onCancel={() => setActiveAction(null)}
        />
      )}

      {activeAction === 'update_stock' && (
        <UpdateStockModal
          onSubmit={handleConfirmAction}
          onCancel={() => setActiveAction(null)}
        />
      )}

      {activeAction === 'change_category' && (
        <ChangeCategoryModal
          onSubmit={handleConfirmAction}
          onCancel={() => setActiveAction(null)}
        />
      )}
    </div>
  )
}

function UpdatePriceModal({ onSubmit, onCancel }: { onSubmit: (data: any) => void; onCancel: () => void }) {
  const [type, setType] = useState<'fixed' | 'percentage'>('fixed')
  const [value, setValue] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit({ type, value: parseFloat(value) })
  }

  return (
    <div className="mt-4 p-4 bg-surface border border-border rounded-lg">
      <h3 className="font-semibold text-text mb-4">تحديث الأسعار</h3>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="flex gap-4">
          <label className="flex items-center gap-2">
            <input
              type="radio"
              value="fixed"
              checked={type === 'fixed'}
              onChange={(e) => setType(e.target.value as any)}
              className="w-4 h-4 text-primary"
            />
            <span>قيمة ثابتة</span>
          </label>
          <label className="flex items-center gap-2">
            <input
              type="radio"
              value="percentage"
              checked={type === 'percentage'}
              onChange={(e) => setType(e.target.value as any)}
              className="w-4 h-4 text-primary"
            />
            <span>نسبة مئوية</span>
          </label>
        </div>
        <input
          type="number"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder={type === 'fixed' ? 'القيمة' : 'النسبة'}
          className="w-full px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          required
        />
        <div className="flex gap-2">
          <button
            type="submit"
            className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
          >
            تطبيق
          </button>
          <button
            type="button"
            onClick={onCancel}
            className="px-4 py-2 border border-border rounded-lg hover:bg-surface-80 transition-colors"
          >
            إلغاء
          </button>
        </div>
      </form>
    </div>
  )
}

function UpdateStockModal({ onSubmit, onCancel }: { onSubmit: (data: any) => void; onCancel: () => void }) {
  const [action, setAction] = useState<'add' | 'subtract' | 'set'>('add')
  const [value, setValue] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit({ action, value: parseInt(value) })
  }

  return (
    <div className="mt-4 p-4 bg-surface border border-border rounded-lg">
      <h3 className="font-semibold text-text mb-4">تحديث المخزون</h3>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="flex gap-4">
          <label className="flex items-center gap-2">
            <input
              type="radio"
              value="add"
              checked={action === 'add'}
              onChange={(e) => setAction(e.target.value as any)}
              className="w-4 h-4 text-primary"
            />
            <span>إضافة</span>
          </label>
          <label className="flex items-center gap-2">
            <input
              type="radio"
              value="subtract"
              checked={action === 'subtract'}
              onChange={(e) => setAction(e.target.value as any)}
              className="w-4 h-4 text-primary"
            />
            <span>طرح</span>
          </label>
          <label className="flex items-center gap-2">
            <input
              type="radio"
              value="set"
              checked={action === 'set'}
              onChange={(e) => setAction(e.target.value as any)}
              className="w-4 h-4 text-primary"
            />
            <span>تعيين</span>
          </label>
        </div>
        <input
          type="number"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder="الكمية"
          className="w-full px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          required
        />
        <div className="flex gap-2">
          <button
            type="submit"
            className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
          >
            تطبيق
          </button>
          <button
            type="button"
            onClick={onCancel}
            className="px-4 py-2 border border-border rounded-lg hover:bg-surface-80 transition-colors"
          >
            إلغاء
          </button>
        </div>
      </form>
    </div>
  )
}

function ChangeCategoryModal({ onSubmit, onCancel }: { onSubmit: (data: any) => void; onCancel: () => void }) {
  const [category, setCategory] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit({ category })
  }

  return (
    <div className="mt-4 p-4 bg-surface border border-border rounded-lg">
      <h3 className="font-semibold text-text mb-4">تغيير التصنيف</h3>
      <form onSubmit={handleSubmit} className="space-y-4">
        <input
          type="text"
          value={category}
          onChange={(e) => setCategory(e.target.value)}
          placeholder="التصنيف الجديد"
          className="w-full px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          required
        />
        <div className="flex gap-2">
          <button
            type="submit"
            className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
          >
            تطبيق
          </button>
          <button
            type="button"
            onClick={onCancel}
            className="px-4 py-2 border border-border rounded-lg hover:bg-surface-80 transition-colors"
          >
            إلغاء
          </button>
        </div>
      </form>
    </div>
  )
}
