import { StatCard } from '../../../design-system/components/stat-card';
import { DollarSign, TrendingUp, Database, AlertTriangle, Package, RotateCcw } from 'lucide-react';

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
          ['الوحدات المتاحة', count(report.total_items), Package, 'وحدات موجودة في المخزون', 'featured'],
          ['قيمة المخزون بالتكلفة', value(report.valuation?.total_cost ?? report.total_value), DollarSign, 'بحسب تكلفة الشراء', 'info'],
          ['قيمة البيع المتوقعة', value(report.valuation?.total_retail_with_tax ?? report.valuation?.total_retail), DollarSign, 'للمخزون الجاهز للبيع شامل الضريبة', 'success'],
          ['الربح المحتمل', value(report.valuation?.potential_profit), TrendingUp, 'قبل البيع', 'success'],
          ['منخفض المخزون', count(report.low_stock_items?.length), AlertTriangle, 'منتجات تحتاج إعادة طلب', 'warning'],
          ['بضاعة بلا حركة', count(report.stagnant_items?.length), AlertTriangle, 'لم تُبع منذ 30 يومًا أو أكثر', 'danger'],
        ];
      case 'products':
        return [
          ['المنتجات النشطة', count(report.total_products), Package, 'منتجات متاحة في الدليل', 'featured'],
          ['التصنيفات', count(Object.keys(report.by_category || {}).length), Database, 'تصنيفات المنتجات', 'info'],
          ['منخفض المخزون', count(report.low_stock_count), AlertTriangle, 'يحتاج إعادة طلب', 'warning'],
        ];
      case 'debts':
        return [
          ['إجمالي الديون الأصلية', value(report.total_debt), DollarSign, 'قيمة الدين قبل التحصيل', 'featured'],
          ['تم السداد', value(report.total_paid), DollarSign, 'إجمالي المبلغ الذي سدده العملاء', 'success'],
          ['غير مسدد', value(report.outstanding), DollarSign, 'المبلغ المتبقي على العملاء', 'warning'],
          ['متأخر', value(report.overdue_debt), AlertTriangle, 'يحتاج متابعة', 'danger'],
        ];
      case 'suppliers':
        return [
          ['التجار النشطون', count(report.total_suppliers), Database, 'متاحون للتعامل اليومي', 'featured'],
          ['إجمالي المشتريات', value(report.total_purchases), DollarSign, 'من التجار', 'info'],
          ['المدفوع للتجار', value(report.total_paid), DollarSign, 'دفعات مسجلة', 'success'],
          ['المستحق للتجار', value(report.total_outstanding), DollarSign, 'الرصيد المفتوح', 'warning'],
        ];
      case 'purchases':
        return [
          ['إجمالي المشتريات', value(report.total_cost), DollarSign, 'قبل خصم مرتجعات التجار', 'featured'],
          ['مرتجعات التجار', value(report.supplier_return_credits), RotateCcw, 'قيمة المرتجعات المكتملة', 'warning'],
          ['صافي المشتريات', value(report.net_purchases ?? Number(report.total_cost || 0) - Number(report.supplier_return_credits || 0)), DollarSign, 'بعد مرتجعات التجار', 'success'],
        ];
      case 'purchases-suppliers':
        return [
          ['مشتريات الفترة', value(report.total_cost), DollarSign, 'قبل خصم مرتجعات الموردين', 'featured'],
          ['مرتجعات الفترة', value(report.supplier_return_credits), RotateCcw, 'المرتجعات المكتملة خلال الفترة', 'warning'],
          ['صافي مشتريات الفترة', value(report.net_purchases ?? Number(report.total_cost || 0) - Number(report.supplier_return_credits || 0)), DollarSign, 'بعد خصم مرتجعات الفترة', 'success'],
          ['المدفوع للموردين', value(report.suppliers_report?.total_paid), DollarSign, 'إجمالي الدفعات المسجلة حتى الآن', 'info'],
          ['المستحق الحالي للموردين', value(report.suppliers_report?.total_outstanding), DollarSign, 'الرصيد المفتوح حاليًا', 'warning'],
          ['الرصيد الدائن الحالي', value(report.suppliers_report?.supplier_credit_balance), DollarSign, 'رصيد الموردين الدائن حاليًا', 'info'],
          ['الموردون النشطون', count(report.suppliers_report?.total_suppliers), Database, 'الحسابات النشطة حاليًا', 'default'],
        ];
      case 'expenses':
        return [
          ['إجمالي المصروفات', value(report.total_expenses), DollarSign, 'خلال الفترة', 'featured'],
          ['عدد الفئات', count(Object.keys(report.by_category || {}).length), Database, 'فئات مصروفات مستخدمة', 'info'],
          ['الأشهر المسجلة', count(report.by_month?.length), Database, 'أشهر تحتوي على مصروفات', 'default'],
        ];
      case 'returns':
        return [
          ['قيمة المرتجعات', value(report.total_refunded), RotateCcw, 'المبلغ المسترد فعليًا', 'featured'],
          ['عدد المرتجعات', count(report.total_returns), RotateCcw, 'مرتجعات مكتملة', 'warning'],
          ['الأسباب', count(Object.keys(report.by_reason || {}).length), Database, 'أسباب مسجلة', 'info'],
          ['المنتجات المرتجعة', count(report.by_product?.length), Package, 'منتجات ظهرت في المرتجعات', 'default'],
        ];
      case 'tax':
        return [
          ['المبيعات الخاضعة للضريبة', value(report.net_taxable_sales ?? report.taxable_sales), DollarSign, 'المبيعات التي تُحسب عليها الضريبة فقط', 'featured'],
          ['المبيعات المعفاة من الضريبة', value(report.exempt_sales), DollarSign, 'مبيعات لا تحتوي على ضريبة مسجلة', 'info'],
          ['الضريبة المستحقة على المبيعات', value(report.net_tax_collected ?? report.tax_collected), DollarSign, 'الضريبة المسجلة على المبيعات الخاضعة فقط', 'warning'],
          ['إجمالي المرتجعات', value(report.returns_total), RotateCcw, 'يُخصم من إجمالي المبيعات للوصول إلى الصافي', 'danger'],
          ['صافي المبيعات بعد المرتجعات', value(report.net_sales_total), DollarSign, 'إجمالي المبيعات ناقص إجمالي المرتجعات', 'success'],
        ];
      case 'net-sales':
        return [
          ['صافي المبيعات', value(report.net_revenue), DollarSign, 'بعد خصم المرتجعات وقبل الضريبة', 'featured'],
          ['صافي عمليات البيع', count(report.net_sales), Database, 'عدد المبيعات بعد خصم المرتجعات', 'success'],
          ['إجمالي المبيعات', value(report.gross_revenue), DollarSign, 'قبل خصم المرتجعات والضريبة', 'info'],
          ['المدفوع المسجل على الفواتير', value(report.total_paid), DollarSign, 'ما سجله النظام كمدفوع من العملاء', 'success'],
          ['النقد المستلم', value(report.cash_received), DollarSign, 'ما أدخله الكاشير قبل احتساب المردود', 'info'],
          ['المردود النقدي', value(report.change_amount), RotateCcw, 'عاد إلى العملاء ولا يدخل في المبيعات', 'warning'],
          ['نسبة المرتجعات', loading ? '...' : `${Number(report.return_rate || 0).toFixed(1)}%`, RotateCcw, 'من إجمالي العمليات', 'warning'],
        ];
      case 'sales-profit':
        return [
          ['إجمالي المبيعات', value(report.total_revenue), DollarSign, 'بعد خصم المرتجعات وقبل الضريبة', 'featured'],
          ['المرتجعات', value(report.total_refunded), RotateCcw, 'إجمالي المبالغ المستردة', 'warning'],
          ['صافي الربح', value(report.net_profit), TrendingUp, 'بعد التكلفة والمصروفات', 'featured'],
          ['تكلفة شراء البضاعة', value(report.total_cogs), DollarSign, 'تكلفة المنتجات التي تم بيعها', 'warning'],
          ['ما دفعه الزبائن', value(report.total_paid), DollarSign, 'مجموع المدفوعات على الفواتير خلال الفترة', 'success'],
          ['النقد الذي استلمه الصندوق', value(report.cash_received), DollarSign, 'المبلغ الذي دفعه الزبائن نقدًا', 'info'],
          ['الباقي للزبائن', value(report.change_amount), RotateCcw, 'المبلغ الذي أُعيد للزبائن', 'warning'],
          ['مصاريف المحل', value(report.total_expenses), DollarSign, 'خلال الفترة المحددة', 'danger'],
          ['عدد المبيعات', count(report.total_sales), Database, 'عمليات البيع المكتملة', 'info'],
          ['نسبة الربح', loading ? '...' : `${Number(report.profit_margin || 0).toFixed(1)}%`, TrendingUp, 'نسبة الربح من المبيعات', 'success'],
        ];
      case 'profit':
        return [
          ['صافي الربح', value(report.net_profit), TrendingUp, 'الربح الصافي', 'featured'],
          ['صافي المبيعات قبل الضريبة', value(report.total_revenue), DollarSign, 'المبيعات', 'info'],
          ['تكلفة البضاعة المباعة', value(report.total_cogs), DollarSign, 'التكلفة', 'warning'],
          ['المصاريف التشغيلية', value(report.total_expenses), DollarSign, 'المصروفات', 'danger'],
          ['نسبة الربح من المبيعات', loading ? '...' : `${Number(report.profit_margin || 0).toFixed(1)}%`, TrendingUp, 'الربح', 'success'],
        ];
      default:
        return [
          ['صافي المبيعات', value(report.total_revenue), DollarSign, 'بعد خصم المرتجعات وقبل الضريبة', 'featured'],
          ['إجمالي الربح', value(report.gross_profit), TrendingUp, 'قبل المصروفات وقبل الضريبة', 'success'],
          ['عدد المبيعات', count(report.total_sales), Database, 'عمليات البيع المكتملة', 'default'],
          ['الوحدات المباعة', count(report.total_items_sold), Package, 'إجمالي القطع المباعة', 'info'],
        ];
    }
  })();

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '10px' }}>
      {cards.map(([title, cardValue, icon, subtitle, variant]) => (
        <StatCard
          key={String(title)}
          title={String(title)}
          value={cardValue as string}
          icon={icon as any}
          subtitle={String(subtitle)}
          variant={variant as any}
          compact
          size="sm"
        />
      ))}
    </div>
  );
}
