import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { Toaster } from 'sonner'
import { AlertTriangle, CheckCircle2, Info, XCircle } from 'lucide-react'
import './index.css'
import './design-system/styles/visual-refresh.css'
import './i18n/config'
import App from './App.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
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
