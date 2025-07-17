import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useLocation } from 'react-router-dom';
import { NavigationIcon } from '../icons';
import LanguageSelector from '../ui/LanguageSelector';

interface LayoutProps {
  children: React.ReactNode;
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
  const location = useLocation();
  const { t } = useTranslation();

  const isActivePath = (path: string) => {
    return (
      location.pathname === path || location.pathname.startsWith(path + '/')
    );
  };

  return (
    <div className='min-h-screen bg-gray-50'>
      <nav className='bg-white shadow-sm border-b border-gray-200'>
        <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8'>
          <div className='flex justify-between items-center h-16'>
            <div className='flex items-center min-w-0 flex-1'>
              <Link to='/' className='flex-shrink-0 flex items-center'>
                <div className='h-8 w-8 bg-blue-600 rounded-lg flex items-center justify-center'>
                  <NavigationIcon className='h-5 w-5 text-white' />
                </div>
                <span className='ml-2 text-lg sm:text-xl font-bold text-gray-900 truncate'>
                  {t('header.title')}
                </span>
              </Link>

              <div className='hidden md:ml-6 md:flex md:space-x-8'>
                <Link
                  to='/'
                  className={`inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium ${
                    isActivePath('/')
                      ? 'border-blue-500 text-gray-900'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }`}
                >
                  {t('header.dashboard')}
                </Link>
              </div>
            </div>

            <div className='flex items-center space-x-2 md:space-x-4 flex-shrink-0'>
              <span className='hidden md:inline text-sm text-gray-500'>
                {t('header.subtitle')}
              </span>
              <LanguageSelector />
            </div>
          </div>
        </div>
      </nav>

      <main className='flex-1'>{children}</main>
    </div>
  );
};
