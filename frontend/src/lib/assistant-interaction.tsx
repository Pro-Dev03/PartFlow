import type { ReactNode } from 'react';
import {
  AlertTriangle,
  BarChart3,
  CircleHelp,
  Coins,
  Database,
  HeartHandshake,
  PackageSearch,
  ShoppingCart,
  Sparkles,
  Users,
  WifiOff,
} from 'lucide-react';

export type AssistantIntent =
  | 'GREETING' | 'DAILY_SUMMARY' | 'ADVICE' | 'ALERTS' | 'NAVIGATION' | 'SALES' | 'INVENTORY' | 'LOW_STOCK' | 'DEBTS' | 'PROFIT'
  | 'PURCHASES' | 'SUPPLIERS' | 'EXPENSES' | 'SETTINGS' | 'RECENT' | 'CUSTOMERS' | 'PAYMENTS'
  | 'RETURNS' | 'REPORTS' | 'SYSTEM_STATUS' | 'MONTHLY_SUMMARY'
  | 'ASSISTANT_IDENTITY' | 'ASSISTANT_CAPABILITIES' | 'ASSISTANT_BEHAVIOR' | 'CLARIFICATION'
  | 'THANKS' | 'GOODBYE' | 'SMALL_TALK' | 'UNKNOWN';

export type AssistantCueTone = 'neutral' | 'data' | 'finance' | 'success' | 'attention' | 'error' | 'offline';

export interface AssistantInteraction {
  label: string;
  icon: ReactNode;
  tone: AssistantCueTone;
}

const interactions: Record<AssistantIntent, AssistantInteraction> = {
  GREETING: { label: 'أهلا فيك', icon: <HeartHandshake size={15} />, tone: 'success' },
  DAILY_SUMMARY: { label: 'ألخص وضع المحل...', icon: <Sparkles size={15} />, tone: 'data' },
  ADVICE: { label: 'أرتب لك الخطوة الأنسب...', icon: <Sparkles size={15} />, tone: 'data' },
  ALERTS: { label: 'أراجع التنبيهات الحالية...', icon: <AlertTriangle size={15} />, tone: 'attention' },
  NAVIGATION: { label: 'أفتح لك القسم المطلوب...', icon: <Sparkles size={15} />, tone: 'neutral' },
  MONTHLY_SUMMARY: { label: 'أراجع نتائج الشهر...', icon: <BarChart3 size={15} />, tone: 'data' },
  SALES: { label: 'أراجع مبيعات اليوم...', icon: <BarChart3 size={15} />, tone: 'data' },
  INVENTORY: { label: 'أتحقق من المخزون...', icon: <PackageSearch size={15} />, tone: 'data' },
  LOW_STOCK: { label: 'أبحث عن النواقص...', icon: <PackageSearch size={15} />, tone: 'attention' },
  DEBTS: { label: 'أراجع الديون...', icon: <Coins size={15} />, tone: 'finance' },
  PROFIT: { label: 'أحسب النتيجة...', icon: <BarChart3 size={15} />, tone: 'data' },
  PURCHASES: { label: 'أراجع آخر المشتريات...', icon: <ShoppingCart size={15} />, tone: 'data' },
  SUPPLIERS: { label: 'أراجع أرصدة الموردين...', icon: <Users size={15} />, tone: 'finance' },
  EXPENSES: { label: 'أراجع مصاريف اليوم...', icon: <Coins size={15} />, tone: 'finance' },
  SETTINGS: { label: 'أتحقق من الإعدادات...', icon: <Database size={15} />, tone: 'neutral' },
  RECENT: { label: 'أراجع آخر العمليات...', icon: <Database size={15} />, tone: 'data' },
  CUSTOMERS: { label: 'أراجع بيانات العملاء...', icon: <Users size={15} />, tone: 'neutral' },
  PAYMENTS: { label: 'أراجع التحصيل...', icon: <Coins size={15} />, tone: 'finance' },
  RETURNS: { label: 'أراجع المرتجعات...', icon: <ShoppingCart size={15} />, tone: 'data' },
  REPORTS: { label: 'أحلل التقارير...', icon: <BarChart3 size={15} />, tone: 'data' },
  SYSTEM_STATUS: { label: 'أتحقق من حالة النظام...', icon: <Database size={15} />, tone: 'neutral' },
  ASSISTANT_IDENTITY: { label: 'أنا مساعد PartFlow', icon: <HeartHandshake size={15} />, tone: 'success' },
  ASSISTANT_CAPABILITIES: { label: 'أوضح لك كيف أساعد...', icon: <Sparkles size={15} />, tone: 'neutral' },
  ASSISTANT_BEHAVIOR: { label: 'معك حق، أراجع ردي...', icon: <AlertTriangle size={15} />, tone: 'attention' },
  CLARIFICATION: { label: 'وضّح لي أكثر...', icon: <CircleHelp size={15} />, tone: 'neutral' },
  THANKS: { label: 'العفو', icon: <HeartHandshake size={15} />, tone: 'success' },
  GOODBYE: { label: 'إلى اللقاء', icon: <HeartHandshake size={15} />, tone: 'neutral' },
  SMALL_TALK: { label: 'أنا معك...', icon: <HeartHandshake size={15} />, tone: 'neutral' },
  UNKNOWN: { label: 'خليني أفهم سؤالك...', icon: <CircleHelp size={15} />, tone: 'neutral' },
};

const normalize = (value: string) => value.trim().toLowerCase();

export function inferAssistantIntent(message: string): AssistantIntent {
  const text = normalize(message);
  if (!text || /^[؟?!.,،؛:ـ]+$/.test(text)) return 'CLARIFICATION';
  if (/ليش.*(تكرر|بتكرر|تعيد)|تكرر.*(كلام|رد)|نفس.*(الكلام|الرد)|كرر/.test(text)) return 'ASSISTANT_BEHAVIOR';
  if (/مين\s*(انت|إنت|أنت)|من\s*(انت|إنت|أنت)|(?:انت|إنت|أنت)\s*من|who\s*are\s*you/.test(text)) return 'ASSISTANT_IDENTITY';
  if (/شو\s*(بتقدر|تقدر|فيك)\s*(تعمل|تساعد)|قدراتك|ماذا\s*تستطيع|شو\s*بتساعد/.test(text)) return 'ASSISTANT_CAPABILITIES';
  if (/شكرا|شكرًا|يعطيك\s*العافية|thanks|thank\s*you/.test(text)) return 'THANKS';
  if (/باي|مع\s*السلامة|إلى\s*اللقاء|goodbye|bye/.test(text)) return 'GOODBYE';
  if (/تمام|ممتاز|ماشي|حسنا|حسنًا|اوكي|أوكي|كيفك|شو\s*الأخبار|شو\s*الاخبار|كيف\s*حالك/.test(text)) return 'SMALL_TALK';
  if (/تنبيه|تنبيهات|يحتاج\s*انتباه|شو\s*في\s*مشاكل|في\s*مشكلة|تحذير/.test(text)) return 'ALERTS';
  if (/نصيحة|نصائح|اقتراح|اقتراحات|شو\s*(أعمل|اعمل|لازم|بتنصح)|ماذا\s*أفعل|كيف\s*أحسن|أولوياتي/.test(text)) return 'ADVICE';
  if (/(افتح|فتح|اذهب|روح|خذني|انتقل|اعرض).*(نقطة\s*البيع|المبيعات|POS|البيع|المخزون|الديون|التقارير|المشتريات|المصروفات|العملاء|الموردين|الأرشيف|سجل النظام)/.test(text)) return 'NAVIGATION';
  if (/مرحبا|اهلا|أهلا|السلام|هلا|صباح|مساء/.test(text)) return 'GREETING';
  if (/هذا الشهر|هالشهر|الشهر/.test(text)) return 'MONTHLY_SUMMARY';
  if (/مخزون|منتج|منتجات|بضاعة|قطعة/.test(text)) return /منخفض|ناقصة|تخلص|قربت/.test(text) ? 'LOW_STOCK' : 'INVENTORY';
  if (/دين|ديون|مستحق|تحصيل|مديون|باقي.*للزبائن|للزبائن/.test(text)) return 'DEBTS';
  if (/مصروف|صرفنا|مصاريف/.test(text)) return 'EXPENSES';
  if (/ربح|أرباح|صافي/.test(text)) return 'PROFIT';
  if (/شراء|مشتريات/.test(text)) return 'PURCHASES';
  if (/مورد/.test(text)) return 'SUPPLIERS';
  if (/عميل|عملاء/.test(text)) return 'CUSTOMERS';
  if (/دفعة|دفعات|تحصيل اليوم/.test(text)) return 'PAYMENTS';
  if (/مرتجع|مرتجعات/.test(text)) return 'RETURNS';
  if (/تقرير|تقارير/.test(text)) return 'REPORTS';
  if (/حالة النظام|النظام شغال/.test(text)) return 'SYSTEM_STATUS';
  if (/مبيعات|بعنا|بيع|دخل|إيرادات/.test(text)) return 'SALES';
  if (/ملخص|وضع المتجر|وضع المحل|كيف.*اليوم|شو.*اليوم|أولويات/.test(text)) return 'DAILY_SUMMARY';
  return 'UNKNOWN';
}

export function getAssistantInteraction(intent: string | undefined, message = ''): AssistantInteraction {
  const normalizedIntent = String(intent || '').toUpperCase() as AssistantIntent;
  return interactions[normalizedIntent] || interactions[inferAssistantIntent(message)];
}

export function getAssistantNavigationPath(message: string): string | null {
  const text = normalize(message);
  if (!/(افتح|فتح|اذهب|روح|خذني|انتقل|اعرض|اعمل|ابدأ|ابدا|سجل|نفذ|بيع\s*جديد|عملية\s*بيع)/.test(text)) return null;
  if (/عملية\s*بيع|بيع\s*جديد|نقطة\s*البيع|المبيعات|POS|البيع/.test(text)) return '/app/sales';
  if (/مخزون/.test(text)) return '/app/inventory';
  if (/ديون/.test(text)) return '/app/debts';
  if (/تقارير/.test(text)) return '/app/reports';
  if (/مشتريات/.test(text)) return '/app/purchases';
  if (/مصروفات/.test(text)) return '/app/expenses';
  if (/عملاء/.test(text)) return '/app/customers';
  if (/موردين|موردون/.test(text)) return '/app/suppliers';
  if (/أرشيف|سجل النظام/.test(text)) return '/app/audit';
  return null;
}

export function getSystemInteraction(state: 'error' | 'offline'): AssistantInteraction {
  return state === 'offline'
    ? { label: 'الاتصال غير متاح حاليًا', icon: <WifiOff size={15} />, tone: 'offline' }
    : { label: 'صار خطأ وأنا أراجع البيانات', icon: <AlertTriangle size={15} />, tone: 'error' };
}
