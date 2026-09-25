import { createPortal } from 'react-dom';
import { useEffect, useRef, useState } from 'react';
import { MoreHorizontal } from 'lucide-react';
import { Button } from './button';
import { cn } from '../../utils';

export interface ActionMenuItem {
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  onClick: () => void;
  danger?: boolean;
}

interface ActionMenuProps {
  items: ActionMenuItem[];
  label?: string;
  widthClassName?: string;
}

export function ActionMenu({ items, label = 'خيارات', widthClassName = 'w-44' }: ActionMenuProps) {
  const [open, setOpen] = useState(false);
  const [position, setPosition] = useState({ top: 0, left: 0, openBelow: false });
  const triggerRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  const normalizedWidthClass = widthClassName === 'w-48' ? 'w-48' : 'w-44';

  useEffect(() => {
    if (!open) return;

    const closeOnOutsideClick = (event: PointerEvent) => {
      if (!triggerRef.current?.contains(event.target as Node) && !menuRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setOpen(false);
    };

    document.addEventListener('pointerdown', closeOnOutsideClick);
    document.addEventListener('keydown', closeOnEscape);
    return () => {
      document.removeEventListener('pointerdown', closeOnOutsideClick);
      document.removeEventListener('keydown', closeOnEscape);
    };
  }, [open]);

  const toggleMenu = () => {
    if (!triggerRef.current) return;
    const rect = triggerRef.current.getBoundingClientRect();
    const menuHeight = Math.min(items.length * 40 + 8, 280);
    const menuWidth = 176;
    const maxLeft = Math.max(8, window.innerWidth - menuWidth - 8);
    const clampLeft = (left: number) => Math.max(8, Math.min(left, maxLeft));
    let openBelow = rect.top < menuHeight + 16;
    let top = openBelow ? rect.bottom + 8 : rect.top - 8;
    let left = clampLeft(rect.left);

    const assistant = document.querySelector<HTMLElement>('[aria-label="فتح مساعد PartFlow"]')?.getBoundingClientRect();
    if (assistant) {
      const overlapsAssistant = (candidateLeft: number, candidateTop: number, candidateBelow: boolean) => {
        const candidateMenuTop = candidateBelow ? candidateTop : candidateTop - menuHeight;
        return candidateLeft < assistant.right
          && candidateLeft + menuWidth > assistant.left
          && candidateMenuTop < assistant.bottom
          && candidateMenuTop + menuHeight > assistant.top;
      };

      if (overlapsAssistant(left, top, openBelow)) {
        const alternateLefts = [
          assistant.right + 8,
          assistant.left - menuWidth - 8,
        ].map(clampLeft);
        const clearLeft = alternateLefts.find((candidateLeft) =>
          !overlapsAssistant(candidateLeft, top, openBelow)
        );

        if (clearLeft !== undefined) {
          left = clearLeft;
        } else {
          const aboveTop = assistant.top - 8;
          const belowTop = assistant.bottom + 8;
          if (aboveTop - menuHeight >= 8) {
            openBelow = false;
            top = aboveTop;
          } else if (belowTop + menuHeight <= window.innerHeight - 8) {
            openBelow = true;
            top = belowTop;
          }
        }
      }
    }

    setPosition({
      top,
      left,
      openBelow,
    });
    setOpen((current) => !current);
  };

  return (
    <>
      <Button
        ref={triggerRef}
        type="button"
        variant="ghost"
        size="icon"
        onClick={(event) => {
          event.stopPropagation();
          toggleMenu();
        }}
        className="pf-action-menu-trigger text-text-secondary hover:text-text-primary"
        aria-label={label}
        title={label}
        aria-expanded={open}
        aria-haspopup="menu"
      >
        <MoreHorizontal className="h-4 w-4" />
      </Button>

      {open && typeof document !== 'undefined' && createPortal(
        <div
          ref={menuRef}
          role="menu"
          className={cn(
            'pf-action-menu z-[10000]',
            'animate-in fade-in-0 zoom-in-95 duration-150',
            normalizedWidthClass
          )}
          style={{
            position: 'fixed',
            top: position.top,
            left: position.left,
            width: normalizedWidthClass === 'w-48' ? 192 : 176,
            transform: position.openBelow ? undefined : 'translateY(-100%)',
          }}
        >
          <div className="pf-action-menu-title">
            {label}
          </div>
          {items.map(({ label: itemLabel, icon: Icon, onClick, danger }) => (
            <button
              key={itemLabel}
              type="button"
              role="menuitem"
              onClick={(event) => {
                event.stopPropagation();
                onClick();
                setOpen(false);
              }}
              className={cn(
                'pf-action-menu-item group',
                danger && 'pf-action-menu-item-danger'
              )}
            >
              <span className={cn('pf-action-menu-icon', danger && 'pf-action-menu-icon-danger')}>
                <Icon className="h-4 w-4" />
              </span>
              <span>{itemLabel}</span>
            </button>
          ))}
        </div>,
        document.body
      )}
    </>
  );
}
