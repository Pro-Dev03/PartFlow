import { useState } from 'react';
import { Card, CardContent } from '../ui/card';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { 
  CheckCircle, 
  AlertCircle, 
  Loader2,
  ArrowRight,
  CreditCard,
  DollarSign
} from 'lucide-react';

interface CheckoutFlowProps {
  cart: any[];
  total: number;
  onComplete: (paymentData: any) => void;
  onCancel: () => void;
}

export function CheckoutFlow({ cart, total, onComplete, onCancel }: CheckoutFlowProps) {
  const [step, setStep] = useState<'payment' | 'confirm' | 'processing' | 'success'>('payment');
  const [paymentMethod, setPaymentMethod] = useState<'cash' | 'card' | 'bank_transfer' | 'credit'>('cash');
  const [paidAmount, setPaidAmount] = useState('');

  const paid = parseFloat(paidAmount) || 0;
  const remaining = total - paid;

  const handlePaymentSubmit = () => {
    if (paymentMethod === 'credit' && paid >= total) {
      setPaymentMethod('cash');
      setPaidAmount(total.toString());
    }
    setStep('confirm');
  };

  const handleConfirm = () => {
    setStep('processing');
    
    const paymentData = {
      paymentMethod,
      paidAmount: paymentMethod === 'credit' ? paid : total,
    };

    // Simulate processing
    setTimeout(() => {
      onComplete(paymentData);
      setStep('success');
    }, 1500);
  };

  const paymentMethods = [
    { value: 'cash', label: 'نقد', icon: DollarSign },
    { value: 'card', label: 'بطاقة', icon: CreditCard },
    { value: 'bank_transfer', label: 'تحويل', icon: ArrowRight },
    { value: 'credit', label: 'دين', icon: CreditCard },
  ];

  if (step === 'success') {
    return (
      <Card className="w-full max-w-md mx-auto">
        <CardContent className="p-8 text-center">
          <div className="w-16 h-16 bg-green-100 dark:bg-green-900/20 rounded-full flex items-center justify-center mx-auto mb-4">
            <CheckCircle className="w-8 h-8 text-green-600" />
          </div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">
            تمت العملية بنجاح
          </h2>
          <p className="text-gray-500 dark:text-gray-400 mb-6">
            تم إتمام البيع بنجاح
          </p>
          <Button onClick={onCancel} className="w-full">
            متابعة
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (step === 'processing') {
    return (
      <Card className="w-full max-w-md mx-auto">
        <CardContent className="p-8 text-center">
          <Loader2 className="w-12 h-12 animate-spin text-primary-600 mx-auto mb-4" />
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-2">
            جاري معالجة العملية
          </h2>
          <p className="text-gray-500 dark:text-gray-400">
            يرجى الانتظار...
          </p>
        </CardContent>
      </Card>
    );
  }

  if (step === 'confirm') {
    return (
      <Card className="w-full max-w-md mx-auto">
        <CardContent className="p-6">
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-6">
            تأكيد العملية
          </h2>

          {/* Summary */}
          <div className="space-y-3 mb-6">
            <div className="flex justify-between text-sm">
              <span className="text-gray-600 dark:text-gray-400">عدد العناصر</span>
              <span className="font-medium">{cart.length}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-600 dark:text-gray-400">المجموع</span>
              <span className="font-bold">₪{total.toLocaleString()}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-600 dark:text-gray-400">طريقة الدفع</span>
              <span className="font-medium">
                {paymentMethods.find(m => m.value === paymentMethod)?.label}
              </span>
            </div>
            {paymentMethod === 'credit' && (
              <>
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600 dark:text-gray-400">المدفوع</span>
                  <span className="font-medium text-green-600">₪{paid.toLocaleString()}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600 dark:text-gray-400">المتبقي</span>
                  <span className="font-medium text-orange-600">₪{remaining.toLocaleString()}</span>
                </div>
              </>
            )}
          </div>

          {/* Actions */}
          <div className="space-y-2">
            <Button onClick={handleConfirm} className="w-full" size="lg">
              تأكيد الدفع
            </Button>
            <Button onClick={() => setStep('payment')} variant="outline" className="w-full">
              عودة
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full max-w-md mx-auto">
      <CardContent className="p-6">
        <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-6">
          إتمام الدفع
        </h2>

        {/* Total */}
        <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 mb-6">
          <div className="text-center">
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-1">المجموع</p>
            <p className="text-3xl font-bold text-gray-900 dark:text-gray-100">
              ₪{total.toLocaleString()}
            </p>
          </div>
        </div>

        {/* Payment Methods */}
        <div className="space-y-3 mb-6">
          <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
            طريقة الدفع
          </label>
          <div className="grid grid-cols-2 gap-2">
            {paymentMethods.map((method) => {
              const Icon = method.icon;
              return (
                <Button
                  key={method.value}
                  variant={paymentMethod === method.value ? 'primary' : 'outline'}
                  onClick={() => setPaymentMethod(method.value as any)}
                  className="flex flex-col items-center gap-1 h-auto py-3"
                >
                  <Icon className="w-4 h-4" />
                  <span className="text-xs">{method.label}</span>
                </Button>
              );
            })}
          </div>
        </div>

        {/* Credit Payment */}
        {paymentMethod === 'credit' && (
          <div className="space-y-3 mb-6">
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
            />
            {remaining > 0 && (
              <div className="flex items-center gap-2 text-sm text-orange-600">
                <AlertCircle className="w-4 h-4" />
                <span>المتبقي كدين: ₪{remaining.toLocaleString()}</span>
              </div>
            )}
          </div>
        )}

        {/* Actions */}
        <div className="space-y-2">
          <Button 
            onClick={handlePaymentSubmit} 
            className="w-full" 
            size="lg"
            disabled={paymentMethod === 'credit' && paid <= 0}
          >
            متابعة
          </Button>
          <Button onClick={onCancel} variant="outline" className="w-full">
            إلغاء
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
