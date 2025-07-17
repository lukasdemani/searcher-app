import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import StatusBadge from './StatusBadge';

interface URLTableRowProps {
  url: URLAnalysis;
  isSelected: boolean;
  onToggleSelect: (id: number) => void;
  onDelete: (id: number) => void;
}
export type URLStatus = 'queued' | 'processing' | 'completed' | 'error';
export interface URLAnalysis {
  id: number;
  url: string;
  title?: string;
  html_version?: string;
  h1_count: number;
  h2_count: number;
  h3_count: number;
  h4_count: number;
  h5_count: number;
  h6_count: number;
  internal_links_count: number;
  external_links_count: number;
  broken_links_count: number;
  has_login_form: boolean;
  status: URLStatus;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export const URLTableRow: React.FC<URLTableRowProps> = React.memo(
  ({ url, isSelected, onToggleSelect, onDelete }) => {
    const { t } = useTranslation();

    return (
      <tr className={`hover:bg-gray-50 ${isSelected ? 'bg-blue-50' : ''}`}>
        <td className='px-6 py-4 whitespace-nowrap'>
          <input
            type='checkbox'
            checked={isSelected}
            onChange={() => onToggleSelect(url.id)}
            className='h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded'
          />
        </td>
        <td className='px-6 py-4 whitespace-nowrap'>
          <div className='max-w-xs'>
            <div className='text-sm font-medium text-gray-900 truncate'>
              {url.title || t('dashboard.table.noTitle')}
            </div>
            <div className='text-sm text-gray-500 truncate'>{url.url}</div>
          </div>
        </td>
        <td className='px-6 py-4 whitespace-nowrap'>
          <div className='text-sm text-gray-900'>
            {url.html_version || 'N/A'}
          </div>
        </td>
        <td className='px-6 py-4 whitespace-nowrap'>
          <div className='text-sm text-gray-900'>
            <div className='space-y-1'>
              <div className='flex justify-between'>
                <span>{t('dashboard.table.internal')}:</span>
                <span className='font-medium'>
                  {url.internal_links_count || 0}
                </span>
              </div>
              <div className='flex justify-between'>
                <span>{t('dashboard.table.external')}:</span>
                <span className='font-medium'>
                  {url.external_links_count || 0}
                </span>
              </div>
              {url.broken_links_count && url.broken_links_count > 0 && (
                <div className='flex justify-between text-red-600'>
                  <span>{t('dashboard.table.broken')}:</span>
                  <span className='font-medium'>{url.broken_links_count}</span>
                </div>
              )}
            </div>
          </div>
        </td>
        <td className='px-6 py-4 whitespace-nowrap'>
          <div className='text-sm text-gray-900'>
            <div className='grid grid-cols-2 gap-1 text-xs'>
              <div>H1: {url.h1_count || 0}</div>
              <div>H2: {url.h2_count || 0}</div>
              <div>H3: {url.h3_count || 0}</div>
              <div>H4: {url.h4_count || 0}</div>
              <div>H5: {url.h5_count || 0}</div>
              <div>H6: {url.h6_count || 0}</div>
            </div>
          </div>
        </td>
        <td className='px-6 py-4 whitespace-nowrap'>
          <span
            className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
              url.has_login_form
                ? 'bg-green-100 text-green-800'
                : 'bg-gray-100 text-gray-800'
            }`}
          >
            {url.has_login_form ? t('common.yes') : t('common.no')}
          </span>
        </td>
        <td className='px-6 py-4 whitespace-nowrap'>
          <StatusBadge status={url.status} />
          {url.error_message && (
            <div className='text-xs text-red-600 mt-1 truncate max-w-xs'>
              {url.error_message}
            </div>
          )}
        </td>
        <td className='px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2'>
          <Link
            to={`/url/${url.id}`}
            className='text-blue-600 hover:text-blue-900'
          >
            {t('dashboard.table.details')}
          </Link>
          <button
            onClick={() => onDelete(url.id)}
            className='text-red-600 hover:text-red-900 ml-2'
          >
            {t('dashboard.table.delete')}
          </button>
        </td>
      </tr>
    );
  }
);

URLTableRow.displayName = 'URLTableRow';

export default URLTableRow;
