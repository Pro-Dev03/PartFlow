import { StatCard } from '../../../components/ui/stat-card';
import { DollarSign, TrendingUp, Database, Target, AlertTriangle, Package, RotateCcw } from 'lucide-react';

interface ReportStatsProps {
  data?: any;
  loading?: boolean;
  reportType?: string;
}

export function ReportStats({ data, loading, reportType = 'sales' }: ReportStatsProps) {
  const report = data?.data ?? data ?? {};
  const value = (number: unknown, currency = true) => loading ? '...' :
    `${currency ? '₪' : ''}${Number(number || 0).toLocaleString()}`;
  const count = (number: unknown) => loading ? '...' : Number(number || 0).toLocaleString();
  const cards = (() => {
    switch (reportType) {
      case 'inventory':
        return [
          ['القطع المتاحة', count(report.total_items), Package, 'وحدة في المخزون', 'featured'],
          ['قيمة التكلفة', value(report.valuation?.total_cost ?? report.total_value), DollarSign, 'حسب تكلفة الشراء', 'info'],
          ['قيمة البيع', value(report.valuation?.total_retail), DollarSign, 'القيمة المتوقعة', 'success'],
          ['الربح المحتمل', value(report.valuation?.potential_profit), TrendingUp, 'قبل البيع', 'success'],
          ['منخفض المخزون', count(report.low_stock_items?.length), AlertTriangle, 'منتجات تحتاج إعادة طلب', 'warning'],
          ['راكد', count(report.stagnant_items?.length), AlertTriangle, 'بدون حركة 30 يومًا', 'danger'],
        ];
      case 'debts':
        return [
          ['إجمالي الديون', value(report.total_debt), DollarSign, 'قيمة الديون الأصلية', 'featured'],
          ['المدفوع', value(report.total_paid), DollarSign, 'دفعات مسجلة', 'success'],
          ['المتبقي', value(report.outstanding), DollarSign, 'الرصيد المستحق', 'warning'],
          ['متأخر', value(report.overdue_debt), AlertTriangle, 'يحتاج متابعة', 'danger'],
          ['العملاء', count(report.by_customer?.length), Database, 'لديهم ديون', 'default'],
          ['الدفعات', count(report.payment_history?.length), Database, 'سجل الدفعات', 'info'],
        ];
      case 'suppliers':
        return [
          ['الموردون النشطون', count(report.total_suppliers), Database, 'متاحون للتعامل اليومي', 'featured'],
          ['الموردون المعطلون', count(report.inactive_suppliers), Database, 'سجلات محفوظة للتاريخ', 'default'],
          ['إجمالي المشتريات', value(report.total_purchases), DollarSign, 'من الموردين', 'info'],
          ['المدفوع للموردين', value(report.total_paid), DollarSign, 'دفعات مسجلة', 'success'],
          ['المستحق للموردين', value(report.total_outstanding), DollarSign, 'الرصيد المفتوح', 'warning'],
          ['لديهم رصيد', count(report.suppliers_with_balance?.length), Database, 'يحتاج متابعة', 'danger'],
        ];
      case 'purchases':
        return [
          ['إجمالي المشتريات', value(report.total_cost), DollarSign, 'خلال الفترة', 'featured'],
          ['عدد الطلبات', count(report.total_purchases), Database, 'طلبات شراء', 'default'],
          ['الموردون', count(report.by_supplier?.length), Database, 'موردون نشطون', 'info'],
          ['الأصناف', count(Object.values(report.by_category || {}).reduce((a: number, b: any) => a + Number(b || 0), 0)), Package, 'قطع مشتراة', 'success'],
        ];
      case 'expenses':
        return [
          ['إجمالي المصروفات', value(report.total_expenses), DollarSign, 'خلال الفترة', 'featured'],
          ['الفئات', count(Object.keys(report.by_category || {}).length), Database, 'فئات مستخدمة', 'info'],
          ['المصروفات الشهرية', count(report.by_month?.length), Database, 'أشهر مسجلة', 'default'],
        ];
      case 'returns':
        return [
          ['قيمة المرتجعات', value(report.total_refunded), RotateCcw, 'مبالغ مستردة', 'featured'],
          ['عدد المرتجعات', count(report.total_returns), RotateCcw, 'مرتجعات مكتملة', 'warning'],
          ['الأسباب', count(Object.keys(report.by_reason || {}).length), Database, 'أسباب مسجلة', 'info'],
          ['المنتجات', count(report.by_product?.length), Package, 'منتجات مرتجعة', 'default'],
        ];
      case 'used-items': {
        const items = Array.isArray(report.items) ? report.items : [];
        const getStatus = (item: any) => String(item.status || '').trim().toUpperCase();
        const availableItems = items.filter((item: any) => getStatus(item) === 'AVAILABLE');
        const soldItems = items.filter((item: any) => getStatus(item) === 'SOLD');
        const damagedItems = items.filter((item: any) => getStatus(item) === 'DAMAGED');
        const availableCost = availableItems.reduce((sum: number, item: any) => sum + Number(item.purchase_cost || 0), 0);
        const availableRetail = availableItems.reduce((sum: number, item: any) => sum + Number(item.selling_price || 0), 0);
        return [
          ['إجمالي القطع', count(items.length), Package, 'قطع مستعملة مسجلة', 'featured'],
          ['المتاحة', count(availableItems.length), Package, 'جاهزة للبيع', 'success'],
          ['المباعة', count(soldItems.length), TrendingUp, 'قطع تم بيعها', 'info'],
          ['التالفة', count(damagedItems.length), AlertTriangle, 'ليست ضمن المخزون المتاح', 'danger'],
          ['قيمة التكلفة', value(availableCost), DollarSign, 'للقطع المتاحة فقط', 'warning'],
          ['قيمة البيع', value(availableRetail), DollarSign, 'المتوقع من المتاح', 'success'],
          ['الربح المحتمل', value(availableRetail - availableCost), TrendingUp, 'للقطع المتاحة فقط', 'default'],
        ];
      }
      case 'net-sales':
        return [
          ['صافي الإيراد', value(report.net_revenue), DollarSign, 'بعد خصم المرتجعات', 'featured'],
          ['صافي العمليات', count(report.net_sales), Database, 'عمليات البيع', 'success'],
          ['المبيعات الإجمالية', value(report.gross_revenue), DollarSign, 'قبل المرتجعات', 'info'],
          ['نسبة المرتجعات', loading ? '...' : `${Number(report.return_rate || 0).toFixed(1)}%`, RotateCcw, 'من إجمالي العمليات', 'warning'],
        ];
      case 'profit':
        return [
          ['صافي الربح', value(report.net_profit), TrendingUp, 'بعد المصروفات', 'featured'],
          ['الإيراد', value(report.total_revenue), DollarSign, 'إجمالي المبيعات', 'info'],
          ['التكلفة', value(report.total_cogs), DollarSign, 'تكلفة البضاعة', 'warning'],
          ['المصروفات', value(report.total_expenses), DollarSign, 'مصروفات الفترة', 'danger'],
          ['هامش الربح', loading ? '...' : `${Number(report.profit_margin || 0).toFixed(1)}%`, TrendingUp, 'من الإيراد', 'success'],
        ];
      default:
        return [
          ['إجمالي المبيعات', value(report.total_revenue), DollarSign, 'خلال الفترة', 'featured'],
          ['إجمالي الربح', value(report.gross_profit), TrendingUp, 'قبل المصروفات', 'success'],
          ['العمليات', count(report.total_sales), Database, 'عمليات بيع', 'default'],
          ['الوحدات', count(report.total_items_sold), Package, 'قطع مباعة', 'info'],
          ['نقدي', value(report.cash_revenue), DollarSign, 'إيراد نقدي', 'default'],
          ['دين', value(report.credit_revenue), DollarSign, 'إيراد على الحساب', 'warning'],
        ];
    }
  })();

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(210px, 1fr))', gap: '14px' }}>
      {cards.map(([title, cardValue, icon, subtitle, variant]) => (
        <StatCard key={String(title)} title={String(title)} value={cardValue as string}
          icon={icon as any} subtitle={String(subtitle)} variant={variant as any} />
      ))}
    </div>
  );
}
