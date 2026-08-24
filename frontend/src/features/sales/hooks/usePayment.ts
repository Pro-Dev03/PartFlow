import { useState, useCallback } from 'react';
import { PaymentMethod } from '../types/pos.types';

export function usePayment() {
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('cash');
  const [paidAmount, setPaidAmount] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);

  const calculateRemaining = useCallback((total: number) => {
    const paid = parseFloat(paidAmount) || 0;
    return total - paid;
  }, [paidAmount]);

  const resetPayment = useCallback(() => {
    setPaymentMethod('cash');
    setPaidAmount('');
    setIsProcessing(false);
  }, []);

  const setProcessing = useCallback((processing: boolean) => {
    setIsProcessing(processing);
  }, []);

  return {
    paymentMethod,
    setPaymentMethod,
    paidAmount,
    setPaidAmount,
    isProcessing,
    setProcessing,
    calculateRemaining,
    resetPayment,
  };
}