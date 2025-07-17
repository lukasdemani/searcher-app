import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import type { URLAnalysis } from '../../types/types';
import StatusBadge from './StatusBadge';

interface MobileURLCardProps {
  url: URLAnalysis;
  isSelected: boolean;
  onToggleSelect: (id: number) => void;
  onDelete: (id: number) => void;
}

const MobileURLCard: React.FC<MobileURLCardProps> = ({
  url,
  isSelected,
  onToggleSelect,
  onDelete,
}) => {
  const { t } = useTranslation();

  return (
    <div className={`border-b border-gray-200 last:border-b-0 ${isSelected ? 'bg-blue-50' : ''}`}>
      <div className='px-4 py-4 space-y-3'>
        <div className='flex items-start justify-between'>
          <div className='flex items-center space-x-3 flex-1'>
            <input
              type='checkbox'
              checked={isSelected}
              onChange={() => onToggleSelect(url.id)}
              className='h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded'
            />
            <div className='flex-1 min-w-0'>
              <div className='text-sm font-medium text-gray-900 truncate'>
                {url.title || t('dashboard.table.noTitle')}
              </div>
              <div className='text-sm text-gray-500 truncate'>
                {url.url}
              </div>
            </div>
          </div>
          <StatusBadge status={url.status} />
        </div>
        
        <div className='grid grid-cols-2 gap-4 text-sm'>
          <div>
            <span className='text-gray-500'>{t('dashboard.table.htmlVersion')}:</span>
            <span className='ml-1 text-gray-900'>{url.html_version || 'N/A'}</span>
          </div>
          <div>
            <span className='text-gray-500'>{t('dashboard.table.loginForm')}:</span>
            <span className='ml-1 text-gray-900'>
              {url.has_login_form ? t('common.yes') : t('common.no')}
            </span>
          </div>
          <div>
            <span className='text-gray-500'>{t('dashboard.table.internal')}:</span>
            <span className='ml-1 text-gray-900'>{url.internal_links_count || 0}</span>
          </div>
          <div>
            <span className='text-gray-500'>{t('dashboard.table.external')}:</span>
            <span className='ml-1 text-gray-900'>{url.external_links_count || 0}</span>
          </div>
        </div>
        
        {url.broken_links_count && url.broken_links_count > 0 && (
          <div className='text-sm'>
            <span className='text-gray-500'>{t('dashboard.table.broken')}:</span>
            <span className='ml-1 text-red-600 font-medium'>{url.broken_links_count}</span>
          </div>
        )}
        
        <div className='flex justify-between items-center'>
          <div className='text-sm text-gray-500'>
            {t('dashboard.table.headingTags')}: {(url.h1_count || 0) + (url.h2_count || 0) + (url.h3_count || 0) + (url.h4_count || 0) + (url.h5_count || 0) + (url.h6_count || 0)}
          </div>
          <div className='flex space-x-2'>
            <Link
              to={`/url/${url.id}`}
              className='text-blue-600 hover:text-blue-900 text-sm font-medium'
            >
              {t('common.view')}
            </Link>
            <button
              onClick={() => onDelete(url.id)}
              className='text-red-600 hover:text-red-900 text-sm font-medium'
            >
              {t('common.delete')}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default MobileURLCard;
