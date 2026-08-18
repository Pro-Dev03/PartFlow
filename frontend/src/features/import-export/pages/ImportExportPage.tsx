import { useState } from 'react'
import { clsx } from 'clsx'
import { ImportWizard } from '../components/ImportWizard'
import { ExportBuilder } from '../components/ExportBuilder'
import { ImportHistory } from '../components/ImportHistory'
import { ExportHistory } from '../components/ExportHistory'

type TabType = 'import' | 'export' | 'import-history' | 'export-history'

export function ImportExportPage() {
  const [activeTab, setActiveTab] = useState<TabType>('import')

  const tabs = [
    { id: 'import' as TabType, label: 'استيراد', icon: '📥' },
    { id: 'export' as TabType, label: 'تصدير', icon: '📤' },
    { id: 'import-history' as TabType, label: 'سجل الاستيراد', icon: '📋' },
    { id: 'export-history' as TabType, label: 'سجل التصدير', icon: '📋' },
  ]

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-text mb-2">الاستيراد والتصدير</h1>
        <p className="text-muted">نقل البيانات من وإلى النظام</p>
      </div>

      {/* Tabs */}
      <div className="bg-surface rounded-lg">
        <div className="border-b border-border">
          <nav className="flex gap-4 px-6 overflow-x-auto">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={clsx(
                  'px-4 py-3 text-sm font-medium border-b-2 transition-colors whitespace-nowrap',
                  activeTab === tab.id
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted hover:text-text'
                )}
              >
                <span className="mr-2">{tab.icon}</span>
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Content */}
        <div className="p-6">
          {activeTab === 'import' && <ImportWizard />}
          {activeTab === 'export' && <ExportBuilder />}
          {activeTab === 'import-history' && <ImportHistory />}
          {activeTab === 'export-history' && <ExportHistory />}
        </div>
      </div>
    </div>
  )
}

export default ImportExportPage
