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
          ['إجمالي قيمة البيع للمخزون المتاح', value(report.valuation?.total_retail_with_tax ?? report.valuation?.total_retail), DollarSign, 'شامل الضريبة للمنتجات الجاهزة للبيع', 'success'],
          ['الربح المحتمل', value(report.valuation?.potential_profit), TrendingUp, 'قبل البيع', 'success'],
          ['منخفض المخزون', count(report.low_stock_items?.length), AlertTriangle, 'منتجات تحتاج إعادة طلب', 'warning'],
          ['راكد', count(report.stagnant_items?.length), AlertTriangle, 'بدون حركة 30 يومًا', 'danger'],
        ];
      case 'products':
        return [
          ['المنتجات النشطة', count(report.total_products), Package, 'منتجات متاحة في الدليل', 'featured'],
          ['التصنيفات', count(Object.keys(report.by_category || {}).length), Database, 'تصنيفات المنتجات', 'info'],
          ['منخفض المخزون', count(report.low_stock_count), AlertTriangle, 'يحتاج إعادة طلب', 'warning'],
        ];
      case 'debts':
        return [
          ['إجمالي الديون', value(report.total_debt), DollarSign, 'قيمة الديون الأصلية', 'featured'],
          ['المدفوع', value(report.total_paid), DollarSign, 'دفعات مسجلة', 'success'],
          ['المتبقي', value(report.outstanding), DollarSign, 'الرصيد المستحق', 'warning'],
          ['متأخر', value(report.overdue_debt), AlertTriangle, 'يحتاج متابعة', 'danger'],
        ];
      case 'suppliers':
        return [
          ['الموردون النشطون', count(report.total_suppliers), Database, 'متاحون للتعامل اليومي', 'featured'],
          ['إجمالي المشتريات', value(report.total_purchases), DollarSign, 'من الموردين', 'info'],
          ['المدفوع للموردين', value(report.total_paid), DollarSign, 'دفعات مسجلة', 'success'],
          ['المستحق للموردين', value(report.total_outstanding), DollarSign, 'الرصيد المفتوح', 'warning'],
        ];
      case 'purchases':
        return [
          ['إجمالي المشتريات', value(report.total_cost), DollarSign, 'قبل خصم مرتجعات الموردين', 'featured'],
          ['مرتجعات الموردين', value(report.supplier_return_credits), RotateCcw, 'قيمة المرتجعات المكتملة', 'warning'],
          ['صافي المشتريات', value(report.net_purchases ?? Number(report.total_cost || 0) - Number(report.supplier_return_credits || 0)), DollarSign, 'بعد مرتجعات الموردين', 'success'],
          ['المدفوع للموردين', value(report.total_paid), DollarSign, 'دفعات مسجلة', 'info'],
          ['المستحق للموردين', value(report.total_outstanding), DollarSign, 'الرصيد المفتوح', 'danger'],
        ];
      case 'expenses':
        return [
          ['إجمالي المصروفات', value(report.total_expenses), DollarSign, 'خلال الفترة', 'featured'],
          ['الفئات', count(Object.keys(report.by_category || {}).length), Database, 'فئات مستخدمة', 'info'],
          ['المصروفات الشهرية', count(report.by_month?.length), Database, 'أشهر مسجلة', 'default'],
        ];
      case 'returns':
        return [
          ['قيمة المرتجعات', value(report.total_refunded), RotateCcw, 'المبلغ المسترد فعليًا', 'featured'],
          ['عدد المرتجعات', count(report.total_returns), RotateCcw, 'مرتجعات مكتملة', 'warning'],
          ['الأسباب', count(Object.keys(report.by_reason || {}).length), Database, 'أسباب مسجلة', 'info'],
          ['المنتجات', count(report.by_product?.length), Package, 'منتجات مرتجعة', 'default'],
        ];
      case 'tax':
        return [
          ['المبيعات الخاضعة للضريبة', value(report.taxable_sales), DollarSign, 'قبل إضافة الضريبة', 'featured'],
          ['الضريبة المحصلة', value(report.tax_collected), DollarSign, 'المبلغ المستحق ضريبيًا', 'warning'],
          ['المرتجعات', value(report.returns_total), RotateCcw, 'مبالغ مستردة', 'danger'],
          ['صافي المبيعات', value(report.net_sales_total), DollarSign, 'بعد المرتجعات وشامل الضريبة', 'success'],
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
          ['صافي الإيراد', value(report.net_revenue), DollarSign, 'بعد خصم المرتجعات وقبل الضريبة', 'featured'],
          ['صافي العمليات', count(report.net_sales), Database, 'عمليات البيع', 'success'],
          ['المبيعات الإجمالية', value(report.gross_revenue), DollarSign, 'قبل المرتجعات وقبل الضريبة', 'info'],
          ['نسبة المرتجعات', loading ? '...' : `${Number(report.return_rate || 0).toFixed(1)}%`, RotateCcw, 'من إجمالي العمليات', 'warning'],
        ];
      case 'profit':
        return [
          ['صافي الربح', value(report.net_profit), TrendingUp, 'بعد المصروفات وقبل الضريبة', 'featured'],
          ['الإيراد', value(report.total_revenue), DollarSign, 'صافي الإيراد قبل الضريبة', 'info'],
          ['التكلفة', value(report.total_cogs), DollarSign, 'تكلفة البضاعة', 'warning'],
          ['المصروفات', value(report.total_expenses), DollarSign, 'مصروفات الفترة', 'danger'],
          ['هامش الربح', loading ? '...' : `${Number(report.profit_margin || 0).toFixed(1)}%`, TrendingUp, 'من الإيراد', 'success'],
        ];
      default:
        return [
          ['صافي المبيعات', value(report.total_revenue), DollarSign, 'بعد خصم المرتجعات وقبل الضريبة', 'featured'],
          ['إجمالي الربح', value(report.gross_profit), TrendingUp, 'قبل المصروفات وقبل الضريبة', 'success'],
          ['العمليات', count(report.total_sales), Database, 'عمليات البيع الأصلية', 'default'],
          ['الوحدات', count(report.total_items_sold), Package, 'الوحدات المباعة أصلًا', 'info'],
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
