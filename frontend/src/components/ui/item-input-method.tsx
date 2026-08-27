import { useState } from 'react';
import { cn } from '../../utils';
import { Barcode, Camera, Keyboard } from 'lucide-react';

interface ItemInputMethodProps {
  onMethodChange: (method: 'barcode' | 'camera') => void;
  defaultMethod?: 'barcode' | 'camera';
  disabled?: boolean;
}

export function ItemInputMethod({ 
  onMethodChange, 
  defaultMethod = 'barcode',
  disabled = false 
}: ItemInputMethodProps) {
  const [activeMethod, setActiveMethod] = useState(defaultMethod);

  const methods = [
    {
      id: 'barcode' as const,
      label: 'مسح باركود',
      icon: Barcode,
      description: 'استخدام ماسح الباركود',
      color: 'text-cyan-400'
    },
    {
      id: 'camera' as const,
      label: 'كاميرا',
      icon: Camera,
      description: 'مسح عبر الكاميرا',
      color: 'text-green-400'
    }
  ];

  const handleMethodChange = (method: 'barcode' | 'camera') => {
    setActiveMethod(method);
    onMethodChange(method);
  };

  return (
    <div className="w-full">
      <div className="flex items-center gap-2 mb-3">
        <Keyboard className="w-4 h-4 text-text-muted/50" />
        <span className="text-xs font-medium text-text-muted/70">
          طريقة الإضافة
        </span>
      </div>
      
      <div className="grid grid-cols-2 gap-2">
        {methods.map((method) => {
          const Icon = method.icon;
          const isActive = activeMethod === method.id;
          
          return (
            <button
              key={method.id}
              onClick={() => !disabled && handleMethodChange(method.id)}
              disabled={disabled}
              className={cn(
                'pf-input-method flex flex-col items-center gap-2 p-3 rounded-xl border transition-all duration-300',
                'hover:-translate-y-0.5 hover:shadow-md',
                'disabled:opacity-50 disabled:cursor-not-allowed',
                isActive
                  ? 'bg-gradient-to-br from-cyan-500/20 to-purple-500/20 border-cyan-500/50 shadow-lg shadow-cyan-500/20'
                  : 'bg-surface/50 border-border-default hover:border-cyan-500/30'
              )}
            >
              <div className={cn(
                'p-2 rounded-lg transition-all duration-300',
                isActive ? 'bg-cyan-500/20' : 'bg-surface/30'
              )}>
                <Icon 
                  className={cn(
                    'w-5 h-5 transition-all duration-300',
                    isActive ? method.color : 'text-text-muted/50',
                    isActive && 'scale-110'
                  )} 
                />
              </div>
              <div className="text-center">
                <span className={cn(
                  'text-xs font-medium block transition-all duration-300',
                  isActive ? 'text-text-primary' : 'text-text-muted/70'
                )}>
                  {method.label}
                </span>
                <span className={cn(
                  'text-[10px] block transition-all duration-300',
                  isActive ? 'text-text-muted/60' : 'text-text-muted/40'
                )}>
                  {method.description}
                </span>
              </div>
            </button>
          );
        })}
      </div>
    </div>
  );
}

export type ItemInputMethodType = 'barcode' | 'camera';