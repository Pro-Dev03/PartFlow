import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Select } from '../../../components/ui/select';
import { Palette, Save, Moon, Sun } from 'lucide-react';

export function AppearanceSettings() {
  const [appearanceSettings, setAppearanceSettings] = useState({
    theme: 'dark',
    language: 'ar',
    fontSize: 'medium',
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Palette className="w-5 h-5 text-cyan" />
          إعدادات المظهر
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-md">
        <div className="flex items-center justify-between p-4 border border-border rounded-lg">
          <div>
            <h4 className="font-medium text-text-primary">الوضع الليلي</h4>
            <p className="text-small text-text-secondary">تغيير مظهر التطبيق</p>
          </div>
          <div className="flex gap-2">
            <Button
              variant={appearanceSettings.theme === 'dark' ? 'primary' : 'secondary'}
              onClick={() => setAppearanceSettings({ ...appearanceSettings, theme: 'dark' })}
              className="gap-2"
            >
              <Moon className="w-4 h-4" />
              داكن
            </Button>
            <Button
              variant={appearanceSettings.theme === 'light' ? 'primary' : 'secondary'}
              onClick={() => setAppearanceSettings({ ...appearanceSettings, theme: 'light' })}
              className="gap-2"
            >
              <Sun className="w-4 h-4" />
              فاتح
            </Button>
          </div>
        </div>
        <div>
          <label className="block text-sm font-medium text-text-primary mb-2">اللغة</label>
          <Select
            value={appearanceSettings.language}
            onChange={(e) => setAppearanceSettings({ ...appearanceSettings, language: e.target.value })}
            options={[
              { value: 'ar', label: 'العربية' },
              { value: 'en', label: 'English' },
            ]}
            emptyMessage="لا توجد لغات"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-text-primary mb-2">حجم الخط</label>
          <Select
            value={appearanceSettings.fontSize}
            onChange={(e) => setAppearanceSettings({ ...appearanceSettings, fontSize: e.target.value })}
            options={[
              { value: 'small', label: 'صغير' },
              { value: 'medium', label: 'متوسط' },
              { value: 'large', label: 'كبير' },
            ]}
            emptyMessage="لا توجد أحجام"
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
