import { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  CalendarClock,
  CheckCircle2,
  Clock3,
  KeyRound,
  Plus,
  Search,
  ShieldAlert,
  ShieldCheck,
  Trash2,
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
  return new Intl.DateTimeFormat('ar-EG', { dateStyle: 'medium', timeZone: 'UTC' }).format(date);
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
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [passwordUser, setPasswordUser] = useState<User | null>(null);
  const [newPassword, setNewPassword] = useState('');
  const [passwordConfirmation, setPasswordConfirmation] = useState('');
  const [createForm, setCreateForm] = useState({
    firstName: '',
    lastName: '',
    email: '',
    password: '',
    phone: '',
    subscriptionDays: '30',
  });

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
    mutationFn: async ({ id, status, expiresAt, subscriptionDays }: { id: string; status: string; expiresAt?: string | null; subscriptionDays?: number }) => {
      const response = await settingsApi.updateSubscriptionStatus(id, {
        subscription_status: status,
        ...(subscriptionDays !== undefined ? { subscription_days: subscriptionDays } : { subscription_expires_at: expiresAt ?? null }),
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

  const createMutation = useMutation({
    mutationFn: async () => {
      const response = await settingsApi.createUser({
        first_name: createForm.firstName.trim(),
        last_name: createForm.lastName.trim(),
        email: createForm.email.trim(),
        password: createForm.password,
        phone: createForm.phone.trim() || undefined,
        subscription_days: Number(createForm.subscriptionDays),
        is_active: true,
      });
      return response.data as User;
    },
    onSuccess: () => {
      setCreateForm({ firstName: '', lastName: '', email: '', password: '', phone: '', subscriptionDays: '30' });
      setShowCreateForm(false);
      toast.success('تم إنشاء حساب المشترك بنجاح');
      queryClient.invalidateQueries({ queryKey: ['subscription-subscribers'] });
      queryClient.invalidateQueries({ queryKey: ['subscription-summary'] });
    },
    onError: (error: any) => toast.error(error?.message || 'تعذر إنشاء الحساب'),
  });

  const passwordMutation = useMutation({
    mutationFn: async ({ user, password }: { user: User; password: string }) => {
      const response = await settingsApi.updateUser(user.id, {
        email: user.email,
        first_name: user.first_name || user.name || '',
        last_name: user.last_name || '',
        phone: user.phone || undefined,
        password,
        is_active: user.is_active ?? true,
      });
      return response.data as User;
    },
    onSuccess: () => toast.success('تم تغيير كلمة مرور المشترك'),
    onError: (error: any) => toast.error(error?.message || 'تعذر تغيير كلمة المرور'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => settingsApi.deleteUser(id),
    onSuccess: () => {
      toast.success('تم حذف حساب المشترك');
      queryClient.invalidateQueries({ queryKey: ['subscription-subscribers'] });
      queryClient.invalidateQueries({ queryKey: ['subscription-summary'] });
    },
    onError: (error: any) => toast.error(error?.message || 'تعذر حذف الحساب'),
  });

  const resetPassword = (user: User) => {
    setPasswordUser(user);
    setNewPassword('');
    setPasswordConfirmation('');
  };

  const closePasswordDialog = () => {
    setPasswordUser(null);
    setNewPassword('');
    setPasswordConfirmation('');
  };

  const submitPasswordChange = () => {
    const password = newPassword.trim();
    if (password.length < 8) {
      toast.error('يجب أن تتكون كلمة المرور من 8 أحرف على الأقل');
      return;
    }
    if (password !== passwordConfirmation.trim()) {
      toast.error('كلمتا المرور غير متطابقتين');
      return;
    }
    if (passwordUser) {
      passwordMutation.mutate({ user: passwordUser, password }, { onSuccess: closePasswordDialog });
    }
  };

  const deleteSubscriber = (user: User) => {
    if (user.email.toLowerCase() === 'owner@partflow.com') {
      toast.error('لا يمكن حذف حساب المالك الأساسي');
      return;
    }
    if (window.confirm(`هل تريد حذف حساب ${user.email}؟ لا يمكن التراجع عن هذا الإجراء.`)) {
      deleteMutation.mutate(user.id);
    }
  };

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
            <Button size="sm" onClick={() => setShowCreateForm((value) => !value)}>
              <Plus className="h-4 w-4" />
              مشترك جديد
            </Button>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {showCreateForm && (
            <div className="rounded-xl border border-primary/20 bg-primary/5 p-4">
              <div className="mb-3 text-sm font-semibold">إنشاء حساب مشترك</div>
              <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                {([
                  ['firstName', 'الاسم الأول', 'text'],
                  ['lastName', 'اسم العائلة', 'text'],
                  ['email', 'البريد الإلكتروني', 'email'],
                  ['phone', 'رقم الهاتف', 'tel'],
                  ['password', 'كلمة المرور المؤقتة', 'password'],
                  ['subscriptionDays', 'مدة الاشتراك بالأيام', 'number'],
                ] as const).map(([field, label, type]) => (
                  <label key={field} className="space-y-1 text-sm">
                    <span className="text-muted-foreground">{label}</span>
                    <input
                      type={type}
                      value={createForm[field]}
                      onChange={(event) => setCreateForm((current) => ({ ...current, [field]: event.target.value }))}
                      className="w-full rounded-lg border border-border bg-background px-3 py-2 outline-none"
                    />
                  </label>
                ))}
              </div>
              <div className="mt-3 flex flex-wrap gap-2">
                <Button
                  onClick={() => createMutation.mutate()}
                  disabled={createMutation.isPending || !createForm.firstName || !createForm.lastName || !createForm.email || createForm.password.length < 8 || Number(createForm.subscriptionDays) < 1}
                >
                  {createMutation.isPending ? 'جارِ الإنشاء...' : 'إنشاء الحساب'}
                </Button>
                <Button variant="secondary" onClick={() => setShowCreateForm(false)}>إلغاء</Button>
              </div>
            </div>
          )}

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
                    const isOwner = user.email.toLowerCase() === 'owner@partflow.com';
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
                              size="icon"
                              variant="secondary"
                              onClick={() => renewMutation.mutate({ id: user.id, days: 30 })}
                              disabled={renewMutation.isPending}
                              title="تجديد 30 يومًا"
                              aria-label="تجديد 30 يومًا"
                            >
                              <CalendarClock className="h-4 w-4" />
                            </Button>
                            <Button
                              size="icon"
                              variant="success"
                              onClick={() => updateStatusMutation.mutate({ id: user.id, status: 'active', subscriptionDays: 30 })}
                              disabled={updateStatusMutation.isPending}
                              title="تفعيل الاشتراك"
                              aria-label="تفعيل الاشتراك"
                            >
                              <CheckCircle2 className="h-4 w-4" />
                              تفعيل
                            </Button>
                            {!isOwner && (
                              <Button
                                size="icon"
                                variant="danger"
                                onClick={() => updateStatusMutation.mutate({ id: user.id, status: 'expired' })}
                                disabled={updateStatusMutation.isPending}
                                title="إنهاء الاشتراك"
                                aria-label="إنهاء الاشتراك"
                              >
                                <XCircle className="h-4 w-4" />
                                إنهاء
                              </Button>
                            )}
                            <Button
                              size="icon"
                              variant="secondary"
                              onClick={() => resetPassword(user)}
                              disabled={passwordMutation.isPending}
                              title="تغيير كلمة المرور"
                              aria-label="تغيير كلمة المرور"
                            >
                              <KeyRound className="h-4 w-4" />
                            </Button>
                            <Button
                              size="icon"
                              variant="danger"
                              onClick={() => deleteSubscriber(user)}
                              disabled={deleteMutation.isPending || user.email.toLowerCase() === 'owner@partflow.com'}
                              title="حذف الحساب"
                              aria-label="حذف الحساب"
                            >
                              <Trash2 className="h-4 w-4" />
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

      {passwordUser && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" role="dialog" aria-modal="true">
          <div className="w-full max-w-md rounded-xl border border-border bg-background p-5 shadow-xl">
            <div className="mb-4">
              <h3 className="text-lg font-semibold">تغيير كلمة المرور</h3>
              <p className="mt-1 text-sm text-muted-foreground">الحساب: {passwordUser.email}</p>
            </div>
            <div className="space-y-3">
              <label className="block space-y-1 text-sm">
                <span className="text-muted-foreground">كلمة المرور الجديدة</span>
                <input
                  type="password"
                  value={newPassword}
                  onChange={(event) => setNewPassword(event.target.value)}
                  className="w-full rounded-lg border border-border bg-background px-3 py-2 outline-none"
                  autoFocus
                />
              </label>
              <label className="block space-y-1 text-sm">
                <span className="text-muted-foreground">تأكيد كلمة المرور</span>
                <input
                  type="password"
                  value={passwordConfirmation}
                  onChange={(event) => setPasswordConfirmation(event.target.value)}
                  onKeyDown={(event) => {
                    if (event.key === 'Enter') submitPasswordChange();
                  }}
                  className="w-full rounded-lg border border-border bg-background px-3 py-2 outline-none"
                />
              </label>
            </div>
            <div className="mt-5 flex justify-end gap-2">
              <Button variant="secondary" onClick={closePasswordDialog}>إلغاء</Button>
              <Button onClick={submitPasswordChange} disabled={passwordMutation.isPending}>
                {passwordMutation.isPending ? 'جارِ الحفظ...' : 'حفظ كلمة المرور'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
