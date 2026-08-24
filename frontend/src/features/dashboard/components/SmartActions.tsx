import { useNavigate } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { Button } from '../../../components/ui/button';
import { Card, CardContent } from '../../../components/ui/card';
import { ShoppingCart, Plus, CreditCard, User, Receipt, Zap } from 'lucide-react';

export function SmartActions() {
  const navigate = useNavigate();
  const { t } = useTranslation();

  return (
    <Card>
      <CardContent>
        <div className="flex items-center gap-2 mb-4">
          <Zap className="w-4 h-4" style={{ color: 'var(--color-primary)' }} />
          <h3 className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>{t('dashboard.smartActions')}</h3>
        </div>
        <div className="flex justify-center flex-wrap gap-2 px-6">
          <Button
            variant="secondary"
            onClick={() => navigate('/app/sales')}
            className="gap-2"
          >
            <ShoppingCart className="w-4 h-4" />
            {t('dashboard.newSale')}
          </Button>
          <Button
            variant="secondary"
            onClick={() => navigate('/app/inventory')}
            className="gap-2"
          >
            <Plus className="w-4 h-4" />
            {t('dashboard.addProduct')}
          </Button>
          <Button
            variant="secondary"
            onClick={() => navigate('/app/customers')}
            className="gap-2"
          >
            <User className="w-4 h-4" />
            {t('dashboard.addCustomer')}
          </Button>
          <Button
            variant="secondary"
            onClick={() => navigate('/app/debts')}
            className="gap-2"
          >
            <CreditCard className="w-4 h-4" />
            {t('dashboard.recordPayment')}
          </Button>
          <Button
            variant="secondary"
            onClick={() => navigate('/app/expenses')}
            className="gap-2"
          >
            <Receipt className="w-4 h-4" />
            {t('dashboard.addExpense')}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}