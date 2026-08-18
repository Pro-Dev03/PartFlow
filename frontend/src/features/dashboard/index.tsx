import * as React from "react"
import { KPICard } from "../../components/dashboard/KPICard"
import { PartCard } from "../../components/inventory/PartCard"
import { Sidebar } from "../../components/navigation/Sidebar"
import { TopBar } from "../../components/navigation/TopBar"
import { useNavigate } from 'react-router-dom'
import { Plus, Scan } from 'lucide-react'
import { useUIStore } from '@stores'

export function Dashboard() {
  const navigate = useNavigate()
  const { addNotification } = useUIStore()

  const handleScanBarcode = () => {
    navigate('/barcode')
    addNotification({
      type: 'info',
      message: 'فتح ماسح الباركود',
      duration: 2000
    })
  }

  const handleNewSale = () => {
    navigate('/sales')
    addNotification({
      type: 'info',
      message: 'فتح عملية بيع جديدة',
      duration: 2000
    })
  }

  const handlePartSell = (partName: string) => {
    addNotification({
      type: 'success',
      message: `تمت إضافة ${partName} إلى سلة البيع`,
      duration: 3000
    })
  }

  return (
    <div className="grid grid-cols-[1fr_84px] min-h-screen" style={{ direction: 'rtl' }}>
      {/* Main Content */}
      <main className="p-7 pb-24 max-w-[1400px]">
        <TopBar />

        {/* KPI Grid */}
        <div className="grid grid-cols-4 gap-3.5 mb-6.5">
          <KPICard 
            label="مبيعات اليوم" 
            value="184.500" 
            currency="د.أ" 
            sub="23 عملية بيع" 
            trend="▲ 12%" 
          />
          <KPICard 
            label="القطع المباعة اليوم" 
            value="31" 
            sub="15 جديدة · 16 مستعملة" 
          />
          <KPICard 
            label="إجمالي الديون المستحقة" 
            value="1,240.000" 
            sub="7 زبائن عليهم دين نشط" 
            variant="warn" 
          />
          <KPICard 
            label="قطع بحاجة لإعادة تخزين" 
            value="5" 
            sub="راجع شاشة المخزون" 
            variant="danger" 
          />
        </div>

        {/* Content Grid */}
        <div className="grid grid-cols-[1.6fr_1fr] gap-3.5 mb-7.5">
          {/* Chart Panel */}
          <div className="bg-surface border border-border rounded-[8px] p-5">
            <div className="flex justify-between items-center mb-4.5">
              <h3 className="text-[14.5px] font-semibold text-text">حركة البيع — آخر 7 أيام</h3>
              <span className="text-[11px] text-text-faint">بالدينار الأردني</span>
            </div>
            <div className="flex items-end gap-2.5 h-[140px]">
              {[
                { day: 'سبت', height: '55%' },
                { day: 'أحد', height: '70%' },
                { day: 'اثنين', height: '40%' },
                { day: 'ثلاثاء', height: '85%' },
                { day: 'أربعاء', height: '62%' },
                { day: 'خميس', height: '95%' },
                { day: 'اليوم', height: '100%', highlight: true },
              ].map((item, index) => (
                <div key={index} className="flex-1 flex flex-col items-center gap-2">
                  <div 
                    className={`w-full rounded-t-[4px] opacity-85 transition-all duration-200 hover:opacity-100 ${
                      item.highlight 
                        ? 'bg-gradient-to-b from-accent to-[#00a37e]' 
                        : 'bg-gradient-to-b from-accent to-accent-dim'
                    }`}
                    style={{ height: item.height }}
                  />
                  <div className={`text-[10.5px] font-mono ${
                    item.highlight ? 'text-accent' : 'text-text-faint'
                  }`}>
                    {item.day}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Activity Panel */}
          <div className="bg-surface border border-border rounded-[8px] p-5">
            <div className="flex justify-between items-center mb-4.5">
              <h3 className="text-[14.5px] font-semibold text-text">آخر العمليات</h3>
            </div>
            <div className="flex flex-col gap-0.5">
              {[
                { name: 'رامة كورسير 8GB مستعملة', time: 'قبل 12 دقيقة', amount: '+25.000' },
                { name: 'معالج Core i5-10400', time: 'قبل 40 دقيقة', amount: '+95.000' },
                { name: 'دفعة دين — محمد أبو عيسى', time: 'قبل ساعة', amount: '-50.000', warn: true },
                { name: 'شاحن لابتوب 65W', time: 'قبل ساعتين', amount: '+12.000' },
              ].map((item, index) => (
                <div key={index} className="flex justify-between items-center py-2.5 px-1 border-b border-border last:border-b-0 text-[12.5px]">
                  <div>
                    <div className="text-text font-medium">{item.name}</div>
                    <div className="text-text-faint text-[11px] mt-0.5">{item.time}</div>
                  </div>
                  <div className={`font-mono font-semibold ${item.warn ? 'text-warn' : 'text-accent'}`}>
                    {item.amount}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Inventory Section */}
        <div className="flex justify-between items-center mb-4 flex-wrap gap-3">
          <h2 className="text-[16px] font-semibold text-text">المخزون</h2>
          <div className="flex gap-2 flex-wrap">
            {['الكل', 'جديد', 'مستعمل', 'نفدت الكمية', 'رامات', 'معالجات'].map((filter, index) => (
              <button
                key={index}
                className={`px-3.5 py-1.5 rounded-full text-[12px] border border-border text-text-dim cursor-pointer transition-all duration-150 whitespace-nowrap ${
                  index === 0 
                    ? 'bg-accent-dim border-accent text-accent' 
                    : 'hover:border-text-faint hover:text-text'
                }`}
              >
                {filter}
              </button>
            ))}
          </div>
        </div>

        {/* Parts Grid */}
        <div className="grid grid-cols-[repeat(auto-fill,minmax(230px,1fr))] gap-3.5">
          <PartCard
            name="رامة كورسير Vengeance 8GB DDR4"
            badges={[
              { type: 'used', label: '⚡ مستعملة' },
              { type: 'stock-ok', label: '🟢 متوفر (4)' }
            ]}
            barcode="PF-USED-000452"
            price={25}
            icon="ram"
            onSell={() => handlePartSell('رامة كورسير Vengeance 8GB DDR4')}
          />
          <PartCard
            name="معالج Intel Core i5-10400F"
            badges={[
              { type: 'new', label: '🔵 جديد' },
              { type: 'stock-ok', label: '🟢 متوفر (7)' }
            ]}
            barcode="8801643-556231"
            price={95}
            icon="cpu"
            onSell={() => handlePartSell('معالج Intel Core i5-10400F')}
          />
          <PartCard
            name="كرت شاشة GTX 1660 Super مستعمل"
            badges={[
              { type: 'used', label: '⚡ مستعملة' },
              { type: 'stock-low', label: '🔴 آخر قطعة' }
            ]}
            barcode="PF-USED-000398"
            price={210}
            lowStock
            icon="gpu"
            onSell={() => handlePartSell('كرت شاشة GTX 1660 Super مستعمل')}
          />
          <PartCard
            name="شاحن لابتوب أصلي 65W Type-C"
            badges={[
              { type: 'new', label: '🔵 جديد' },
              { type: 'stock-ok', label: '🟢 متوفر (14)' }
            ]}
            barcode="6934567-118820"
            price={12}
            icon="charger"
            onSell={() => handlePartSell('شاحن لابتوب أصلي 65W Type-C')}
          />
          <PartCard
            name="قرص SSD كنكستون 240GB مستعمل"
            badges={[
              { type: 'used', label: '⚡ مستعملة' },
              { type: 'stock-ok', label: '🟢 متوفر (2)' }
            ]}
            barcode="PF-USED-000471"
            price={18}
            icon="ssd"
            onSell={() => handlePartSell('قرص SSD كنكستون 240GB مستعمل')}
          />
        </div>
      </main>

      {/* Sidebar */}
      <Sidebar />

      {/* FAB Buttons */}
      <div className="fixed bottom-6 left-6 flex flex-col gap-2.5 items-start z-50">
        <button 
          onClick={handleScanBarcode}
          className="flex items-center gap-2.5 bg-surface-elevated border border-border text-text px-4.5 py-3 rounded-full text-[13px] font-semibold cursor-pointer shadow-[0_6px_20px_rgba(0,0,0,0.4)] transition-all duration-150 hover:-translate-y-0.5"
        >
          <Scan className="w-[17px] h-[17px]" />
          مسح باركود
        </button>
        <button 
          onClick={handleNewSale}
          className="flex items-center gap-2.5 bg-accent text-[#04140F] border border-accent px-4.5 py-3 rounded-full text-[13px] font-semibold cursor-pointer shadow-[0_6px_20px_rgba(0,0,0,0.4)] transition-all duration-150 hover:-translate-y-0.5"
        >
          <Plus className="w-[17px] h-[17px]" />
          بيع جديد
        </button>
      </div>
    </div>
  )
}

export default Dashboard