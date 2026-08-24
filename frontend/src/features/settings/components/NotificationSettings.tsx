import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Bell, Save } from 'lucide-react';

export function NotificationSettings() {
  const [notificationSettings, setNotificationSettings] = useState({
    lowStockAlerts: true,
    debtReminders: true,
    warrantyExpiry: true,
    emailNotifications: false,
    emailAddress: '',
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Bell className="w-5 h-5 text-cyan" />
          إعدادات الإشعارات
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-md">
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">تنبيهات المخزون المنخفض</h4>
            <p className="text-small text-text-secondary">إشعار عند وصول المخزون للحد الأدنى</p>
          </div>
          <Button
            variant={notificationSettings.lowStockAlerts ? 'primary' : 'secondary'}
            onClick={() => setNotificationSettings({ ...notificationSettings, lowStockAlerts: !notificationSettings.lowStockAlerts })}
          >
            {notificationSettings.lowStockAlerts ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">تذكيرات الديون</h4>
            <p className="text-small text-text-secondary">إشعار بتأخر سداد الديون</p>
          </div>
          <Button
            variant={notificationSettings.debtReminders ? 'primary' : 'secondary'}
            onClick={() => setNotificationSettings({ ...notificationSettings, debtReminders: !notificationSettings.debtReminders })}
          >
            {notificationSettings.debtReminders ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">تنبيهات انتهاء الضمان</h4>
            <p className="text-small text-text-secondary">إشعار قبل انتهاء الضمان</p>
          </div>
          <Button
            variant={notificationSettings.warrantyExpiry ? 'primary' : 'secondary'}
            onClick={() => setNotificationSettings({ ...notificationSettings, warrantyExpiry: !notificationSettings.warrantyExpiry })}
          >
            {notificationSettings.warrantyExpiry ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">الإشعارات عبر البريد</h4>
            <p className="text-small text-text-secondary">إرسال الإشعارات عبر البريد الإلكتروني</p>
          </div>
          <Button
            variant={notificationSettings.emailNotifications ? 'primary' : 'secondary'}
            onClick={() => setNotificationSettings({ ...notificationSettings, emailNotifications: !notificationSettings.emailNotifications })}
          >
            {notificationSettings.emailNotifications ? 'مفعّل' : 'معطّل'}
          </Button>
        </div>
        {notificationSettings.emailNotifications && (
          <Input
            label="البريد الإلكتروني"
            type="email"
            value={notificationSettings.emailAddress}
            onChange={(e) => setNotificationSettings({ ...notificationSettings, emailAddress: e.target.value })}
          />
        )}
        <Button variant="primary" className="gap-2">
          <Save className="w-4 h-4" />
          حفظ التغييرات
        </Button>
      </CardContent>
    </Card>
  );
}
