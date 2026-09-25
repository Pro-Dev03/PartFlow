import { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { ArrowLeft, ArrowRight } from 'lucide-react';

type DataField = HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement | HTMLElement;

const DATA_FIELD_SELECTOR = [
  'input:not([type="hidden"]):not([type="button"]):not([type="submit"]):not([type="reset"]):not([type="image"]):not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[contenteditable="true"]:not([aria-disabled="true"])',
].join(',');
const NEXT_ACTION_SELECTOR = 'button, input[type="submit"], [data-next-action]';
const NEXT_ACTION_LABEL = /(حفظ|تأكيد|إضافة|إنشاء|تسجيل|اعتماد|متابعة|تطبيق|بحث|save|submit|add|create|record|confirm|continue|apply|search|next)/i;
const UNSAFE_ACTION_LABEL = /(إلغاء|حذف|إغلاق|رجوع|cancel|delete|close|back|remove|clear)/i;

function isDataField(element: Element | null): element is DataField {
  return Boolean(
    element instanceof HTMLElement
    && element.matches(DATA_FIELD_SELECTOR)
    && !element.matches(':disabled, [readonly], [aria-disabled="true"], [data-next-skip]')
    && !element.closest('[data-next-disabled]')
    && element.getClientRects().length > 0
  );
}

function isSafeNextAction(action: HTMLElement, current: DataField): boolean {
  if (action.matches(':disabled, [aria-disabled="true"], [data-next-skip]') || action.getClientRects().length === 0) return false;
  if (!(current.compareDocumentPosition(action) & Node.DOCUMENT_POSITION_FOLLOWING)) return false;

  const label = [
    action.textContent,
    action.getAttribute('aria-label'),
    action.getAttribute('title'),
    action instanceof HTMLInputElement ? action.value : '',
  ].join(' ').trim();
  return action.hasAttribute('data-next-action') || (NEXT_ACTION_LABEL.test(label) && !UNSAFE_ACTION_LABEL.test(label));
}

function getNextTarget(current: DataField): HTMLElement | null {
  // Keep field navigation inside the nearest data-entry area. A form takes
  // precedence over the page so its inputs never jump into unrelated filters
  // or search fields elsewhere on the screen.
  const scope = current.closest('[data-next-scope]')
    ?? current.closest('form')
    ?? current.closest('[role="dialog"], dialog, .pf-modal-surface')
    ?? current.closest('main')
    ?? document;
  const fields = Array.from(scope.querySelectorAll(DATA_FIELD_SELECTOR)).filter(isDataField);
  const currentIndex = fields.indexOf(current);
  if (currentIndex >= 0 && currentIndex + 1 < fields.length) return fields[currentIndex + 1];

  const form = current.closest('form');
  const actionScope = form ?? scope;
  return Array.from(actionScope.querySelectorAll<HTMLElement>(NEXT_ACTION_SELECTOR))
    .find((action) => isSafeNextAction(action, current)) ?? null;
}

function focusField(field: DataField): void {
  field.focus();
  if (field instanceof HTMLInputElement && !['checkbox', 'radio', 'file', 'color', 'range'].includes(field.type)) {
    try {
      field.select();
    } catch {
      // Some native input types accept focus but do not support text selection.
    }
  }
}

function focusNextTarget(current: DataField): void {
  const next = getNextTarget(current);
  if (!next) return;
  focusField(next as DataField);
}

export function NextFieldNavigator() {
  const [activeField, setActiveField] = useState<DataField | null>(null);
  const activeFieldRef = useRef<DataField | null>(null);
  const nextButtonRef = useRef<HTMLButtonElement | null>(null);

  useEffect(() => {
    const updateActiveField = () => {
      // Keep the source field while the user activates this floating button.
      // Otherwise focusing the button clears the field before its click handler runs.
      if (document.activeElement === nextButtonRef.current) return;
      const focused = document.activeElement;
      const nextActive = isDataField(focused) && getNextTarget(focused) ? focused : null;
      activeFieldRef.current = nextActive;
      setActiveField(nextActive);
    };

    const onFocusOut = (event: FocusEvent) => {
      if (event.relatedTarget === nextButtonRef.current) return;
      window.setTimeout(updateActiveField, 0);
    };
    document.addEventListener('focusin', updateActiveField);
    document.addEventListener('focusout', onFocusOut);
    let updateFrame = 0;
    const observer = new MutationObserver(() => {
      if (!activeFieldRef.current || updateFrame) return;
      updateFrame = window.requestAnimationFrame(() => {
        updateFrame = 0;
        updateActiveField();
      });
    });
    observer.observe(document.body, { childList: true, subtree: true, attributes: true, attributeFilter: ['disabled', 'hidden', 'readonly', 'aria-hidden'] });
    updateActiveField();

    return () => {
      document.removeEventListener('focusin', updateActiveField);
      document.removeEventListener('focusout', onFocusOut);
      observer.disconnect();
      if (updateFrame) window.cancelAnimationFrame(updateFrame);
    };
  }, []);

  if (!activeField || typeof document === 'undefined') return null;

  const modalSurface = activeField.closest<HTMLElement>('.pf-modal-surface')
    ?? activeField.closest<HTMLElement>('[role="dialog"], dialog');
  const modalNavigationSlot = modalSurface?.querySelector<HTMLElement>('.pf-modal-navigation-slot');
  const destination = modalNavigationSlot ?? modalSurface ?? document.body;
  const isEnglish = document.documentElement.lang.toLowerCase().startsWith('en');
  const label = isEnglish ? 'Next' : 'التالي';
  const accessibleLabel = isEnglish ? 'Move to next field' : 'انتقل إلى الحقل التالي';
  const DirectionIcon = document.documentElement.dir === 'rtl' ? ArrowLeft : ArrowRight;

  return createPortal(
    <button
      ref={nextButtonRef}
      type="button"
      onPointerDown={(event) => event.preventDefault()}
      onClick={() => {
        const current = activeFieldRef.current;
        if (current) focusNextTarget(current);
      }}
      aria-label={accessibleLabel}
      title={accessibleLabel}
      className="inline-flex min-h-10 items-center justify-center gap-2 rounded-xl border border-cyan-700 bg-cyan-700 px-4 text-sm font-semibold text-white shadow-lg transition hover:bg-cyan-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-400 focus-visible:ring-offset-2"
      style={modalNavigationSlot
        ? { position: 'static', zIndex: 10020, marginInlineStart: 'auto', display: 'flex', width: 'fit-content' }
        : modalSurface
        ? { position: 'sticky', bottom: 8, zIndex: 10020, marginInlineStart: 'auto', display: 'flex', width: 'fit-content' }
        : { position: 'fixed', bottom: 'max(5.5rem, env(safe-area-inset-bottom))', left: '50%', transform: 'translateX(-50%)', zIndex: 10020 }
      }
    >
      {label}
      <DirectionIcon className="h-4 w-4" aria-hidden="true" />
    </button>,
    destination,
  );
}
