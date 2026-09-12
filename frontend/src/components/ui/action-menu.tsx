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
    const openBelow = rect.top < menuHeight + 16;
    const menuWidth = 176;
    setPosition({
      top: openBelow ? rect.bottom + 8 : rect.top - 8,
      left: Math.max(8, Math.min(rect.left, window.innerWidth - menuWidth - 8)),
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
            'pf-action-menu z-[100]',
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
