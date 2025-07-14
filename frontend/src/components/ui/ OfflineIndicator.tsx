import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

interface OfflineIndicatorProps {
  className?: string;
}

const OfflineIndicator: React.FC<OfflineIndicatorProps> = ({
  className = '',
}) => {
  const { t } = useTranslation();
  const [isOnline, setIsOnline] = useState(navigator.onLine);
  const [showOfflineMessage, setShowOfflineMessage] = useState(false);

  useEffect(() => {
    const handleOnline = () => {
      setIsOnline(true);
      setShowOfflineMessage(false);
    };

    const handleOffline = () => {
      setIsOnline(false);
      setShowOfflineMessage(true);
    };

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    if (!navigator.onLine) {
      setShowOfflineMessage(true);
    }

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  useEffect(() => {
    if (isOnline && showOfflineMessage) {
      const timer = setTimeout(() => {
        setShowOfflineMessage(false);
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [isOnline, showOfflineMessage]);

  if (!showOfflineMessage) {
    return null;
  }

  return (
    <div className={`fixed top-0 left-0 right-0 z-50 ${className}`}>
      <div
        className={`
        px-4 py-3 text-sm font-medium text-center transition-all duration-300
        ${isOnline ? 'bg-green-600 text-white' : 'bg-yellow-600 text-white'}
      `}
      >
        <div className='flex items-center justify-center space-x-2'>
          <div
            className={`
            w-2 h-2 rounded-full
            ${isOnline ? 'bg-green-200' : 'bg-yellow-200 animate-pulse'}
          `}
          />
          <span>
            {isOnline
              ? t('offline.backOnline', "You're back online!")
              : t(
                  'offline.usingCachedData',
                  "You're offline - showing cached data"
                )}
          </span>
          {!isOnline && (
            <button
              onClick={() => window.location.reload()}
              className='ml-2 px-2 py-1 text-xs bg-yellow-700 hover:bg-yellow-800 rounded transition-colors'
            >
              {t('offline.retry', 'Retry')}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default OfflineIndicator;
