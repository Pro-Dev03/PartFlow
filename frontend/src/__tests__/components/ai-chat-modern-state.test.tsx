import { act, fireEvent, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from '../../test/utils';
import AIChatModern from '../../components/ui/ai-chat-modern';
import { assistantApi } from '../../services/api/endpoints';

vi.mock('../../services/api/endpoints', () => ({
  assistantApi: {
    reply: vi.fn(),
  },
}));

describe('AIChatModern assistant states', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.mocked(assistantApi.reply).mockReset();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('moves from listening to thinking to speaking and then idle', async () => {
    vi.mocked(assistantApi.reply).mockResolvedValue({
      data: { reply: 'مبيعات اليوم وصلت لـ ₪850.', intent: 'SALES' },
    } as never);
    const states: string[] = [];

    render(<AIChatModern onStateChange={(state) => states.push(state)} />);
    const input = screen.getByPlaceholderText('اكتب رسالتك...');

    await act(async () => {
      fireEvent.change(input, { target: { value: 'كم مبيعات اليوم؟' } });
    });
    expect(states).toContain('listening');

    await act(async () => {
      fireEvent.keyDown(input, { key: 'Enter', code: 'Enter', charCode: 13 });
    });
    expect(states).toContain('thinking');

    await act(async () => {
      await Promise.resolve();
    });
    expect(states).toContain('speaking');

    await act(async () => {
      vi.advanceTimersByTime(1800);
    });
    expect(states).toContain('idle');
    expect(states).toEqual(expect.arrayContaining(['listening', 'thinking', 'speaking', 'idle']));
  });

  it('shows offline state and does not call the backend', () => {
    const states: string[] = [];
    render(<AIChatModern isOffline onStateChange={(state) => states.push(state)} />);
    const input = screen.getByPlaceholderText('اكتب رسالتك...');

    fireEvent.change(input, { target: { value: 'كم المخزون؟' } });
    fireEvent.keyDown(input, { key: 'Enter', code: 'Enter', charCode: 13 });

    expect(screen.getByText(/أوفلاين/)).toBeInTheDocument();
    expect(assistantApi.reply).not.toHaveBeenCalled();
    expect(states).toContain('offline');
  });

  it('shows a contextual cue while reviewing sales', async () => {
    vi.mocked(assistantApi.reply).mockResolvedValue({
      data: { reply: 'مبيعات اليوم وصلت لـ ₪850.', intent: 'SALES' },
    } as never);

    render(<AIChatModern />);
    const input = screen.getByPlaceholderText('اكتب رسالتك...');
    fireEvent.change(input, { target: { value: 'كم مبيعات اليوم؟' } });
    await act(async () => {
      fireEvent.keyDown(input, { key: 'Enter', code: 'Enter', charCode: 13 });
      await Promise.resolve();
    });

    expect(screen.getByRole('status')).toHaveTextContent('أراجع مبيعات اليوم...');
  });

  it('shows an error cue only after the backend fails', async () => {
    vi.mocked(assistantApi.reply).mockRejectedValue(new Error('network')); 

    render(<AIChatModern />);
    const input = screen.getByPlaceholderText('اكتب رسالتك...');
    fireEvent.change(input, { target: { value: 'كم مبيعات اليوم؟' } });
    await act(async () => {
      fireEvent.keyDown(input, { key: 'Enter', code: 'Enter', charCode: 13 });
      await Promise.resolve();
    });

    expect(screen.getByRole('status')).toHaveTextContent('صار خطأ وأنا أراجع البيانات');
  });
});
