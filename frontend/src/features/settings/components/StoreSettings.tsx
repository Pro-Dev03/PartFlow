import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Store, Save } from 'lucide-react';
import { settingsApi } from '../../../services/api/endpoints';
import { toast } from 'sonner';

export function StoreSettings() {
  const queryClient = useQueryClient();
  const [storeSettings, setStoreSettings] = useState({
    storeName: 'PartFlow Store',
  });
  const { data: storeNameSetting, isLoading } = useQuery({
    queryKey: ['settings', 'store_name'],
    queryFn: () => settingsApi.getSetting('store_name'),
  });
  const updateStoreNameMutation = useMutation({
    mutationFn: (storeName: string) => settingsApi.updateSetting('store_name', storeName),
    onSuccess: () => {
      localStorage.setItem('partflow-store-name', storeSettings.storeName.trim());
      void queryClient.invalidateQueries({ queryKey: ['settings', 'store_name'] });
      void queryClient.invalidateQueries({ queryKey: ['settings', 'public'] });
      toast.success('تم حفظ اسم المتجر بنجاح');
    },
    onError: (error: any) => {
      toast.error(error?.message || 'تعذر حفظ اسم المتجر');
    },
  });

  useEffect(() => {
    const value = storeNameSetting?.data?.value;
    if (typeof value === 'string' && value.trim()) {
      setStoreSettings({ storeName: value });
      localStorage.setItem('partflow-store-name', value.trim());
    }
  }, [storeNameSetting]);

  const handleSave = () => {
    const storeName = storeSettings.storeName.trim();
    if (!storeName) {
      toast.error('يرجى إدخال اسم المتجر');
      return;
    }
    updateStoreNameMutation.mutate(storeName);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Store className="w-5 h-5 text-cyan" />
          إعدادات المتجر
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-md">
        <Input
          label="اسم المتجر"
          value={storeSettings.storeName}
          disabled={isLoading || updateStoreNameMutation.isPending}
          onChange={(e) => setStoreSettings({ ...storeSettings, storeName: e.target.value })}
        />
        <Button variant="primary" className="gap-2" onClick={handleSave} disabled={isLoading || updateStoreNameMutation.isPending}>
          <Save className="w-4 h-4" />
          {updateStoreNameMutation.isPending ? 'جارٍ الحفظ...' : 'حفظ التغييرات'}
        </Button>
      </CardContent>
    </Card>
  );
}