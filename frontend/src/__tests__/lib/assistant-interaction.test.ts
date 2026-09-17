import { describe, expect, it } from 'vitest';
import { getAssistantInteraction, getAssistantNavigationPath, getSystemInteraction, inferAssistantIntent } from '../../lib/assistant-interaction';

describe('assistant interaction cues', () => {
  it.each([
    ['كم مبيعات اليوم؟', 'SALES', 'أراجع مبيعات اليوم...'],
    ['كم قطعة عندي من RTX؟', 'INVENTORY', 'أتحقق من المخزون...'],
    ['كم باقي للزبائن؟', 'DEBTS', 'أراجع الديون...'],
    ['كم صرفنا اليوم؟', 'EXPENSES', 'أراجع مصاريف اليوم...'],
    ['والربح؟', 'PROFIT', 'أحسب النتيجة...'],
    ['شو آخر مشترياتنا؟', 'PURCHASES', 'أراجع آخر المشتريات...'],
    ['ما الذي حدث في هذا الشهر؟', 'MONTHLY_SUMMARY', 'أراجع نتائج الشهر...'],
    ['مين انت؟', 'ASSISTANT_IDENTITY', 'أنا مساعد PartFlow'],
    ['شو بتقدر تعمل؟', 'ASSISTANT_CAPABILITIES', 'أوضح لك كيف أساعد...'],
    ['؟', 'CLARIFICATION', 'وضّح لي أكثر...'],
    ['ليش بتكرر نفس الكلام؟', 'ASSISTANT_BEHAVIOR', 'معك حق، أراجع ردي...'],
    ['شو بتنصحني أعمل؟', 'ADVICE', 'أرتب لك الخطوة الأنسب...'],
    ['شو التنبيهات؟', 'ALERTS', 'أراجع التنبيهات الحالية...'],
    ['افتح نقطة البيع', 'NAVIGATION', 'أفتح لك القسم المطلوب...'],
  ])('maps %s to a contextual cue', (message, intent, label) => {
    expect(inferAssistantIntent(message)).toBe(intent);
    expect(getAssistantInteraction(intent, message).label).toBe(label);
  });

  it('uses calm system cues for offline and real errors', () => {
    expect(getSystemInteraction('offline').label).toContain('غير متاح');
    expect(getSystemInteraction('offline').tone).toBe('offline');
    expect(getSystemInteraction('error').label).toContain('خطأ');
    expect(getSystemInteraction('error').tone).toBe('error');
  });

  it('resolves direct navigation commands locally', () => {
    expect(getAssistantNavigationPath('افتح نقطة البيع')).toBe('/app/sales');
    expect(getAssistantNavigationPath('خذني للمخزون')).toBe('/app/inventory');
    expect(getAssistantNavigationPath('ابدأ عملية بيع')).toBe('/app/sales');
    expect(getAssistantNavigationPath('ابدا بعملية بيع')).toBe('/app/sales');
    expect(getAssistantNavigationPath('بيع جديد')).toBe('/app/sales');
  });
});
