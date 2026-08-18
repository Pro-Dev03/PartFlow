import { useState, useCallback, useEffect } from 'react';
import { barcodeApi } from '../services/api/endpoints';

export const BarcodeContext = {
  SALE: 'sale',
  INVENTORY: 'inventory',
  PURCHASE: 'purchase',
  RETURN: 'return',
  LOOKUP: 'lookup'
} as const;

export type BarcodeContext = typeof BarcodeContext[keyof typeof BarcodeContext];

interface UseBarcodeContextResult {
  context: BarcodeContext;
  setContext: (context: BarcodeContext) => void;
  processBarcode: (barcode: string) => Promise<any>;
  isProcessing: boolean;
  autoDetectContext: () => void;
}

export function useBarcodeContext(initialContext?: BarcodeContext): UseBarcodeContextResult {
  const [context, setContext] = useState<BarcodeContext>(initialContext || BarcodeContext.LOOKUP);
  const [isProcessing, setIsProcessing] = useState(false);

  // Auto-detect context based on current URL
  const autoDetectContext = useCallback(() => {
    const currentPath = window.location.pathname;
    
    if (currentPath.includes('/sales') || currentPath.includes('/pos')) {
      setContext(BarcodeContext.SALE);
    } else if (currentPath.includes('/inventory')) {
      setContext(BarcodeContext.INVENTORY);
    } else if (currentPath.includes('/purchases')) {
      setContext(BarcodeContext.PURCHASE);
    } else if (currentPath.includes('/returns')) {
      setContext(BarcodeContext.RETURN);
    } else {
      setContext(BarcodeContext.LOOKUP);
    }
  }, []);

  // Auto-detect context on mount and route changes
  useEffect(() => {
    if (!initialContext) {
      autoDetectContext();
    }
  }, [autoDetectContext, initialContext]);

  const processBarcode = useCallback(async (barcode: string) => {
    setIsProcessing(true);
    
    try {
      const response = await barcodeApi.scan(barcode);
      const result = response as any;
      
      // Add context information to the result
      return {
        ...result,
        context,
        scannedAt: new Date().toISOString(),
        suggestedAction: getSuggestedAction(context, result),
      };
    } catch (error) {
      console.error('Barcode processing failed:', error);
      throw error;
    } finally {
      setIsProcessing(false);
    }
  }, [context]);

  return {
    context,
    setContext,
    processBarcode,
    isProcessing,
    autoDetectContext,
  };
}

export function getSuggestedAction(context: BarcodeContext, result: any): string {
  // Smart action suggestion based on context and result
  switch (context) {
    case BarcodeContext.SALE:
      if (result?.stock > 0) {
        return 'addToCart';
      } else {
        return 'outOfStock';
      }
    case BarcodeContext.INVENTORY:
      if (result?.status === 'AVAILABLE') {
        return 'showDetails';
      } else if (result?.status === 'RESERVED') {
        return 'releaseReservation';
      } else {
        return 'showDetails';
      }
    case BarcodeContext.PURCHASE:
      return 'addToPurchase';
    case BarcodeContext.RETURN:
      return 'processReturn';
    case BarcodeContext.LOOKUP:
    default:
      return 'showDetails';
  }
}

export function getBarcodeContextAction(context: BarcodeContext, result: any): string {
  return getSuggestedAction(context, result);
}

export function getBarcodeContextLabel(context: BarcodeContext): string {
  switch (context) {
    case BarcodeContext.SALE:
      return 'بيع';
    case BarcodeContext.INVENTORY:
      return 'مخزون';
    case BarcodeContext.PURCHASE:
      return 'شراء';
    case BarcodeContext.RETURN:
      return 'مرتجع';
    case BarcodeContext.LOOKUP:
    default:
      return 'بحث';
  }
}

export function playScanSound(success: boolean = true) {
  try {
    const audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();
    const oscillator = audioContext.createOscillator();
    const gainNode = audioContext.createGain();
    
    oscillator.connect(gainNode);
    gainNode.connect(audioContext.destination);
    
    if (success) {
      // Success sound - high beep
      oscillator.frequency.value = 800;
      oscillator.type = 'sine';
      gainNode.gain.setValueAtTime(0.1, audioContext.currentTime);
      gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.1);
      oscillator.start(audioContext.currentTime);
      oscillator.stop(audioContext.currentTime + 0.1);
    } else {
      // Error sound - low buzz
      oscillator.frequency.value = 200;
      oscillator.type = 'square';
      gainNode.gain.setValueAtTime(0.1, audioContext.currentTime);
      gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.2);
      oscillator.start(audioContext.currentTime);
      oscillator.stop(audioContext.currentTime + 0.2);
    }
  } catch (error) {
    console.error('Audio playback failed:', error);
  }
}
