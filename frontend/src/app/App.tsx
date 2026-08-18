import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { I18nextProvider } from 'react-i18next'
import i18n from '../i18n'
import { AppRouter } from './router'
import { ThemeProvider } from '../components/theme/ThemeProvider'
import { DirectionProvider } from '../components/theme/DirectionProvider'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
      staleTime: 5 * 60 * 1000, // 5 minutes
    },
  },
})

function App() {
  return (
    <ThemeProvider>
      <I18nextProvider i18n={i18n}>
        <DirectionProvider>
          <QueryClientProvider client={queryClient}>
            <BrowserRouter>
              <AppRouter />
            </BrowserRouter>
          </QueryClientProvider>
        </DirectionProvider>
      </I18nextProvider>
    </ThemeProvider>
  )
}

export default App
