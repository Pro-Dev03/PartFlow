import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { FileText, Save, Download } from 'lucide-react';

export function AuditSettings() {
  const [auditSettings, setAuditSettings] = useState({
    logRetentionDays: 90,
    logAllActions: true,
    logFailedAttempts: true,
    autoExportLogs: false,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <FileText className="w-5 h-5 text-cyan" />
          إعدادات التدقيق
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-md">
        <Input
          label="فترة الاحتفاظ بالسجلات (أيام)"
          type="number"
          value={auditSettings.logRetentionDays}
          onChange={(e) => setAuditSettings({ ...auditSettings, logRetentionDays: Number(e.target.value) })}
        />
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">تسجيل جميع الإجراءات</h4>
            <p className="text-small text-text-secondary">تسجيل كل نشاط في النظام</p>
          </div>
          <Button
            variant={auditSettings.logAllActions ? 'primary' : 'secondary'}
            onClick={() => setAuditSettings({ ...auditSettings, logAllActions: !auditSettings.logAllActions })}
          >
            {auditSettings.logAllActions ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">تسجيل المحاولات الفاشلة</h4>
            <p className="text-small text-text-secondary">تسجيل محاولات الدخول الفاشلة</p>
          </div>
          <Button
            variant={auditSettings.logFailedAttempts ? 'primary' : 'secondary'}
            onClick={() => setAuditSettings({ ...auditSettings, logFailedAttempts: !auditSettings.logFailedAttempts })}
          >
            {auditSettings.logFailedAttempts ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">تصدير تلقائي للسجلات</h4>
            <p className="text-small text-text-secondary">تصدير السجلات بشكل دوري</p>
          </div>
          <Button
            variant={auditSettings.autoExportLogs ? 'primary' : 'secondary'}
            onClick={() => setAuditSettings({ ...auditSettings, autoExportLogs: !auditSettings.autoExportLogs })}
          >
            {auditSettings.autoExportLogs ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <div className="flex gap-2">
          <Button variant="primary" className="gap-2 flex-1">
            <Save className="w-4 h-4" />
            حفظ التغييرات
          </Button>
          <Button variant="secondary" className="gap-2">
            <Download className="w-4 h-4" />
            تصدير السجلات
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
