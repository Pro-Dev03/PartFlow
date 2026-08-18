import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Select } from '../ui/select';
import { Badge } from '../ui/badge';
import { 
  DollarSign, 
  AlertTriangle, 
  CheckCircle,
  User,
  Calendar,
  TrendingUp
} from 'lucide-react';

interface DebtSaleProps {
  total: number;
  customers: any[];
  onComplete: (debtData: any) => void;
  onCancel: () => void;
}

export function DebtSale({ total, customers, onComplete, onCancel }: DebtSaleProps) {
  const [selectedCustomer, setSelectedCustomer] = useState('');
  const [paidAmount, setPaidAmount] = useState('');
  const [dueDate, setDueDate] = useState('');

  const paid = parseFloat(paidAmount) || 0;
  const remaining = total - paid;

  const customer = customers.find((c: any) => c.id === selectedCustomer);
  
  // Calculate default due date (30 days from now)
  const defaultDueDate = new Date();
  defaultDueDate.setDate(defaultDueDate.getDate() + 30);
  const defaultDueDateStr = defaultDueDate.toISOString().split('T')[0];

  const handleSubmit = () => {
    if (!selectedCustomer) return;

    const debtData = {
      customerId: selectedCustomer,
      paidAmount: paid,
      remainingAmount: remaining,
      dueDate: dueDate || defaultDueDateStr,
    };

    onComplete(debtData);
  };

  // Quick amount buttons
  const quickAmounts = [0, total * 0.25, total * 0.5, total * 0.75, total];

  return (
    <Card className="w-full max-w-md mx-auto">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <DollarSign className="w-5 h-5 text-orange-500" />
          بيع على الدين
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Total */}
        <div className="bg-gradient-to-r from-orange-50 to-orange-100 dark:from-orange-900/20 dark:to-orange-800/20 rounded-lg p-4 border border-orange-200 dark:border-orange-800">
          <div className="text-center">
            <p className="text-sm text-orange-600 dark:text-orange-400 mb-1">المجموع</p>
            <p className="text-3xl font-bold text-orange-900 dark:text-orange-100">
              ₪{total.toLocaleString()}
            </p>
          </div>
        </div>

        {/* Customer Selection */}
        <div className="space-y-2">
          <label className="text-sm font-medium text-gray-700 dark:text-gray-300 flex items-center gap-2">
            <User className="w-4 h-4" />
            العميل
          </label>
          <Select
            value={selectedCustomer}
            onChange={(e) => setSelectedCustomer(e.target.value)}
            options={[
              { value: '', label: 'اختر العميل...' },
              ...customers.map((c: any) => ({ 
                value: c.id, 
                label: `${c.name} (دين حالي: ₪${c.outstanding?.toLocaleString() || 0})` 
              })),
            ]}
          />
          {customer && (
            <div className="flex flex-wrap gap-2">
              {customer.creditLimit && (
                <Badge variant="outline" className="text-xs">
                  حد الائتمان: ₪{customer.creditLimit.toLocaleString()}
                </Badge>
              )}
              {customer.outstanding > 0 && (
                <Badge variant="warning" className="text-xs">
                  دين حالي: ₪{customer.outstanding.toLocaleString()}
                </Badge>
              )}
            </div>
          )}
        </div>

        {/* Payment Breakdown */}
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
              المبلغ المدفوع الآن
            </label>
            <Input
              type="number"
              value={paidAmount}
              onChange={(e) => setPaidAmount(e.target.value)}
              placeholder="0.00"
              min="0"
              max={total}
              step="0.01"
            />
            {/* Quick amount buttons */}
            <div className="flex gap-2 flex-wrap">
              {quickAmounts.map((amount, index) => (
                <Button
                  key={index}
                  variant="outline"
                  size="sm"
                  onClick={() => setPaidAmount(amount.toString())}
                  className="text-xs"
                >
                  {amount === 0 ? '0' : amount === total ? 'الكل' : `₪${Math.round(amount)}`}
                </Button>
              ))}
            </div>
          </div>

          <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 space-y-3 border border-gray-200 dark:border-gray-700">
            <div className="flex justify-between text-sm">
              <span className="text-gray-600 dark:text-gray-400">المجموع</span>
              <span className="font-medium">₪{total.toLocaleString()}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-600 dark:text-gray-400">المدفوع</span>
              <span className="font-medium text-green-600">₪{paid.toLocaleString()}</span>
            </div>
            <div className="border-t border-gray-200 dark:border-gray-700 pt-3">
              <div className="flex justify-between items-center">
                <span className="text-base font-bold text-orange-600 dark:text-orange-400">
                  المتبقي كدين
                </span>
                <div className="flex items-center gap-2">
                  <TrendingUp className="w-4 h-4 text-orange-500" />
                  <span className="text-xl font-bold text-orange-600 dark:text-orange-500">
                    ₪{remaining.toLocaleString()}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Due Date */}
        <div className="space-y-2">
          <label className="text-sm font-medium text-gray-700 dark:text-gray-300 flex items-center gap-2">
            <Calendar className="w-4 h-4" />
            تاريخ الاستحقاق
          </label>
          <Input
            type="date"
            value={dueDate}
            onChange={(e) => setDueDate(e.target.value)}
            min={new Date().toISOString().split('T')[0]}
            defaultValue={defaultDueDateStr}
          />
          <p className="text-xs text-gray-500 dark:text-gray-400">
            افتراضياً بعد 30 يوم من اليوم
          </p>
        </div>

        {/* Warning */}
        {remaining > 0 && (
          <div className="flex items-start gap-2 p-3 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg border border-yellow-200 dark:border-yellow-800">
            <AlertTriangle className="w-4 h-4 text-yellow-600 mt-0.5 flex-shrink-0" />
            <p className="text-sm text-yellow-800 dark:text-yellow-200">
              سيتم إضافة <strong>₪{remaining.toLocaleString()}</strong> كدين على العميل. تأكد من إمكانية السداد.
            </p>
          </div>
        )}

        {/* Actions */}
        <div className="space-y-2">
          <Button
            onClick={handleSubmit}
            className="w-full"
            disabled={!selectedCustomer || paid < 0}
            size="lg"
          >
            <CheckCircle className="w-4 h-4 mr-2" />
            تأكيد البيع على الدين
          </Button>
          <Button onClick={onCancel} variant="outline" className="w-full">
            إلغاء
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
