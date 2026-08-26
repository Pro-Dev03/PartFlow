import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { ItemInputMethod, type ItemInputMethodType } from '../../../components/ui/item-input-method';
import { CameraScanner } from '../../../components/ui/camera-scanner';
import { Camera, Plus } from 'lucide-react';
import { toast } from 'sonner';

interface InventoryScannerProps {
  inputMethod: ItemInputMethodType;
  setInputMethod: (method: ItemInputMethodType) => void;
  barcodeInput: string;
  setBarcodeInput: (value: string) => void;
  onBarcodeScan: (e: React.FormEvent) => void;
  onCameraScan: (barcode: string) => void;
  onManualAdd: () => void;
  isCameraScannerOpen: boolean;
  onCameraOpen: () => void;
  onCameraClose: () => void;
}

export function InventoryScanner({
  inputMethod,
  setInputMethod,
  barcodeInput,
  setBarcodeInput,
  onBarcodeScan,
  onCameraScan,
  onManualAdd,
  isCameraScannerOpen,
  onCameraOpen,
  onCameraClose,
}: InventoryScannerProps) {
  return (
    <>
      <Card style={{
        background: 'linear-gradient(135deg, var(--color-primary-08) 0%, var(--color-info-08) 100%)',
        border: '1px solid var(--color-primary-15)',
        backdropFilter: 'blur(10px)',
        transition: 'all 0.3s ease'
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.background = 'linear-gradient(135deg, var(--color-primary-12) 0%, var(--color-info-12) 100%)';
        e.currentTarget.style.borderColor = 'var(--color-primary-25)';
        e.currentTarget.style.transform = 'translateY(-2px)';
        e.currentTarget.style.boxShadow = '0 8px 25px var(--color-primary-15)';
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.background = 'linear-gradient(135deg, var(--color-primary-08) 0%, var(--color-info-08) 100%)';
        e.currentTarget.style.borderColor = 'var(--color-primary-15)';
        e.currentTarget.style.transform = 'translateY(0)';
        e.currentTarget.style.boxShadow = 'none';
      }}>
        <CardHeader>
          <CardTitle style={{ 
            fontSize: '15px',
            fontWeight: '600',
            color: 'var(--text-primary)'
          }}>
            إضافة قطع للمخزون
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ItemInputMethod 
            onMethodChange={setInputMethod}
            defaultMethod="barcode"
          />
          
          {inputMethod === 'barcode' && (
            <form onSubmit={onBarcodeScan} className="mt-4">
              <div style={{ display: 'flex', gap: '12px', alignItems: 'stretch' }}>
                <div style={{ 
                  flex: 1, 
                  position: 'relative',
                  display: 'flex',
                  alignItems: 'center'
                }}>
                  <Input
                    placeholder="مسح أو أدخل الباركود..."
                    value={barcodeInput}
                    onChange={(e) => setBarcodeInput(e.target.value)}
                    style={{
                      fontSize: '14px',
                      fontWeight: '500',
                      letterSpacing: '0.3px',
                      background: 'var(--bg-surface-elevated)',
                      border: '1px solid var(--color-primary-20)',
                      transition: 'all 0.3s ease'
                    }}
                    className="hover:border-indigo-400/50 focus:border-indigo-400/70 focus:shadow-lg focus:shadow-indigo-500/10"
                  />
                </div>
                <Button 
                  type="submit" 
                  variant="primary"
                  style={{
                    minWidth: '56px',
                    height: '36px',
                    fontSize: '12px',
                    fontWeight: '600',
                    letterSpacing: '0.2px',
                    background: 'linear-gradient(135deg, var(--button-primary-bg) 0%, var(--color-primary-85) 100%)',
                    border: '1px solid var(--color-primary-25)',
                    boxShadow: '0 2px 8px var(--color-primary-15)',
                    transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
                    borderRadius: '8px'
                  }}
                >
                  بحث
                </Button>
              </div>
            </form>
          )}

          {inputMethod === 'camera' && (
            <div className="mt-4">
              <Button
                variant="primary"
                onClick={onCameraOpen}
                className="w-full"
                style={{
                  height: '42px',
                  fontSize: '13px',
                  fontWeight: '600',
                  letterSpacing: '0.3px',
                  background: 'linear-gradient(135deg, var(--button-primary-bg) 0%, var(--color-primary-85) 100%)',
                  border: '1px solid var(--color-primary-25)',
                  boxShadow: '0 2px 8px var(--color-primary-15)',
                  transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
                  borderRadius: '8px'
                }}
              >
                <Camera className="w-4 h-4 mr-2" />
                فتح الكاميرا للمسح
              </Button>
            </div>
          )}

          {inputMethod === 'manual' && (
            <div className="mt-4">
              <Button
                variant="primary"
                onClick={onManualAdd}
                className="w-full"
                style={{
                  height: '42px',
                  fontSize: '13px',
                  fontWeight: '600',
                  letterSpacing: '0.3px',
                  background: 'linear-gradient(135deg, var(--button-primary-bg) 0%, var(--color-primary-85) 100%)',
                  border: '1px solid var(--color-primary-25)',
                  boxShadow: '0 2px 8px var(--color-primary-15)',
                  transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
                  borderRadius: '8px'
                }}
              >
                <Plus className="w-4 h-4 mr-2" />
                إضافة قطعة يدوياً
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Camera Scanner Modal */}
      {isCameraScannerOpen && (
        <CameraScanner
          onScan={onCameraScan}
          onClose={onCameraClose}
          onError={(error) => console.error('Camera scan error:', error)}
        />
      )}
    </>
  );
}