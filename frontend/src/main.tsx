import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { Toaster } from 'sonner'
import { AlertTriangle, CheckCircle2, Info, XCircle } from 'lucide-react'
import '@fontsource/noto-naskh-arabic/arabic.css'
import './index.css'
import './i18n/config'
import App from './App.tsx'
import { appConfig } from './lib/config/app'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    {appConfig.developmentMode && (
      <div
        role="status"
        className="relative z-[200] bg-amber-300 px-4 py-2 text-center text-sm font-bold text-amber-950"
      >
        وضع التطوير نشط — تأكد من الخادم وقاعدة البيانات قبل إجراء تغييرات تجارية.
      </div>
    )}
    <App />
    <Toaster
      position="top-right"
      richColors
      closeButton
      duration={4200}
      icons={{
        success: <CheckCircle2 className="h-5 w-5" />,
        error: <XCircle className="h-5 w-5" />,
        warning: <AlertTriangle className="h-5 w-5" />,
        info: <Info className="h-5 w-5" />,
      }}
      toastOptions={{
        classNames: {
          toast: 'pf-sonner-toast',
          title: 'pf-sonner-title',
          description: 'pf-sonner-description',
          success: 'pf-sonner-success',
          error: 'pf-sonner-error',
          warning: 'pf-sonner-warning',
          info: 'pf-sonner-info',
          closeButton: 'pf-sonner-close',
        },
      }}
    />
  </StrictMode>,
)
