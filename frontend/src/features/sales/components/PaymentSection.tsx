import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { CreditCard, Banknote, Send, Loader2 } from 'lucide-react';
import { PaymentMethod } from '../types/pos.types';

interface PaymentSectionProps {
  paymentMethod: PaymentMethod;
  setPaymentMethod: (method: PaymentMethod) => void;
  paidAmount: string;
  setPaidAmount: (value: string) => void;
  total: number;
  isProcessing: boolean;
  onCheckout: () => void;
}

export function PaymentSection({
  paymentMethod,
  setPaymentMethod,
  paidAmount,
  setPaidAmount,
  total,
  isProcessing,
  onCheckout,
}: PaymentSectionProps) {
  const paid = parseFloat(paidAmount) || 0;
  const remaining = total - paid;

  return (
    <Card style={{
      background: 'linear-gradient(135deg, var(--color-primary-05) 0%, rgba(147, 51, 234, 0.05) 100%)',
      border: '1px solid var(--color-primary-10)',
      backdropFilter: 'blur(10px)',
      transition: 'all 0.3s ease'
    }}
    onMouseEnter={(e) => {
      e.currentTarget.style.background = 'linear-gradient(135deg, var(--color-primary-08) 0%, rgba(147, 51, 234, 0.08) 100%)';
      e.currentTarget.style.borderColor = 'var(--color-primary-20)';
      e.currentTarget.style.transform = 'translateY(-2px)';
      e.currentTarget.style.boxShadow = '0 8px 25px var(--color-primary-10)';
    }}
    onMouseLeave={(e) => {
      e.currentTarget.style.background = 'linear-gradient(135deg, var(--color-primary-05) 0%, rgba(147, 51, 234, 0.05) 100%)';
      e.currentTarget.style.borderColor = 'var(--color-primary-10)';
      e.currentTarget.style.transform = 'translateY(0)';
      e.currentTarget.style.boxShadow = 'none';
    }}>
      <CardHeader>
        <CardTitle style={{ 
          display: 'flex', 
          alignItems: 'center', 
          gap: '10px',
          fontSize: '15px',
          fontWeight: '600',
          color: 'var(--text-primary)'
        }}>
          <div style={{
            width: '36px',
            height: '36px',
            borderRadius: '10px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'linear-gradient(135deg, var(--color-primary-15) 0%, rgba(147, 51, 234, 0.15) 100%)',
            border: '1px solid var(--color-primary-25)'
          }}>
            <CreditCard className="w-5 h-5 text-blue-400" />
          </div>
          ملخص الدفع
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            fontSize: '16px',
            fontWeight: '700',
            color: 'var(--text-primary)',
            paddingTop: '12px',
            borderTop: '1px solid var(--color-primary-15)',
            letterSpacing: '0.2px'
          }}>
            <span>الإجمالي</span>
            <span style={{
              color: 'var(--color-info)',
              textShadow: '0 0 20px var(--color-info-30)'
            }}>₪{total.toFixed(2)}</span>
          </div>

          {/* Payment Method */}
          <div style={{ marginTop: '14px' }}>
            <label style={{
              fontSize: '12px',
              color: 'var(--text-secondary)',
              marginBottom: '8px',
              display: 'block',
              fontWeight: '500',
              letterSpacing: '0.2px'
            }}>
              طريقة الدفع
            </label>
            <div style={{ display: 'flex', gap: '10px' }}>
              <Button
                variant={paymentMethod === 'cash' ? 'primary' : 'secondary'}
                onClick={() => setPaymentMethod('cash')}
                style={{
                  flex: 1,
                  height: '40px',
                  fontSize: '13px',
                  fontWeight: '600',
                  letterSpacing: '0.2px',
                  transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
                  borderRadius: '8px',
                  background: paymentMethod === 'cash'
                    ? 'linear-gradient(135deg, var(--button-primary-bg) 0%, var(--color-primary-85) 100%)'
                    : 'linear-gradient(135deg, var(--bg-surface-elevated) 0%, var(--bg-surface) 100%)',
                  border: paymentMethod === 'cash'
                    ? '1px solid var(--color-primary-30)'
                    : '1px solid var(--color-primary-20)',
                  boxShadow: paymentMethod === 'cash'
                    ? '0 4px 12px var(--color-primary-20)'
                    : 'none',
                  backdropFilter: 'blur(10px)'
                }}
                onMouseEnter={(e) => {
                  if (paymentMethod !== 'cash') {
                    e.currentTarget.style.borderColor = 'var(--color-primary-40)';
                    e.currentTarget.style.background = 'linear-gradient(135deg, var(--color-primary-10) 0%, var(--color-info-10) 100%)';
                  }
                }}
                onMouseLeave={(e) => {
                  if (paymentMethod !== 'cash') {
                    e.currentTarget.style.borderColor = 'var(--color-primary-20)';
                    e.currentTarget.style.background = 'linear-gradient(135deg, var(--bg-surface-elevated) 0%, var(--bg-surface) 100%)';
                  }
                }}
              >
                <Banknote className="w-3.5 h-3.5 mr-1.5" style={{ color: paymentMethod === 'cash' ? '#fff' : 'var(--color-info)' }} />
                <span style={{ color: paymentMethod === 'cash' ? '#fff' : 'var(--text-primary)' }}>نقداً</span>
              </Button>
              <Button
                variant={paymentMethod === 'card' ? 'primary' : 'secondary'}
                onClick={() => setPaymentMethod('card')}
                style={{
                  flex: 1,
                  height: '40px',
                  fontSize: '13px',
                  fontWeight: '600',
                  letterSpacing: '0.2px',
                  transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
                  borderRadius: '8px',
                  background: paymentMethod === 'card'
                    ? 'linear-gradient(135deg, var(--button-primary-bg) 0%, var(--color-primary-85) 100%)'
                    : 'linear-gradient(135deg, var(--bg-surface-elevated) 0%, var(--bg-surface) 100%)',
                  border: paymentMethod === 'card'
                    ? '1px solid var(--color-primary-30)'
                    : '1px solid var(--color-primary-20)',
                  boxShadow: paymentMethod === 'card'
                    ? '0 4px 12px var(--color-primary-20)'
                    : 'none',
                  backdropFilter: 'blur(10px)'
                }}
                onMouseEnter={(e) => {
                  if (paymentMethod !== 'card') {
                    e.currentTarget.style.borderColor = 'var(--color-primary-40)';
                    e.currentTarget.style.background = 'linear-gradient(135deg, var(--color-primary-10) 0%, var(--color-info-10) 100%)';
                  }
                }}
                onMouseLeave={(e) => {
                  if (paymentMethod !== 'card') {
                    e.currentTarget.style.borderColor = 'var(--color-primary-20)';
                    e.currentTarget.style.background = 'linear-gradient(135deg, var(--bg-surface-elevated) 0%, var(--bg-surface) 100%)';
                  }
                }}
              >
                <CreditCard className="w-3.5 h-3.5 mr-1.5" style={{ color: paymentMethod === 'card' ? '#fff' : 'var(--color-info)' }} />
                <span style={{ color: paymentMethod === 'card' ? '#fff' : 'var(--text-primary)' }}>بطاقة</span>
              </Button>
            </div>
          </div>

          {/* Amount Paid */}
          {paymentMethod === 'cash' && (
            <div style={{ marginTop: '14px' }}>
              <label style={{
                fontSize: '12px',
                color: 'var(--text-secondary)',
                marginBottom: '8px',
                display: 'block',
                fontWeight: '500',
                letterSpacing: '0.2px'
              }}>
                المبلغ المدفوع
              </label>
              <Input
                type="number"
                value={paidAmount}
                onChange={(e) => setPaidAmount(e.target.value)}
                placeholder="أدخل المبلغ..."
                style={{
                  background: 'linear-gradient(135deg, var(--bg-surface-elevated) 0%, var(--bg-surface) 100%)',
                  border: '1px solid var(--color-primary-20)',
                  backdropFilter: 'blur(10px)',
                  color: 'var(--text-primary)',
                  fontSize: '14px',
                  fontWeight: '500',
                  letterSpacing: '0.2px',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                  borderRadius: '8px'
                }}
                className="hover:border-indigo-400/50 focus:border-indigo-400/70 focus:shadow-lg focus:shadow-indigo-500/10"
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = 'var(--color-primary-40)';
                  e.currentTarget.style.boxShadow = '0 4px 15px var(--color-primary-15)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = 'var(--color-primary-20)';
                  e.currentTarget.style.boxShadow = 'none';
                }}
              />
              {remaining > 0 && (
                <div style={{
                  fontSize: '12px',
                  color: 'var(--color-danger)',
                  marginTop: '6px',
                  fontWeight: '500',
                  letterSpacing: '0.2px'
                }}>
                  المتبقي: ₪{remaining.toFixed(2)}
                </div>
              )}
              {remaining < 0 && (
                <div style={{
                  fontSize: '12px',
                  color: 'var(--color-success)',
                  marginTop: '6px',
                  fontWeight: '500',
                  letterSpacing: '0.2px'
                }}>
                  المردود: ₪{Math.abs(remaining).toFixed(2)}
                </div>
              )}
            </div>
          )}

          {/* Checkout Button */}
          <Button
            variant="primary"
            onClick={onCheckout}
            disabled={isProcessing || (paymentMethod === 'cash' && paid < total)}
            style={{
              width: '100%',
              marginTop: '16px',
              height: '48px',
              fontSize: '15px',
              fontWeight: '700',
              letterSpacing: '0.3px',
              background: isProcessing || (paymentMethod === 'cash' && paid < total)
                ? 'linear-gradient(135deg, rgba(100, 116, 139, 0.3) 0%, rgba(75, 85, 99, 0.3) 100%)'
                : 'linear-gradient(135deg, var(--button-primary-bg) 0%, var(--color-primary-90) 100%)',
              border: isProcessing || (paymentMethod === 'cash' && paid < total)
                ? '1px solid rgba(100, 116, 139, 0.3)'
                : '1px solid var(--color-primary-30)',
              boxShadow: isProcessing || (paymentMethod === 'cash' && paid < total)
                ? 'none'
                : '0 4px 20px var(--color-primary-30)',
              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
              borderRadius: '10px',
              backdropFilter: 'blur(10px)',
              cursor: isProcessing || (paymentMethod === 'cash' && paid < total)
                ? 'not-allowed'
                : 'pointer'
            }}
            onMouseEnter={(e) => {
              if (!(isProcessing || (paymentMethod === 'cash' && paid < total))) {
                e.currentTarget.style.transform = 'translateY(-2px) scale(1.01)';
                e.currentTarget.style.boxShadow = '0 8px 30px var(--color-primary-40)';
              }
            }}
            onMouseLeave={(e) => {
              if (!(isProcessing || (paymentMethod === 'cash' && paid < total))) {
                e.currentTarget.style.transform = 'translateY(0) scale(1)';
                e.currentTarget.style.boxShadow = '0 4px 20px var(--color-primary-30)';
              }
            }}
            onMouseDown={(e) => {
              if (!(isProcessing || (paymentMethod === 'cash' && paid < total))) {
                e.currentTarget.style.transform = 'translateY(0) scale(0.98)';
              }
            }}
            onMouseUp={(e) => {
              if (!(isProcessing || (paymentMethod === 'cash' && paid < total))) {
                e.currentTarget.style.transform = 'translateY(-2px) scale(1.01)';
              }
            }}
          >
            {isProcessing ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" style={{ color: '#fff' }} />
                <span style={{ color: '#fff' }}>جاري المعالجة...</span>
              </>
            ) : (
              <>
                <Send className="w-4 h-4 mr-2" style={{ color: '#fff' }} />
                <span style={{ color: '#fff' }}>إتمام البيع</span>
              </>
            )}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}