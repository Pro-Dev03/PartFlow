import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { NextFieldNavigator } from './NextFieldNavigator';

describe('NextFieldNavigator', () => {
  beforeEach(() => {
    vi.spyOn(HTMLElement.prototype, 'getClientRects').mockImplementation(
      () => [{ width: 100, height: 24, top: 0, left: 0, right: 100, bottom: 24, x: 0, y: 0, toJSON: () => ({}) }] as DOMRectList,
    );
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('shows a visible next button and stays within the current form', () => {
    render(
      <>
        <main>
          <input aria-label="Page search" />
          <form aria-label="Customer form">
            <input aria-label="Customer name" />
            <input aria-label="Customer phone" />
            <button type="submit">Save</button>
          </form>
        </main>
        <NextFieldNavigator />
      </>,
    );

    const name = screen.getByLabelText('Customer name');
    const phone = screen.getByLabelText('Customer phone');
    act(() => name.focus());

    const next = screen.getByRole('button', { name: /next field|الحقل التالي/i });
    expect(next).toBeVisible();
    fireEvent.click(next);

    expect(phone).toHaveFocus();
  });

  it('keeps modal field navigation inside the modal', () => {
    render(
      <>
        <main>
          <input aria-label="Background first" />
          <input aria-label="Background second" />
        </main>
        <section className="pf-modal-surface">
          <div className="pf-modal-navigation-slot" />
          <input aria-label="Modal first" />
          <input aria-label="Modal second" />
        </section>
        <NextFieldNavigator />
      </>,
    );

    const first = screen.getByLabelText('Modal first');
    const second = screen.getByLabelText('Modal second');
    const backgroundSecond = screen.getByLabelText('Background second');
    act(() => first.focus());

    fireEvent.click(screen.getByRole('button', { name: /next field|الحقل التالي/i }));

    expect(second).toHaveFocus();
    expect(backgroundSecond).not.toHaveFocus();
  });

  it('does not show the next button inside excluded forms', () => {
    render(
      <>
        <form data-next-disabled>
          <input aria-label="Login email" />
          <input aria-label="Login password" />
        </form>
        <NextFieldNavigator />
      </>,
    );

    act(() => screen.getByLabelText('Login email').focus());

    expect(screen.queryByRole('button', { name: /next field|الحقل التالي/i })).not.toBeInTheDocument();
  });
});
