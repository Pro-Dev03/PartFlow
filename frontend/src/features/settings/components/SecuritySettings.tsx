import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Lock, Save, Shield, Key } from 'lucide-react';

export function SecuritySettings() {
  const [securitySettings, setSecuritySettings] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: '',
    twoFactorEnabled: false,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Lock className="w-5 h-5 text-cyan" />
          إعدادات الأمان
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-md">
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">المصادقة الثنائية</h4>
            <p className="text-small text-text-secondary">تأكيد الهوية بخطوة إضافية</p>
          </div>
          <Button
            variant={securitySettings.twoFactorEnabled ? 'primary' : 'secondary'}
            onClick={() => setSecuritySettings({ ...securitySettings, twoFactorEnabled: !securitySettings.twoFactorEnabled })}
          >
            {securitySettings.twoFactorEnabled ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <div className="border-t border-border pt-4">
          <h4 className="font-medium text-text-primary mb-4 flex items-center gap-2">
            <Key className="w-4 h-4" />
            تغيير كلمة المرور
          </h4>
          <Input
            label="كلمة المرور الحالية"
            type="password"
            value={securitySettings.currentPassword}
            onChange={(e) => setSecuritySettings({ ...securitySettings, currentPassword: e.target.value })}
          />
          <Input
            label="كلمة المرور الجديدة"
            type="password"
            value={securitySettings.newPassword}
            onChange={(e) => setSecuritySettings({ ...securitySettings, newPassword: e.target.value })}
          />
          <Input
            label="تأكيد كلمة المرور"
            type="password"
            value={securitySettings.confirmPassword}
            onChange={(e) => setSecuritySettings({ ...securitySettings, confirmPassword: e.target.value })}
          />
        </div>
        <Button variant="primary" className="gap-2">
          <Save className="w-4 h-4" />
          حفظ التغييرات
        </Button>
      </CardContent>
    </Card>
  );
}
