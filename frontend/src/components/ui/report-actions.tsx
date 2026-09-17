import { ActionMenu } from './action-menu';
import { Download, Printer } from 'lucide-react';

interface ReportActionsProps {
  onExportCurrent: () => void;
  onPrintCurrent: () => void;
  onExportAll?: () => void;
  onPrintAll?: () => void;
  disabled?: boolean;
  loading?: boolean;
}

export function ReportActions({
  onExportCurrent,
  onPrintCurrent,
  onExportAll = onExportCurrent,
  onPrintAll = onPrintCurrent,
  disabled = false,
  loading = false,
}: ReportActionsProps) {
  const isDisabled = disabled || loading;

  return (
    <div className={isDisabled ? 'pointer-events-none opacity-60' : undefined} aria-busy={loading || undefined}>
      <ActionMenu
        label="تقرير البيانات"
        widthClassName="w-48"
        items={[
          { label: 'تصدير الصفحة الحالية', icon: Download, onClick: onExportCurrent },
          { label: 'تصدير كل النتائج', icon: Download, onClick: onExportAll },
          { label: 'طباعة الصفحة الحالية', icon: Printer, onClick: onPrintCurrent },
          { label: 'طباعة كل النتائج', icon: Printer, onClick: onPrintAll },
        ]}
      />
    </div>
  );
}
