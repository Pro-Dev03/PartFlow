import { useState, useEffect, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from '../../hooks/useTranslation';
import { notificationsApi } from '../../services/api/endpoints';
import { Button } from '../ui/button';
import { Card, CardContent } from '../ui/card';
import { Badge } from '../ui/badge';
import { playScanSound } from '../../hooks/useBarcodeContext';
import { 
  Bell, 
  X, 
  Check, 
  CheckCheck, 
  Settings,
  Package,
  DollarSign,
  Clock,
  ArrowRight,
  Filter,
  Volume2,
  VolumeX,
  AlertTriangle,
  Info,
  CheckCircle as CheckCircleIcon,
  BellOff
} from 'lucide-react';

export function NotificationCenter() {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);
  const [soundEnabled, setSoundEnabled] = useState(true);
  const [browserNotificationsEnabled, setBrowserNotificationsEnabled] = useState(false);
  const [filterType, setFilterType] = useState<string>('all');
  const [previousUnreadCount, setPreviousUnreadCount] = useState(0);
  const [previousNotifications, setPreviousNotifications] = useState<any[]>([]);
  const queryClient = useQueryClient();

  // Request browser notification permission on mount
  useEffect(() => {
    if ('Notification' in window) {
      setBrowserNotificationsEnabled(Notification.permission === 'granted');
    }
  }, []);

  const requestBrowserNotificationPermission = async () => {
    if ('Notification' in window) {
      const permission = await Notification.requestPermission();
      setBrowserNotificationsEnabled(permission === 'granted');
      return permission === 'granted';
    }
    return false;
  };

  const showBrowserNotification = (notification: any) => {
    if (browserNotificationsEnabled && 'Notification' in window && Notification.permission === 'granted') {
      const notif = new Notification(notification.title, {
        body: notification.message,
        icon: '/favicon.ico',
        badge: '/favicon.ico',
        tag: notification.id,
        requireInteraction: notification.priority === 'urgent' || notification.type === 'debt_overdue',
        silent: !soundEnabled,
      });

      // Handle notification click
      notif.onclick = (event) => {
        event.preventDefault();
        window.focus();
        if (notification.actionUrl) {
          window.location.href = notification.actionUrl;
        }
        notif.close();
      };

      // Auto-close after 5 seconds for non-urgent notifications
      if (notification.priority !== 'urgent' && notification.type !== 'debt_overdue') {
        setTimeout(() => notif.close(), 5000);
      }
    }
  };

  const { data: notificationsData, isLoading } = useQuery({
    queryKey: ['notifications'],
    queryFn: () => notificationsApi.list(),
    refetchInterval: 30000, // Poll every 30 seconds
  });

  const { data: unreadCountData } = useQuery({
    queryKey: ['notifications', 'unread-count'],
    queryFn: () => notificationsApi.getUnreadCount(),
    refetchInterval: 30000,
  });

  const notifications = (notificationsData as unknown) as any[];
  const unreadCount = (unreadCountData as unknown) as number;

  // Play sound and show browser notifications when new notifications arrive
  useEffect(() => {
    const currentUnreadCount = unreadCount || 0;
    if (currentUnreadCount > previousUnreadCount && soundEnabled) {
      playScanSound(true);
    }
    setPreviousUnreadCount(currentUnreadCount);
  }, [unreadCount, previousUnreadCount, soundEnabled]);

  // Show browser notifications for new notifications
  useEffect(() => {
    if (notifications && notifications.length > previousNotifications.length) {
      const newNotifications = notifications.slice(previousNotifications.length);
      newNotifications.forEach((notification: any) => {
        if (!notification.read) {
          showBrowserNotification(notification);
        }
      });
    }
    setPreviousNotifications(notifications || []);
  }, [notifications, previousNotifications]);

  const markAsReadMutation = useMutation({
    mutationFn: (id: string) => notificationsApi.markAsRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      queryClient.invalidateQueries({ queryKey: ['notifications', 'unread-count'] });
    },
  });

  const markAllAsReadMutation = useMutation({
    mutationFn: () => notificationsApi.markAllAsRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      queryClient.invalidateQueries({ queryKey: ['notifications', 'unread-count'] });
    },
  });

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case 'LOW_STOCK':
      case 'low_stock':
        return <Package className="w-5 h-5 text-orange-500" />;
      case 'OVERDUE_DEBT':
      case 'debt_overdue':
        return <DollarSign className="w-5 h-5 text-red-500" />;
      case 'WARRANTY_EXPIRING':
      case 'warranty_expiring':
        return <Clock className="w-5 h-5 text-yellow-500" />;
      case 'ITEM_RESERVED':
        return <Package className="w-5 h-5 text-blue-500" />;
      case 'PAYMENT_RECEIVED':
        return <DollarSign className="w-5 h-5 text-green-500" />;
      case 'PURCHASE_RECEIVED':
        return <Package className="w-5 h-5 text-green-500" />;
      case 'INSPECTION_REQUIRED':
        return <AlertTriangle className="w-5 h-5 text-purple-500" />;
      case 'RESERVATION_EXPIRING':
        return <Clock className="w-5 h-5 text-orange-500" />;
      case 'daily_insights':
        return <Info className="w-5 h-5 text-blue-500" />;
      case 'INFO':
        return <Info className="w-5 h-5 text-blue-500" />;
      case 'SUCCESS':
        return <CheckCircleIcon className="w-5 h-5 text-green-500" />;
      default:
        return <Bell className="w-5 h-5 text-gray-500" />;
    }
  };

  const getNotificationColor = (type: string) => {
    switch (type) {
      case 'LOW_STOCK':
      case 'low_stock':
        return 'bg-orange-50 dark:bg-orange-900/20 border-orange-200 dark:border-orange-800';
      case 'OVERDUE_DEBT':
      case 'debt_overdue':
        return 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800';
      case 'WARRANTY_EXPIRING':
      case 'warranty_expiring':
      case 'RESERVATION_EXPIRING':
        return 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800';
      case 'ITEM_RESERVED':
        return 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800';
      case 'PAYMENT_RECEIVED':
      case 'PURCHASE_RECEIVED':
      case 'SUCCESS':
        return 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800';
      case 'INSPECTION_REQUIRED':
        return 'bg-purple-50 dark:bg-purple-900/20 border-purple-200 dark:border-purple-800';
      case 'daily_insights':
      case 'INFO':
        return 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800';
      default:
        return 'bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700';
    }
  };

  const handleMarkAsRead = useCallback((id: string) => {
    markAsReadMutation.mutate(id);
  }, [markAsReadMutation]);

  const handleMarkAllAsRead = useCallback(() => {
    markAllAsReadMutation.mutate();
  }, [markAllAsReadMutation]);

  // Filter notifications based on type
  const filteredNotifications = notifications?.filter((notification: any) => {
    if (filterType === 'all') return true;
    return notification.type === filterType;
  }) || [];

  if (!isOpen) {
    return (
      <div className="relative">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setIsOpen(true)}
          className="relative"
          title={t('notifications.title')}
        >
          <Bell className="w-5 h-5" />
          {unreadCount && unreadCount > 0 && (
            <Badge
              variant="destructive"
              className="absolute -top-1 -end-1 h-5 w-5 flex items-center justify-center p-0 text-xs"
            >
              {unreadCount > 9 ? '9+' : unreadCount}
            </Badge>
          )}
        </Button>
        {/* Notification toggles when closed */}
        <div className="absolute -top-8 -right-0 flex gap-1 opacity-0 hover:opacity-100 transition-opacity">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setSoundEnabled(!soundEnabled)}
            title={soundEnabled ? 'إيقاف الصوت' : 'تشغيل الصوت'}
          >
            {soundEnabled ? (
              <Volume2 className="w-4 h-4" />
            ) : (
              <VolumeX className="w-4 h-4" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              if (!browserNotificationsEnabled) {
                requestBrowserNotificationPermission();
              } else {
                setBrowserNotificationsEnabled(false);
              }
            }}
            title={browserNotificationsEnabled ? 'إيقاف إشعارات المتصفح' : 'تفعيل إشعارات المتصفح'}
          >
            {browserNotificationsEnabled ? (
              <Bell className="w-4 h-4" />
            ) : (
              <BellOff className="w-4 h-4" />
            )}
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-end p-4">
      <div className="bg-black/50 absolute inset-0" onClick={() => setIsOpen(false)} />
      <Card className="relative w-full max-w-md max-h-[80vh] overflow-hidden">
        <CardContent className="p-0">
          {/* Header */}
          <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center gap-2">
              <Bell className="w-5 h-5" />
              <h2 className="text-lg font-bold text-gray-900 dark:text-gray-100">
                {t('notifications.title')}
              </h2>
              {unreadCount && unreadCount > 0 && (
                <Badge variant="secondary">{unreadCount} {t('common.new')}</Badge>
              )}
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setSoundEnabled(!soundEnabled)}
                title={soundEnabled ? 'إيقاف الصوت' : 'تشغيل الصوت'}
              >
                {soundEnabled ? (
                  <Volume2 className="w-4 h-4" />
                ) : (
                  <VolumeX className="w-4 h-4" />
                )}
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  if (!browserNotificationsEnabled) {
                    requestBrowserNotificationPermission();
                  } else {
                    setBrowserNotificationsEnabled(false);
                  }
                }}
                title={browserNotificationsEnabled ? 'إيقاف إشعارات المتصفح' : 'تفعيل إشعارات المتصفح'}
              >
                {browserNotificationsEnabled ? (
                  <Bell className="w-4 h-4" />
                ) : (
                  <BellOff className="w-4 h-4" />
                )}
              </Button>
              {unreadCount && unreadCount > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleMarkAllAsRead}
                  disabled={markAllAsReadMutation.isPending}
                >
                  <CheckCheck className="w-4 h-4 me-1" />
                  {t('notifications.markAllRead')}
                </Button>
              )}
              <Button variant="ghost" size="sm" onClick={() => setIsOpen(false)}>
                <X className="w-4 h-4" />
              </Button>
            </div>
          </div>

          {/* Filter Tabs */}
          <div className="flex gap-2 p-4 border-b border-gray-200 dark:border-gray-700 overflow-x-auto">
            <Button
              variant={filterType === 'all' ? 'primary' : 'ghost'}
              size="sm"
              onClick={() => setFilterType('all')}
              className="flex-shrink-0"
            >
              {t('common.all')}
            </Button>
            <Button
              variant={filterType === 'LOW_STOCK' || filterType === 'low_stock' ? 'primary' : 'ghost'}
              size="sm"
              onClick={() => setFilterType('low_stock')}
              className="flex-shrink-0 gap-1"
            >
              <Package className="w-3 h-3" />
              {t('notifications.types.LOW_STOCK')}
            </Button>
            <Button
              variant={filterType === 'OVERDUE_DEBT' || filterType === 'debt_overdue' ? 'primary' : 'ghost'}
              size="sm"
              onClick={() => setFilterType('debt_overdue')}
              className="flex-shrink-0 gap-1"
            >
              <DollarSign className="w-3 h-3" />
              {t('notifications.types.OVERDUE_DEBT')}
            </Button>
            <Button
              variant={filterType === 'WARRANTY_EXPIRING' || filterType === 'warranty_expiring' ? 'primary' : 'ghost'}
              size="sm"
              onClick={() => setFilterType('warranty_expiring')}
              className="flex-shrink-0 gap-1"
            >
              <Clock className="w-3 h-3" />
              {t('notifications.types.WARRANTY_EXPIRING')}
            </Button>
            <Button
              variant={filterType === 'daily_insights' ? 'primary' : 'ghost'}
              size="sm"
              onClick={() => setFilterType('daily_insights')}
              className="flex-shrink-0 gap-1"
            >
              <Info className="w-3 h-3" />
              رؤى يومية
            </Button>
          </div>

          {/* Notifications List */}
          <div className="overflow-y-auto max-h-[60vh]">
            {isLoading ? (
              <div className="p-4 text-center text-gray-500 dark:text-gray-400">
                {t('common.loading')}
              </div>
            ) : filteredNotifications && filteredNotifications.length > 0 ? (
              filteredNotifications.map((notification: any) => (
                <div
                  key={notification.id}
                  className={`p-4 border-b border-gray-100 dark:border-gray-800 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors ${
                    !notification.read ? 'bg-blue-50/50 dark:bg-blue-900/10' : ''
                  } ${getNotificationColor(notification.type)}`}
                  onClick={() => !notification.read && handleMarkAsRead(notification.id)}
                >
                  <div className="flex items-start gap-3">
                    <div className="flex-shrink-0 mt-1">
                      {getNotificationIcon(notification.type)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-start justify-between gap-2">
                        <p className="font-medium text-gray-900 dark:text-gray-100 text-sm">
                          {notification.title}
                        </p>
                        {!notification.read && (
                          <div className="flex-shrink-0">
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-6 w-6 p-0"
                              onClick={(e) => {
                                e.stopPropagation();
                                handleMarkAsRead(notification.id);
                              }}
                            >
                              <Check className="w-3 h-3" />
                            </Button>
                          </div>
                        )}
                      </div>
                      <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                        {notification.message}
                      </p>
                      <p className="text-xs text-gray-400 dark:text-gray-500 mt-2">
                        {new Date(notification.createdAt).toLocaleString('ar-SA')}
                      </p>
                    </div>
                    {notification.actionUrl && (
                      <ArrowRight className="w-4 h-4 text-gray-400 flex-shrink-0 mt-1" />
                    )}
                  </div>
                </div>
              ))
            ) : (
              <div className="p-8 text-center">
                <Bell className="w-12 h-12 text-gray-300 dark:text-gray-600 mx-auto mb-3" />
                <p className="text-gray-500 dark:text-gray-400">
                  {t('notifications.noNotifications')}
                </p>
              </div>
            )}
          </div>

          {/* Footer */}
          <div className="p-4 border-t border-gray-200 dark:border-gray-700">
            <Button
              variant="outline"
              className="w-full gap-2"
              onClick={() => {
                // Open settings
                setIsOpen(false);
              }}
            >
              <Settings className="w-4 h-4" />
              {t('notifications.preferences')}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
