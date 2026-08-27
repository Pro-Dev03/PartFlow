import { Modal } from './modal';
import { Button } from './button';
import { AlertTriangle, Info, CheckCircle } from 'lucide-react';
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
  danger: AlertTriangle,
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
    >
      <div className="flex flex-col gap-4">
        {/* Icon and Message */}
        <div className="flex items-start gap-3">
          <div
            className="flex-shrink-0 w-10 h-10 rounded-lg flex items-center justify-center"
            style={{ background: `${style.iconColor}15` }}
          >
            <Icon className="w-5 h-5" style={{ color: style.iconColor }} />
          </div>
          <p className="text-sm leading-relaxed" style={{ color: 'var(--text-secondary)' }}>
            {message}
          </p>
        </div>
        {children}

        {/* Actions */}
        <div className="flex justify-end gap-3 pt-2">
          <Button
            variant="secondary"
            onClick={onClose}
            disabled={isLoading}
          >
            {cancelText}
          </Button>
          <Button
            variant={style.confirmVariant}
            onClick={handleConfirm}
            disabled={isLoading}
            loading={isLoading}
          >
            {confirmText}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
