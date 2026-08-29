import { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  CalendarClock,
  CheckCircle2,
  Clock3,
  Search,
  ShieldAlert,
  ShieldCheck,
  Users,
  XCircle,
} from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '../../../components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { settingsApi } from '../../../services/api/endpoints';
import type { User } from '../../../types/models';

const formatExpiry = (value?: string | null) => {
  if (!value) return 'غير محدد';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'غير محدد';
  return new Intl.DateTimeFormat('ar-EG', { dateStyle: 'medium' }).format(date);
};

const getDaysRemaining = (value?: string | null) => {
  if (!value) return null;
  const expiry = new Date(value);
  if (Number.isNaN(expiry.getTime())) return null;
  const diffMs = expiry.getTime() - Date.now();
  return Math.max(0, Math.ceil(diffMs / (1000 * 60 * 60 * 24)));
};

const statusStyles: Record<string, string> = {
  active: 'bg-emerald-100 text-emerald-700 border border-emerald-200',
  expired: 'bg-red-100 text-red-700 border border-red-200',
  canceled: 'bg-orange-100 text-orange-700 border border-orange-200',
  cancelled: 'bg-orange-100 text-orange-700 border border-orange-200',
  trial: 'bg-sky-100 text-sky-700 border border-sky-200',
};

export function SubscriptionManagement() {
  const queryClient = useQueryClient();
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'expired' | 'canceled'>('all');

  const { data: summaryData } = useQuery({
    queryKey: ['subscription-summary'],
    queryFn: async () => {
      const response = await settingsApi.getSubscriptionSummary();
      return (response.data as Record<string, number> | undefined) ?? {};
    },
  });

  const { data: subscribersData, isLoading } = useQuery({
    queryKey: ['subscription-subscribers'],
    queryFn: async () => {
      const response = await settingsApi.getSubscribers({ page: 1, per_page: 100 });
      return ((response.data ?? []) as User[]);
    },
  });

  const subscribers = useMemo(() => {
    const list = Array.isArray(subscribersData) ? subscribersData : [];
    const normalized = list.filter((user) => {
      const matchesSearch = !search || `${user.first_name ?? ''} ${user.last_name ?? ''} ${user.email}`.toLowerCase().includes(search.toLowerCase());
      const status = (user.subscription_status ?? 'active').toLowerCase();
      const matchesStatus = statusFilter === 'all' || status === statusFilter;
      return matchesSearch && matchesStatus;
    });

    return normalized.sort((a, b) => {
      const aDate = a.subscription_expires_at ? new Date(a.subscription_expires_at).getTime() : Number.MAX_SAFE_INTEGER;
      const bDate = b.subscription_expires_at ? new Date(b.subscription_expires_at).getTime() : Number.MAX_SAFE_INTEGER;
      return aDate - bDate;
    });
  }, [search, statusFilter, subscribersData]);

  const renewMutation = useMutation({
    mutationFn: async ({ id, days }: { id: string; days: number }) => {
      const response = await settingsApi.renewSubscription(id, days);
      return response.data as User;
    },
    onSuccess: () => {
      toast.success('تم تجديد الاشتراك بنجاح');
      queryClient.invalidateQueries({ queryKey: ['subscription-subscribers'] });
      queryClient.invalidateQueries({ queryKey: ['subscription-summary'] });
    },
    onError: (error: any) => {
      toast.error(error?.message || 'تعذر تجديد الاشتراك');
    },
  });

  const updateStatusMutation = useMutation({
    mutationFn: async ({ id, status, expiresAt }: { id: string; status: string; expiresAt?: string | null }) => {
      const response = await settingsApi.updateSubscriptionStatus(id, {
        subscription_status: status,
        subscription_expires_at: expiresAt ?? null,
      });
      return response.data as User;
    },
    onSuccess: () => {
      toast.success('تم تحديث حالة الاشتراك');
      queryClient.invalidateQueries({ queryKey: ['subscription-subscribers'] });
      queryClient.invalidateQueries({ queryKey: ['subscription-summary'] });
    },
    onError: (error: any) => {
      toast.error(error?.message || 'تعذر تحديث الاشتراك');
    },
  });

  const summary = {
    total: summaryData?.total ?? subscribers.length,
    active: summaryData?.active ?? 0,
    expired: summaryData?.expired ?? 0,
    canceled: summaryData?.canceled ?? 0,
  };

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground">إجمالي الحسابات</p>
            <div className="mt-2 flex items-center justify-between">
              <span className="text-2xl font-bold">{summary.total}</span>
              <Users className="h-5 w-5 text-primary" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground">نشط</p>
            <div className="mt-2 flex items-center justify-between">
              <span className="text-2xl font-bold text-emerald-600">{summary.active}</span>
              <ShieldCheck className="h-5 w-5 text-emerald-600" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground">منتهي</p>
            <div className="mt-2 flex items-center justify-between">
              <span className="text-2xl font-bold text-red-600">{summary.expired}</span>
              <Clock3 className="h-5 w-5 text-red-600" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground">موقوف</p>
            <div className="mt-2 flex items-center justify-between">
              <span className="text-2xl font-bold text-orange-600">{summary.canceled}</span>
              <ShieldAlert className="h-5 w-5 text-orange-600" />
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between gap-3">
            <span className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              إدارة المشتركين
            </span>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <div className="relative min-w-0 flex-1">
              <Search className="absolute right-3 top-3.5 h-4 w-4 text-muted-foreground" />
              <input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="ابحث باسم المستخدم أو البريد..."
                className="w-full rounded-xl border border-border bg-background px-10 py-2.5 text-sm outline-none ring-0 placeholder:text-muted-foreground"
              />
            </div>
            <div className="flex flex-wrap gap-2">
              {(['all', 'active', 'expired', 'canceled'] as const).map((value) => (
                <Button
                  key={value}
                  size="sm"
                  variant={statusFilter === value ? 'primary' : 'secondary'}
                  onClick={() => setStatusFilter(value)}
                >
                  {value === 'all' ? 'الكل' : value === 'active' ? 'نشط' : value === 'expired' ? 'منتهي' : 'موقوف'}
                </Button>
              ))}
            </div>
          </div>

          {isLoading ? (
            <div className="flex items-center justify-center rounded-xl border border-dashed border-border py-10 text-sm text-muted-foreground">
              جاري تحميل المشتركين...
            </div>
          ) : subscribers.length === 0 ? (
            <div className="flex items-center justify-center rounded-xl border border-dashed border-border py-10 text-sm text-muted-foreground">
              لا توجد حسابات مطابقة للبحث الحالي.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full text-right text-sm">
                <thead className="text-muted-foreground">
                  <tr className="border-b border-border">
                    <th className="px-3 py-3 font-medium">المشترك</th>
                    <th className="px-3 py-3 font-medium">حالة الاشتراك</th>
                    <th className="px-3 py-3 font-medium">تاريخ الانتهاء</th>
                    <th className="px-3 py-3 font-medium">الأيام المتبقية</th>
                    <th className="px-3 py-3 font-medium">إجراءات</th>
                  </tr>
                </thead>
                <tbody>
                  {subscribers.map((user) => {
                    const status = (user.subscription_status ?? 'active').toLowerCase();
                    const remainingDays = getDaysRemaining(user.subscription_expires_at ?? undefined);
                    const isExpired = status === 'expired' || (remainingDays === 0 && !!user.subscription_expires_at);
                    return (
                      <tr key={user.id} className="border-b border-border/70 align-middle">
                        <td className="px-3 py-3">
                          <div>
                            <div className="font-semibold text-foreground">
                              {user.first_name || user.name || 'بدون اسم'} {user.last_name || ''}
                            </div>
                            <div className="text-xs text-muted-foreground">{user.email}</div>
                          </div>
                        </td>
                        <td className="px-3 py-3">
                          <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${statusStyles[status] ?? statusStyles.active}`}>
                            {status === 'active' ? 'نشط' : status === 'expired' ? 'منتهي' : status === 'canceled' || status === 'cancelled' ? 'موقوف' : status}
                          </span>
                        </td>
                        <td className="px-3 py-3">
                          <div className="flex items-center gap-2 text-foreground">
                            <CalendarClock className="h-4 w-4 text-muted-foreground" />
                            <span>{formatExpiry(user.subscription_expires_at ?? undefined)}</span>
                          </div>
                        </td>
                        <td className="px-3 py-3">
                          <span className={isExpired ? 'text-red-600' : 'text-emerald-600'}>
                            {remainingDays === null ? 'غير محدد' : `${remainingDays} يوم`}
                          </span>
                        </td>
                        <td className="px-3 py-3">
                          <div className="flex flex-wrap gap-2">
                            <Button
                              size="sm"
                              variant="secondary"
                              onClick={() => renewMutation.mutate({ id: user.id, days: 30 })}
                              disabled={renewMutation.isPending}
                            >
                              +30 يوم
                            </Button>
                            <Button
                              size="sm"
                              variant="success"
                              onClick={() => updateStatusMutation.mutate({ id: user.id, status: 'active', expiresAt: new Date(Date.now() + 30 * 86400000).toISOString() })}
                              disabled={updateStatusMutation.isPending}
                            >
                              <CheckCircle2 className="h-4 w-4" />
                              تفعيل
                            </Button>
                            <Button
                              size="sm"
                              variant="danger"
                              onClick={() => updateStatusMutation.mutate({ id: user.id, status: 'expired', expiresAt: new Date(Date.now() - 1000).toISOString() })}
                              disabled={updateStatusMutation.isPending}
                            >
                              <XCircle className="h-4 w-4" />
                              إنهاء
                            </Button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
