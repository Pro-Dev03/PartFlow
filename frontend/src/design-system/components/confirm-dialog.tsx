import { Modal } from './modal';
import { Button } from './button';
import { AlertTriangle, Info, CheckCircle, Trash2, X } from 'lucide-react';
import type { ReactNode } from 'react';

export interface ConfirmDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'danger' | 'warning' | 'info' | 'success';
  isLoading?: boolean;
  children?: ReactNode;
}

const variantIcons = {
  danger: Trash2,
  warning: AlertTriangle,
  info: Info,
  success: CheckCircle,
};

const variantStyles = {
  danger: {
    iconColor: 'var(--color-danger)',
    confirmVariant: 'danger' as const,
  },
  warning: {
    iconColor: 'var(--color-warning)',
    confirmVariant: 'primary' as const,
  },
  info: {
    iconColor: 'var(--color-info)',
    confirmVariant: 'primary' as const,
  },
  success: {
    iconColor: 'var(--color-success)',
    confirmVariant: 'primary' as const,
  },
};

export function ConfirmDialog({
  isOpen,
  onClose,
  onConfirm,
  title,
  message,
  confirmText = 'تأكيد',
  cancelText = 'إلغاء',
  variant = 'danger',
  isLoading = false,
  children,
}: ConfirmDialogProps) {
  const Icon = variantIcons[variant];
  const style = variantStyles[variant];
  const ActionIcon = variant === 'danger' ? Trash2 : CheckCircle;

  const handleConfirm = () => {
    onConfirm();
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={title}
      size="sm"
      variant="modern"
      showCloseButton={!isLoading}
      aria-describedby="confirm-dialog-message"
    >
      <div className="flex flex-col gap-5">
        <div
          className="relative overflow-hidden rounded-2xl border p-4"
          style={{
            borderColor: `${style.iconColor}35`,
            background: `linear-gradient(135deg, ${style.iconColor}12 0%, transparent 72%)`,
          }}
        >
          <div className="absolute -end-8 -top-8 h-24 w-24 rounded-full opacity-30" style={{ background: style.iconColor }} />
          <div className="relative flex items-start gap-3">
            <div
              className="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border"
              style={{
                background: `${style.iconColor}18`,
                borderColor: `${style.iconColor}30`,
              }}
            >
              <Icon className="h-5 w-5" style={{ color: style.iconColor }} aria-hidden="true" />
            </div>
            <div className="min-w-0 pt-0.5">
              <p className="text-sm font-bold" style={{ color: 'var(--text-primary)' }}>
                {variant === 'danger' ? 'تأكيد حذف البيانات' : title}
              </p>
              <p
                id="confirm-dialog-message"
                className="mt-1 text-sm leading-6"
                style={{ color: 'var(--text-secondary)' }}
              >
                {message}
              </p>
            </div>
          </div>
        </div>
        {children}

        <div className="flex flex-col-reverse gap-2 border-t border-[var(--border-subtle)] pt-4 sm:flex-row sm:justify-end">
          <Button
            variant="secondary"
            onClick={onClose}
            disabled={isLoading}
            className="sm:min-w-[108px]"
          >
            <X className="h-4 w-4" aria-hidden="true" />
            {cancelText}
          </Button>
          <Button
            variant={style.confirmVariant}
            onClick={handleConfirm}
            disabled={isLoading}
            loading={isLoading}
            className="sm:min-w-[132px]"
          >
            {!isLoading && <ActionIcon className="h-4 w-4" aria-hidden="true" />}
            {confirmText}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
