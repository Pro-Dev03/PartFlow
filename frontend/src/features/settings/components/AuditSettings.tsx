import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Download, FileText, ListChecks } from 'lucide-react';
import { toast } from 'sonner';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Button } from '../../../design-system/components/button';
import { auditApi } from '../../../services/api/endpoints';

export function AuditSettings() {
  const navigate = useNavigate();
  const [isExporting, setIsExporting] = useState(false);

  const exportAuditLog = async () => {
    setIsExporting(true);
    try {
      const blob = await auditApi.exportCsv();
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = url;
      anchor.download = `partflow-audit-${new Date().toISOString().slice(0, 10)}.csv`;
      document.body.appendChild(anchor);
      anchor.click();
      anchor.remove();
      URL.revokeObjectURL(url);
      toast.success('تم تصدير سجل التدقيق بنجاح');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'تعذر تصدير سجل التدقيق');
    } finally {
      setIsExporting(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <FileText className="h-5 w-5 text-cyan" />
          سجل التدقيق
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-sm text-text-secondary">
          يعرض سجل التدقيق العمليات المسجلة فعليًا في قاعدة البيانات ويمكن تصديره بصيغة CSV للمراجعة والأرشفة.
        </p>
        <div className="flex flex-wrap gap-2">
          <Button variant="primary" className="gap-2" onClick={() => navigate('/app/audit')}>
            <ListChecks className="h-4 w-4" />
            فتح سجل التدقيق
          </Button>
          <Button variant="secondary" className="gap-2" onClick={() => void exportAuditLog()} disabled={isExporting}>
            <Download className="h-4 w-4" />
            {isExporting ? 'جارِ التصدير...' : 'تصدير السجل'}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
