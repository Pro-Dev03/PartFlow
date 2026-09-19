import { afterEach, describe, expect, it, vi } from 'vitest';
import { printHtmlDocument } from '../../services/documents/print-html';

describe('printHtmlDocument', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
    delete window.partflowDesktop;
    document.body.replaceChildren();
    document.head.querySelectorAll('[data-partflow-print]').forEach((style) => style.remove());
  });

  it('prints only the supplied HTML in the browser runtime', async () => {
    vi.useFakeTimers();
    const pageContent = document.createElement('main');
    pageContent.textContent = 'page content';
    document.body.appendChild(pageContent);
    const print = vi.spyOn(window, 'print').mockImplementation(() => undefined);

    await printHtmlDocument('<style>.invoice{color:red}</style><main class="invoice">invoice content</main>');

    expect(print).toHaveBeenCalledOnce();
    expect(document.querySelector('[id^="partflow-print-"]')).not.toBeNull();
    expect(document.head.querySelector('[data-partflow-print]')?.textContent).toContain('body > *:not(#partflow-print-');

    vi.advanceTimersByTime(1500);
    expect(document.querySelector('[id^="partflow-print-"]')).toBeNull();
    expect(document.body.textContent).toContain('page content');
  });

  it('delegates printing to Electron without calling browser print', async () => {
    const printHtml = vi.fn().mockResolvedValue(true);
    window.partflowDesktop = { appVersion: 'test', platform: 'win32', printHtml };
    const print = vi.spyOn(window, 'print').mockImplementation(() => undefined);

    await printHtmlDocument('<main>invoice content</main>');

    expect(printHtml).toHaveBeenCalledWith('<main>invoice content</main>');
    expect(print).not.toHaveBeenCalled();
  });
});