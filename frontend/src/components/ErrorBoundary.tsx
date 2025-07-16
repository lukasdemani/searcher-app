import { Component, ErrorInfo, ReactNode } from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { ExclamationTriangleIcon } from './icons';
import Button from './ui/Button';

interface Props extends WithTranslation {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
  errorInfo?: ErrorInfo;
}

class ErrorBoundaryComponent extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.setState({
      error,
      errorInfo,
    });

    if (process.env.NODE_ENV === 'development') {
      console.error('ErrorBoundary caught an error:', error, errorInfo);
    }
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: undefined, errorInfo: undefined });
  };

  render() {
    const { t } = this.props;

    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
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
              <Button
                onClick={this.handleRetry}
                variant='primary'
                className='w-full'
              >
                {t('errorBoundary.retryButton')}
              </Button>

              <Button
                onClick={() => window.location.reload()}
                variant='secondary'
                className='w-full'
              >
                {t('errorBoundary.reloadButton')}
              </Button>
            </div>

            {process.env.NODE_ENV === 'development' && this.state.error && (
              <details className='mt-6 text-left'>
                <summary className='cursor-pointer text-sm font-medium text-gray-700 hover:text-gray-900'>
                  {t('errorBoundary.errorDetails')}
                </summary>
                <div className='mt-2 p-3 bg-gray-100 rounded text-xs'>
                  <div className='font-semibold text-red-600 mb-2'>
                    {this.state.error.toString()}
                  </div>
                  <pre className='whitespace-pre-wrap text-gray-600'>
                    {this.state.errorInfo?.componentStack}
                  </pre>
                </div>
              </details>
            )}
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export const ErrorBoundary = withTranslation()(ErrorBoundaryComponent);
export default ErrorBoundary;
