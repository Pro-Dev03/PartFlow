import { Component, type ReactNode } from 'react';
import { Card, CardContent } from './card';
import { Button } from './button';
import { AlertTriangle, RefreshCw } from 'lucide-react';
import { getArabicErrorMessage } from '../../lib/error-messages';

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: any) {
    console.error('Error caught by boundary:', error, errorInfo);
    
    // Log error details for debugging
    const errorDetails = {
      message: error.message,
      stack: error.stack,
      componentStack: errorInfo.componentStack,
      timestamp: new Date().toISOString(),
    };
    
    // Store in localStorage for debugging
    try {
      const recentErrors = JSON.parse(localStorage.getItem('recent_errors') || '[]');
      recentErrors.push(errorDetails);
      // Keep only last 5 errors
      if (recentErrors.length > 5) {
        recentErrors.shift();
      }
      localStorage.setItem('recent_errors', JSON.stringify(recentErrors));
    } catch (e) {
      console.error('Failed to store error:', e);
    }
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
  };

  handleReload = () => {
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      const arabicMessage = getArabicErrorMessage(this.state.error);

      return (
        <div className="flex items-center justify-center min-h-screen bg-gray-50 dark:bg-gray-900 p-4">
          <Card className="max-w-md w-full">
            <CardContent className="p-6 text-center">
              <AlertTriangle className="w-16 h-16 text-red-500 mx-auto mb-4" />
              <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">
                حدث خطأ
              </h2>
              <p className="text-gray-600 dark:text-gray-400 mb-6 text-sm leading-relaxed">
                {arabicMessage}
              </p>
              
              {/* Show technical details in development */}
              {import.meta.env.DEV && this.state.error?.message && (
                <div className="mb-6 p-3 bg-red-50 dark:bg-red-900/20 rounded-lg text-left">
                  <p className="text-xs text-red-800 dark:text-red-300 font-mono">
                    {this.state.error.message}
                  </p>
                </div>
              )}
              
              <div className="flex flex-col gap-3">
                <Button
                  onClick={this.handleReset}
                  className="w-full gap-2"
                  variant="outline"
                >
                  <RefreshCw className="w-4 h-4" />
                  محاولة مرة أخرى
                </Button>
                <Button
                  onClick={this.handleReload}
                  className="w-full"
                >
                  إعادة تحميل الصفحة
                </Button>
              </div>
              
              <p className="text-xs text-gray-400 dark:text-gray-500 mt-4">
                إذا استمرت المشكلة، يرجى التواصل مع الدعم الفني
              </p>
            </CardContent>
          </Card>
        </div>
      );
    }

    return this.props.children;
  }
}