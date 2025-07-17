import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import AddURLForm from '../components/features/urls/AddURLForm';
import BulkActions from '../components/features/urls/BulkActions';
import {
  EmptyStateIcon,
  ErrorIcon,
  NextIcon,
  PreviousIcon,
  ProcessingIcon,
  StatsIcon,
  SuccessIcon,
} from '../components/icons';
import Button from '../components/ui/Button';
import ColumnFilter from '../components/ui/ColumnFilter';
import Input from '../components/ui/Input';
import MobileURLCard from '../components/ui/MobileURLCard';
import Select from '../components/ui/Select';
import SortIcon from '../components/ui/SortIcon';
import StatCard from '../components/ui/StatCard';
import URLTableRow from '../components/ui/URLTableRow';
import { useURLs } from '../hooks/useURLs';

type SortField =
  | 'title'
  | 'url'
  | 'html_version'
  | 'internal_links_count'
  | 'external_links_count'
  | 'has_login_form'
  | 'created_at';
type SortDirection = 'asc' | 'desc';

interface ColumnFilters {
  title: string;
  url: string;
  html_version: string;
  internal_links_count: string;
  external_links_count: string;
  has_login_form: string;
  status: string;
}

const DashboardPage: React.FC = () => {
  const { t } = useTranslation();
  const [searchTerm, setSearchTerm] = useState('');
  const [debouncedSearchTerm, setDebouncedSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [sortField, setSortField] = useState<SortField>('created_at');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(10);
  
  const [columnFilters, setColumnFilters] = useState<ColumnFilters>({
    title: '',
    url: '',
    html_version: '',
    internal_links_count: '',
    external_links_count: '',
    has_login_form: '',
    status: 'all',
  });

  const [debouncedColumnFilters, setDebouncedColumnFilters] = useState<ColumnFilters>(columnFilters);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchTerm(searchTerm);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchTerm]);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedColumnFilters(columnFilters);
    }, 300);

    return () => clearTimeout(timer);
  }, [columnFilters]);

  const {
    urls,
    loading,
    selectedURLs,
    addURL,
    deleteURL,
    toggleSelect,
    clearSelection,
    refetch,
  } = useURLs({
    search: debouncedSearchTerm,
    status: statusFilter,
    autoRefresh: true,
    refreshInterval: 5000,
  });

  const handleBulkAnalyze = useCallback(async (ids: number[]) => {
    try {
      const response = await fetch('/api/urls/bulk-analyze', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': 'dev-api-key-2025',
        },
        body: JSON.stringify({ ids }),
      });
      
      if (!response.ok) {
        throw new Error('Failed to start bulk analysis');
      }
      
      await refetch();
    } catch (error) {
      console.error('Error in bulk analyze:', error);
      throw error;
    }
  }, [refetch]);

  const handleBulkDelete = useCallback(async (ids: number[]) => {
    try {
      const response = await fetch('/api/urls/bulk-delete', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': 'dev-api-key-2025',
        },
        body: JSON.stringify({ ids }),
      });
      
      if (!response.ok) {
        throw new Error('Failed to delete URLs');
      }
      
      await refetch();
    } catch (error) {
      console.error('Error in bulk delete:', error);
      throw error;
    }
  }, [refetch]);

  const handleSort = useCallback(
    (field: SortField) => {
      if (sortField === field) {
        setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
      } else {
        setSortField(field);
        setSortDirection('asc');
      }
    },
    [sortField, sortDirection]
  );

  const sortedUrls = useMemo(() => {
    if (!urls) return [];

    return [...urls].sort((a, b) => {
      let aValue: string | number | boolean | null | undefined = a[sortField];
      let bValue: string | number | boolean | null | undefined = b[sortField];

      if (aValue === null || aValue === undefined) aValue = '';
      if (bValue === null || bValue === undefined) bValue = '';

      if (typeof aValue === 'string' && typeof bValue === 'string') {
        aValue = aValue.toLowerCase();
        bValue = bValue.toLowerCase();
      }

      if (typeof aValue === 'boolean' && typeof bValue === 'boolean') {
        aValue = aValue ? 1 : 0;
        bValue = bValue ? 1 : 0;
      }

      if (sortDirection === 'asc') {
        return aValue > bValue ? 1 : aValue < bValue ? -1 : 0;
      } else {
        return aValue < bValue ? 1 : aValue > bValue ? -1 : 0;
      }
    });
  }, [urls, sortField, sortDirection]);

  const matchesNumberFilter = useCallback((value: number, filter: string): boolean => {
    if (!filter.trim()) return true;
    
    const trimmed = filter.trim();
    
    if (trimmed.startsWith('>=')) {
      const num = parseInt(trimmed.substring(2).trim());
      return !isNaN(num) && value >= num;
    }
    if (trimmed.startsWith('<=')) {
      const num = parseInt(trimmed.substring(2).trim());
      return !isNaN(num) && value <= num;
    }
    if (trimmed.startsWith('>')) {
      const num = parseInt(trimmed.substring(1).trim());
      return !isNaN(num) && value > num;
    }
    if (trimmed.startsWith('<')) {
      const num = parseInt(trimmed.substring(1).trim());
      return !isNaN(num) && value < num;
    }
    
    const num = parseInt(trimmed);
    return !isNaN(num) && value === num;
  }, []);

  const filteredUrls = useMemo(() => {
    if (!sortedUrls) return [];

    return sortedUrls.filter((url) => {
      const globalSearchMatch = !debouncedSearchTerm || 
        url.url.toLowerCase().includes(debouncedSearchTerm.toLowerCase()) ||
        url.title?.toLowerCase().includes(debouncedSearchTerm.toLowerCase()) ||
        url.html_version?.toLowerCase().includes(debouncedSearchTerm.toLowerCase()) ||
        url.status.toLowerCase().includes(debouncedSearchTerm.toLowerCase());

      const titleMatch = !debouncedColumnFilters.title || 
        url.title?.toLowerCase().includes(debouncedColumnFilters.title.toLowerCase());
      
      const urlMatch = !debouncedColumnFilters.url || 
        url.url.toLowerCase().includes(debouncedColumnFilters.url.toLowerCase());
      
      const htmlVersionMatch = !debouncedColumnFilters.html_version || 
        url.html_version?.toLowerCase().includes(debouncedColumnFilters.html_version.toLowerCase());
      
      const internalLinksMatch = !debouncedColumnFilters.internal_links_count || 
        matchesNumberFilter(url.internal_links_count, debouncedColumnFilters.internal_links_count);
      
      const externalLinksMatch = !debouncedColumnFilters.external_links_count || 
        matchesNumberFilter(url.external_links_count, debouncedColumnFilters.external_links_count);
      
      const loginFormMatch = !debouncedColumnFilters.has_login_form || 
        (debouncedColumnFilters.has_login_form === 'true' && url.has_login_form) ||
        (debouncedColumnFilters.has_login_form === 'false' && !url.has_login_form);
      
      const statusMatch = debouncedColumnFilters.status === 'all' || 
        url.status === debouncedColumnFilters.status;

      const legacyStatusMatch = statusFilter === 'all' || url.status === statusFilter;

      return globalSearchMatch && titleMatch && urlMatch && htmlVersionMatch && 
             internalLinksMatch && externalLinksMatch && loginFormMatch && 
             statusMatch && legacyStatusMatch;
    });
  }, [sortedUrls, debouncedSearchTerm, debouncedColumnFilters, statusFilter, matchesNumberFilter]);

  const paginationData = useMemo(() => {
    const totalPages = Math.ceil(filteredUrls?.length / itemsPerPage);
    const startIndex = (currentPage - 1) * itemsPerPage;
    const endIndex = startIndex + itemsPerPage;
    const paginatedUrls = filteredUrls?.slice(startIndex, endIndex);

    return {
      totalPages,
      startIndex,
      endIndex,
      paginatedUrls,
    };
  }, [filteredUrls, currentPage, itemsPerPage]);

  const { totalPages, startIndex, endIndex, paginatedUrls } = paginationData;

  const stats = useMemo(
    () => ({
      total: urls?.length || 0,
      processing:
        urls?.filter((url) => url.status === 'processing')?.length || 0,
      completed: urls?.filter((url) => url.status === 'completed')?.length || 0,
      failed: urls?.filter((url) => url.status === 'error')?.length || 0,
    }),
    [urls]
  );

  const handleSelectAll = useCallback(() => {
    if (selectedURLs.length === paginatedUrls?.length) {
      clearSelection();
    } else {
      paginatedUrls?.forEach((url) => {
        if (!selectedURLs.includes(url.id)) {
          toggleSelect(url.id);
        }
      });
    }
  }, [selectedURLs, paginatedUrls, clearSelection, toggleSelect]);

  const isAllSelected =
    paginatedUrls?.length > 0 && selectedURLs?.length === paginatedUrls?.length;
  const isSomeSelected =
    selectedURLs?.length > 0 && selectedURLs?.length < paginatedUrls?.length;

  const handleAddURL = useCallback(
    async (url: string) => {
      await addURL(url);
    },
    [addURL]
  );

  const statusOptions = useMemo(
    () => [
      { value: 'all', label: t('dashboard.filters.allStatuses') },
      { value: 'queued', label: t('dashboard.filters.queued') },
      { value: 'processing', label: t('dashboard.filters.processing') },
      { value: 'completed', label: t('dashboard.filters.completed') },
      { value: 'error', label: t('dashboard.filters.error') },
    ],
    [t]
  );

  const itemsPerPageOptions = useMemo(
    () => [
      { value: 5, label: '5' },
      { value: 10, label: '10' },
      { value: 25, label: '25' },
      { value: 50, label: '50' },
    ],
    []
  );

  const statCards = useMemo(
    () => [
      {
        icon: StatsIcon,
        bgColor: 'bg-blue-500',
        label: t('dashboard.stats.totalUrls'),
        value: stats.total,
      },
      {
        icon: ProcessingIcon,
        bgColor: 'bg-yellow-500',
        label: t('dashboard.stats.processing'),
        value: stats.processing,
      },
      {
        icon: SuccessIcon,
        bgColor: 'bg-green-500',
        label: t('dashboard.stats.completed'),
        value: stats.completed,
      },
      {
        icon: ErrorIcon,
        bgColor: 'bg-red-500',
        label: t('dashboard.stats.failed'),
        value: stats.failed,
      },
    ],
    [t, stats]
  );

  const createSortableColumn = useCallback(
    (field: SortField, label: string) => (
      <th
        key={field}
        className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100'
        onClick={() => handleSort(field)}
      >
        <div className='flex items-center space-x-1'>
          <span>{label}</span>
          <SortIcon
            field={field}
            currentSortField={sortField}
            sortDirection={sortDirection}
          />
        </div>
      </th>
    ),
    [handleSort, sortField, sortDirection]
  );

  const handleColumnFilterChange = useCallback((column: keyof ColumnFilters, value: string) => {
    setColumnFilters(prev => ({
      ...prev,
      [column]: value
    }));
    setCurrentPage(1);
  }, []);

  const clearAllFilters = useCallback(() => {
    setSearchTerm('');
    setStatusFilter('all');
    setColumnFilters({
      title: '',
      url: '',
      html_version: '',
      internal_links_count: '',
      external_links_count: '',
      has_login_form: '',
      status: 'all',
    });
    setCurrentPage(1);
  }, []);

  const hasActiveFilters = useMemo(() => {
    return searchTerm || statusFilter !== 'all' || 
           Object.values(columnFilters).some(filter => filter !== '' && filter !== 'all');
  }, [searchTerm, statusFilter, columnFilters]);

  return (
    <div className='min-h-screen bg-gray-50'>
      <div className='bg-white shadow'>
        <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8'>
          <div className='flex flex-col sm:flex-row sm:justify-between sm:items-center py-6 space-y-4 sm:space-y-0'>
            <div>
              <h1 className='text-2xl sm:text-3xl font-bold text-gray-900'>
                {t('dashboard.title')}
              </h1>
              <p className='mt-1 text-sm text-gray-500'>
                {t('dashboard.subtitle')}
                {stats.processing > 0 && (
                  <span className='ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800'>
                    {stats.processing} {t('dashboard.processing')}
                  </span>
                )}
              </p>
            </div>
            <div className='flex flex-col sm:flex-row space-y-2 sm:space-y-0 sm:space-x-3'>
              <Button
                variant='secondary'
                onClick={() => refetch()}
                loading={loading}
                className='w-full sm:w-auto'
              >
                {t('dashboard.refresh')}
              </Button>
              <AddURLForm onAddURL={handleAddURL} />
            </div>
          </div>
        </div>
      </div>

      <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8'>
        <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6 mb-6 sm:mb-8'>
          {statCards.map((card, index) => (
            <StatCard
              key={index}
              icon={card.icon}
              bgColor={card.bgColor}
              label={card.label}
              value={card.value}
            />
          ))}
        </div>

        <div className='bg-white shadow rounded-lg mb-6'>
          <BulkActions
            selectedURLs={selectedURLs}
            onClearSelection={clearSelection}
            onRefresh={refetch}
            onBulkAnalyze={handleBulkAnalyze}
            onBulkDelete={handleBulkDelete}
          />
          
          <div className='px-4 sm:px-6 py-4 border-b border-gray-200'>
            <div className='flex flex-col space-y-4 sm:space-y-3'>
              <div className='flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-3 sm:space-y-0'>
                <div className='flex-1 max-w-full sm:max-w-lg'>
                  <Input
                    placeholder={t('dashboard.filters.searchPlaceholder')}
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className='w-full'
                  />
                </div>
                <div className='flex flex-col sm:flex-row space-y-2 sm:space-y-0 sm:space-x-3'>
                  <Select
                    value={statusFilter}
                    onChange={(value) => setStatusFilter(value as string)}
                    options={statusOptions}
                    className='w-full sm:w-40'
                  />
                  <Select
                    value={itemsPerPage}
                    onChange={(value) => setItemsPerPage(value as number)}
                    options={itemsPerPageOptions}
                    className='w-full sm:w-20'
                  />
                  {hasActiveFilters && (
                    <Button
                      variant='secondary'
                      onClick={clearAllFilters}
                      className='w-full sm:w-auto px-3 py-2 text-sm'
                    >
                      {t('dashboard.filters.clearAll')}
                    </Button>
                  )}
                </div>
              </div>
            </div>
            {hasActiveFilters && (
              <div className='mt-3 flex flex-wrap gap-2'>
                {searchTerm && (
                  <span className='inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800'>
                    {t('dashboard.filters.globalSearch')}: "{searchTerm}"
                  </span>
                )}
                {statusFilter !== 'all' && (
                  <span className='inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800'>
                    {t('dashboard.filters.status')}: {statusFilter}
                  </span>
                )}
                {Object.entries(columnFilters).map(([key, value]) => 
                  value && value !== 'all' && (
                    <span key={key} className='inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800'>
                      {t(`dashboard.table.${key}`)}: {value}
                    </span>
                  )
                )}
              </div>
            )}
          </div>
        </div>

        <div className='bg-white shadow rounded-lg overflow-hidden'>
          <div className='px-4 sm:px-6 py-4 border-b border-gray-200'>
            <div className='flex flex-col sm:flex-row sm:justify-between sm:items-center space-y-2 sm:space-y-0'>
              <h3 className='text-lg leading-6 font-medium text-gray-900'>
                {t('dashboard.table.urls')} ({filteredUrls?.length})
              </h3>
              <div className='text-sm text-gray-500'>
                {t('dashboard.table.showing', {
                  start: startIndex + 1,
                  end: Math.min(endIndex, filteredUrls?.length),
                  total: filteredUrls?.length,
                })}
              </div>
            </div>
          </div>
          
          <div className='hidden lg:block overflow-x-auto'>
            <table className='min-w-full divide-y divide-gray-200'>
              <thead className='bg-gray-50'>
                <tr>
                  <th className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'>
                    <input
                      type='checkbox'
                      checked={isAllSelected}
                      ref={(input) => {
                        if (input) input.indeterminate = isSomeSelected;
                      }}
                      onChange={handleSelectAll}
                      className='h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded'
                    />
                  </th>
                  {createSortableColumn('title', t('dashboard.table.title'))}
                  {createSortableColumn(
                    'html_version',
                    t('dashboard.table.htmlVersion')
                  )}
                  {createSortableColumn(
                    'internal_links_count',
                    t('dashboard.table.links')
                  )}
                  <th className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'>
                    {t('dashboard.table.headingTags')}
                  </th>
                  {createSortableColumn(
                    'has_login_form',
                    t('dashboard.table.loginForm')
                  )}
                  <th className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'>
                    {t('dashboard.table.status')}
                  </th>
                  <th className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'>
                    {t('common.actions')}
                  </th>
                </tr>
                <tr className='bg-gray-100'>
                  <td className='px-6 py-2'></td>
                  <td className='px-6 py-2'>
                    <ColumnFilter
                      column="title"
                      value={columnFilters.title}
                      onChange={(value) => handleColumnFilterChange('title', value)}
                      type="text"
                    />
                  </td>
                  <td className='px-6 py-2'>
                    <ColumnFilter
                      column="html_version"
                      value={columnFilters.html_version}
                      onChange={(value) => handleColumnFilterChange('html_version', value)}
                      type="text"
                    />
                  </td>
                  <td className='px-6 py-2'>
                    <ColumnFilter
                      column="internal_links_count"
                      value={columnFilters.internal_links_count}
                      onChange={(value) => handleColumnFilterChange('internal_links_count', value)}
                      type="number"
                      placeholder="e.g., >5, <=10, 0"
                    />
                  </td>
                  <td className='px-6 py-2'>
                    <ColumnFilter
                      column="external_links_count"
                      value={columnFilters.external_links_count}
                      onChange={(value) => handleColumnFilterChange('external_links_count', value)}
                      type="number"
                      placeholder="e.g., >10, <=50"
                    />
                  </td>
                  <td className='px-6 py-2'>
                    <ColumnFilter
                      column="has_login_form"
                      value={columnFilters.has_login_form}
                      onChange={(value) => handleColumnFilterChange('has_login_form', value)}
                      type="boolean"
                    />
                  </td>
                  <td className='px-6 py-2'>
                    <ColumnFilter
                      column="status"
                      value={columnFilters.status}
                      onChange={(value) => handleColumnFilterChange('status', value)}
                      type="select"
                      options={statusOptions}
                    />
                  </td>
                  <td className='px-6 py-2'></td>
                </tr>
              </thead>
              <tbody className='bg-white divide-y divide-gray-200'>
                {paginatedUrls?.map((url) => (
                  <URLTableRow
                    key={url.id}
                    url={url}
                    isSelected={selectedURLs.includes(url.id)}
                    onToggleSelect={toggleSelect}
                    onDelete={deleteURL}
                  />
                ))}
              </tbody>
            </table>
          </div>

          <div className='lg:hidden'>
            {paginatedUrls?.map((url) => (
              <MobileURLCard
                key={url.id}
                url={url}
                isSelected={selectedURLs.includes(url.id)}
                onToggleSelect={toggleSelect}
                onDelete={deleteURL}
              />
            ))}
          </div>

          {totalPages > 1 && (
            <div className='px-4 sm:px-6 py-4 border-t border-gray-200'>
              <div className='flex items-center justify-between'>
                <div className='flex-1 flex justify-between sm:hidden'>
                  <button
                    onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                    disabled={currentPage === 1}
                    className='relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50'
                  >
                    {t('common.previous')}
                  </button>
                  <button
                    onClick={() =>
                      setCurrentPage(Math.min(totalPages, currentPage + 1))
                    }
                    disabled={currentPage === totalPages}
                    className='ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50'
                  >
                    {t('common.next')}
                  </button>
                </div>
                <div className='hidden sm:flex-1 sm:flex sm:items-center sm:justify-between'>
                  <div>
                    <p className='text-sm text-gray-700'>
                      {t('dashboard.pagination.page', {
                        current: currentPage,
                        total: totalPages,
                      })}
                    </p>
                  </div>
                  <div>
                    <nav className='relative z-0 inline-flex rounded-md shadow-sm -space-x-px'>
                      <button
                        onClick={() =>
                          setCurrentPage(Math.max(1, currentPage - 1))
                        }
                        disabled={currentPage === 1}
                        className='relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50'
                      >
                        <span className='sr-only'>{t('common.previous')}</span>
                        <PreviousIcon className='h-5 w-5' />
                      </button>

                      {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
                        const page = totalPages <= 5 ? i + 1 : 
                          currentPage <= 3 ? i + 1 :
                          currentPage >= totalPages - 2 ? totalPages - 4 + i :
                          currentPage - 2 + i;
                        return (
                          <button
                            key={page}
                            onClick={() => setCurrentPage(page)}
                            className={`relative inline-flex items-center px-3 sm:px-4 py-2 border text-sm font-medium ${
                              page === currentPage
                                ? 'z-10 bg-blue-50 border-blue-500 text-blue-600'
                                : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50'
                            }`}
                          >
                            {page}
                          </button>
                        );
                      })}

                      <button
                        onClick={() =>
                          setCurrentPage(Math.min(totalPages, currentPage + 1))
                        }
                        disabled={currentPage === totalPages}
                        className='relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50'
                      >
                        <span className='sr-only'>{t('common.next')}</span>
                        <NextIcon className='h-5 w-5' />
                      </button>
                    </nav>
                  </div>
                </div>
              </div>
            </div>
          )}

          {filteredUrls?.length === 0 && (
            <div className='text-center py-12'>
              <EmptyStateIcon className='mx-auto h-12 w-12 text-gray-400' />
              <h3 className='mt-2 text-sm font-medium text-gray-900'>
                {t('dashboard.empty.title')}
              </h3>
              <p className='mt-1 text-sm text-gray-500'>
                {searchTerm || statusFilter !== 'all'
                  ? t('dashboard.empty.noResults')
                  : t('dashboard.empty.noUrls')}
              </p>
              {!searchTerm && statusFilter === 'all' && (
                <div className='mt-6'>
                  <AddURLForm onAddURL={handleAddURL} />
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default DashboardPage;
