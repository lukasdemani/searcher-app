import React, {
  Component,
  useCallback,
  useState,
} from 'react';
import type { ErrorInfo, ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { ExclamationTriangleIcon } from './icons';
import Button from './ui/Button';

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
  errorInfo?: ErrorInfo;
}

class ErrorBoundaryClass extends Component<
  ErrorBoundaryProps & {
    onError: (error: Error, errorInfo: ErrorInfo) => void;
  },
  ErrorBoundaryState
> {
  constructor(
    props: ErrorBoundaryProps & {
      onError: (error: Error, errorInfo: ErrorInfo) => void;
    }
  ) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.setState({
      error,
      errorInfo,
    });

    this.props.onError(error, errorInfo);

    if (process.env.NODE_ENV === 'development') {
      console.error('ErrorBoundary caught an error:', error, errorInfo);
    }
  }

  render() {
    if (this.state.hasError) {
      return this.props.children;
    }

    return this.props.children;
  }
}

const ErrorBoundary: React.FC<ErrorBoundaryProps> = ({
  children,
  fallback,
}) => {
  const { t } = useTranslation();
  const [errorState, setErrorState] = useState<ErrorBoundaryState>({
    hasError: false,
  });

  const handleError = useCallback((error: Error, errorInfo: ErrorInfo) => {
    setErrorState({ hasError: true, error, errorInfo });
  }, []);

  const handleRetry = useCallback(() => {
    setErrorState({ hasError: false, error: undefined, errorInfo: undefined });
  }, []);

  const handleReload = useCallback(() => {
    window.location.reload();
  }, []);

  if (errorState.hasError) {
    if (fallback) {
      return <>{fallback}</>;
    }

    return (
      <div className='min-h-screen bg-gray-50 flex items-center justify-center px-4'>
        <div className='max-w-md w-full bg-white rounded-lg shadow-lg p-6 text-center'>
          <div className='mb-4'>
            <div className='mx-auto h-12 w-12 bg-red-100 rounded-full flex items-center justify-center'>
              <ExclamationTriangleIcon className='h-6 w-6 text-red-600' />
            </div>
          </div>

          <h3 className='text-lg font-medium text-gray-900 mb-2'>
            {t('errorBoundary.title')}
          </h3>

          <p className='text-sm text-gray-500 mb-6'>
            {t('errorBoundary.message')}
          </p>

          <div className='space-y-3'>
            <Button onClick={handleRetry} variant='primary' className='w-full'>
              {t('errorBoundary.retryButton')}
            </Button>

            <Button
              onClick={handleReload}
              variant='secondary'
              className='w-full'
            >
              {t('errorBoundary.reloadButton')}
            </Button>
          </div>

          {process.env.NODE_ENV === 'development' && errorState.error && (
            <details className='mt-6 text-left'>
              <summary className='cursor-pointer text-sm font-medium text-gray-700 hover:text-gray-900'>
                {t('errorBoundary.errorDetails')}
              </summary>
              <div className='mt-2 p-3 bg-gray-100 rounded text-xs'>
                <div className='font-semibold text-red-600 mb-2'>
                  {errorState.error.toString()}
                </div>
                <pre className='whitespace-pre-wrap text-gray-600'>
                  {errorState.errorInfo?.componentStack}
                </pre>
              </div>
            </details>
          )}
        </div>
      </div>
    );
  }

  return (
    <ErrorBoundaryClass onError={handleError}>{children}</ErrorBoundaryClass>
  );
};

export default ErrorBoundary;
