import { describe, expect, it } from 'vitest';
import { formatExpenseDate, isExpenseInCurrentMonth, normalizeExpenseForDisplay } from './ExpensesPage';

describe('formatExpenseDate', () => {
  it('formats Gregorian dates with Latin numerals without timezone conversion', () => {
    expect(formatExpenseDate('2026-09-01T12:00:00Z')).toBe('01/09/2026');
  });

  it('rejects incomplete dates instead of guessing the day', () => {
    expect(formatExpenseDate('1/9')).toBe('-');
  });
});

describe('isExpenseInCurrentMonth', () => {
  it('matches the calendar month from the date value', () => {
    expect(isExpenseInCurrentMonth('2026-09-01T00:00:00Z', new Date(2026, 8, 2))).toBe(true);
    expect(isExpenseInCurrentMonth('2026-08-31T23:00:00Z', new Date(2026, 8, 2))).toBe(false);
  });
});

describe('normalizeExpenseForDisplay', () => {
  it('maps the backend expense shape into the fields used by the UI', () => {
    const item = {
      id: 'exp-1',
      title: 'إيجار الشهر',
      description: '',
      amount: 1500,
      currency: 'SAR',
      expense_date: '2026-08-01T00:00:00Z',
      category_name: 'الإيجار',
      payment_method: 'bank_transfer',
      status: 'approved',
      is_recurring: true,
      recurring_period: 'monthly',
    };

    const normalized = normalizeExpenseForDisplay(item);

    expect(normalized.description).toBe('إيجار الشهر');
    expect(normalized.category).toBe('الإيجار');
    expect(normalized.date).toBe('2026-08-01T00:00:00Z');
    expect(normalized.recurring).toBe(true);
    expect(normalized.recurringPeriod).toBe('monthly');
  });

  it('keeps legacy UI fields when they already exist', () => {
    const legacyItem = {
      id: 'exp-2',
      description: 'مستلزمات',
      category: 'supplies',
      date: '2026-08-02',
      amount: 250,
      recurring: false,
      recurringPeriod: 'monthly',
      receipt: 'legacy.pdf',
    };

    const normalized = normalizeExpenseForDisplay(legacyItem);

    expect(normalized.description).toBe('مستلزمات');
    expect(normalized.category).toBe('supplies');
    expect(normalized.date).toBe('2026-08-02');
    expect(normalized.recurring).toBe(false);
    expect(normalized.receipt).toBe('legacy.pdf');
  });
});
