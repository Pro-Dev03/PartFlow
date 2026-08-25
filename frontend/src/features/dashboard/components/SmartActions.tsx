import { useNavigate } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { Card, CardContent } from '../../../components/ui/card';
import { ShoppingCart, Plus, User, CreditCard, Receipt, Zap } from 'lucide-react';

interface ActionItem {
  title: string;
  icon: any;
  path: string;
  variant?: 'primary' | 'secondary' | 'success' | 'warning' | 'info';
  color?: string;
  bgColor?: string;
}

export function SmartActions() {
  const navigate = useNavigate();
  const { t } = useTranslation();

  const actions: ActionItem[] = [
    {
      title: t('dashboard.newSale') || 'بيع جديد',
      icon: ShoppingCart,
      path: '/app/sales',
      variant: 'primary',
      color: 'var(--color-primary)',
      bgColor: 'var(--color-primary-10)'
    },
    {
      title: t('dashboard.addProduct') || 'إضافة قطعة',
      icon: Plus,
      path: '/app/inventory',
      variant: 'secondary',
      color: 'var(--text-primary)',
      bgColor: 'var(--bg-surface-elevated)'
    },
    {
      title: t('dashboard.addCustomer') || 'إضافة عميل',
      icon: User,
      path: '/app/customers',
      variant: 'secondary',
      color: 'var(--text-primary)',
      bgColor: 'var(--bg-surface-elevated)'
    },
    {
      title: t('dashboard.recordPayment') || 'تسجيل دفعة',
      icon: CreditCard,
      path: '/app/debts',
      variant: 'secondary',
      color: 'var(--text-primary)',
      bgColor: 'var(--bg-surface-elevated)'
    },
    {
      title: t('dashboard.addExpense') || 'إضافة مصروف',
      icon: Receipt,
      path: '/app/expenses',
      variant: 'secondary',
      color: 'var(--text-primary)',
      bgColor: 'var(--bg-surface-elevated)'
    }
  ];

  return (
    <Card style={{
      background: 'var(--bg-surface)',
      border: '1px solid var(--border-default)'
    }}>
      <CardContent>
        <div className="flex items-center gap-2 mb-4">
          <div 
            className="w-8 h-8 rounded-lg flex items-center justify-center"
            style={{ 
              background: 'var(--color-primary-15)',
              border: '1px solid var(--color-primary-25)'
            }}
          >
            <Zap className="w-4 h-4" style={{ color: 'var(--color-primary)' }} />
          </div>
          <h3 className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>
            {t('dashboard.smartActions')}
          </h3>
        </div>

        <div 
          className="grid gap-3"
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))',
            gap: '12px'
          }}
        >
          {actions.map((action, index) => {
            const Icon = action.icon;
            const isPrimary = action.variant === 'primary';
            
            return (
              <button
                key={index}
                onClick={() => navigate(action.path)}
                className="flex flex-col items-center justify-center gap-2 p-4 rounded-xl transition-all duration-200 cursor-pointer group"
                style={{
                  background: isPrimary ? 'var(--gradient-primary)' : action.bgColor,
                  border: `1px solid ${isPrimary ? 'var(--color-primary-30)' : 'var(--border-default)'}`,
                  color: isPrimary ? '#FFFFFF' : action.color,
                  boxShadow: isPrimary ? 'var(--shadow-glow)' : 'var(--shadow-sm)',
                  transform: 'translateY(0)',
                  minHeight: '100px'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.transform = 'translateY(-3px)';
                  e.currentTarget.style.boxShadow = isPrimary 
                    ? 'var(--shadow-glow-strong)' 
                    : 'var(--shadow-md)';
                  e.currentTarget.style.borderColor = isPrimary 
                    ? 'var(--color-primary-40)' 
                    : 'var(--color-primary-20)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.transform = 'translateY(0)';
                  e.currentTarget.style.boxShadow = isPrimary 
                    ? 'var(--shadow-glow)' 
                    : 'var(--shadow-sm)';
                  e.currentTarget.style.borderColor = isPrimary 
                    ? 'var(--color-primary-30)' 
                    : 'var(--border-default)';
                }}
              >
                <div 
                  className="w-10 h-10 rounded-lg flex items-center justify-center transition-transform duration-200 group-hover:scale-110"
                  style={{ 
                    background: isPrimary 
                      ? 'rgba(255, 255, 255, 0.15)' 
                      : 'var(--color-primary-10)',
                    border: `1px solid ${isPrimary ? 'rgba(255, 255, 255, 0.2)' : 'var(--color-primary-15)'}`
                  }}
                >
                  <Icon 
                    className="w-5 h-5" 
                    style={{ 
                      color: isPrimary ? '#FFFFFF' : 'var(--color-primary)'
                    }} 
                  />
                </div>
                <span 
                  className="text-xs font-medium text-center leading-tight"
                  style={{ 
                    color: isPrimary ? '#FFFFFF' : 'var(--text-primary)',
                    fontSize: '12px',
                    fontWeight: '500'
                  }}
                >
                  {action.title}
                </span>
              </button>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
