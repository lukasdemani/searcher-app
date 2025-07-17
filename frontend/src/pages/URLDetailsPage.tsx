import React, { useCallback, useEffect, useState } from 'react';
import { toast } from 'react-hot-toast';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate, useParams } from 'react-router-dom';
import LinksChart from '../components/charts/LinksChart';
import HeadingsChart from '../components/charts/HeadingsChart';
import { ChevronLeftIcon, NoResultsIcon } from '../components/icons';
import Button from '../components/ui/Button';
import StatusBadge from '../components/ui/StatusBadge';
import APIService from '../services/api';
import type { BrokenLink, URLAnalysis } from '../types/types';

const URLDetailsPage: React.FC = () => {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [url, setUrl] = useState<URLAnalysis | null>(null);
  const [brokenLinks, setBrokenLinks] = useState<BrokenLink[]>([]);
  const [loading, setLoading] = useState(true);
  const [reanalyzing, setReanalyzing] = useState(false);

  const fetchURLDetails = useCallback(async (urlId: string) => {
    try {
      setLoading(true);
      const [urlData, brokenLinksData] = await Promise.all([
        APIService.getURL(Number(urlId)),
        APIService.getBrokenLinks(Number(urlId)),
      ]);
      setUrl(urlData);
      setBrokenLinks(brokenLinksData);
    } catch {
      toast.error(t('messages.errorLoadingDetails'));
      navigate('/dashboard');
    } finally {
      setLoading(false);
    }
  }, [navigate, t]);

  useEffect(() => {
    if (id) {
      fetchURLDetails(id);
    }
  }, [id, fetchURLDetails]);

  const handleReanalyze = async () => {
    if (!url) return;

    try {
      setReanalyzing(true);
      await APIService.analyzeURL(url.id);
      toast.success(t('messages.reanalysisStarted'));
      setTimeout(() => fetchURLDetails(String(url.id)), 2000);
    } catch {
      toast.error(t('messages.errorStartingReanalysis'));
    } finally {
      setReanalyzing(false);
    }
  };

  if (loading) {
    return (
      <div className='min-h-screen bg-gray-50 flex items-center justify-center'>
        <div className='text-center'>
          <div className='animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto'></div>
          <p className='mt-2 text-gray-600'>{t('urlDetails.loadingDetails')}</p>
        </div>
      </div>
    );
  }

  if (!url) {
    return (
      <div className='min-h-screen bg-gray-50 flex items-center justify-center'>
        <div className='text-center'>
          <h1 className='text-2xl font-bold text-gray-900'>
            {t('urlDetails.notFound')}
          </h1>
          <Link to='/dashboard' className='text-blue-600 hover:text-blue-800'>
            {t('common.back')}
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className='min-h-screen bg-gray-50'>
      <div className='bg-white shadow'>
        <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8'>
          <div className='flex flex-col sm:flex-row sm:justify-between sm:items-center py-6 space-y-4 sm:space-y-0'>
            <div className='flex items-center space-x-4'>
              <Link
                to='/dashboard'
                className='text-gray-500 hover:text-gray-700'
              >
                <ChevronLeftIcon className='w-6 h-6' />
              </Link>
              <div className='min-w-0 flex-1'>
                <h1 className='text-xl sm:text-2xl font-bold text-gray-900'>
                  {t('urlDetails.title')}
                </h1>
                <p className='text-sm text-gray-600 truncate'>
                  {url.url}
                </p>
              </div>
            </div>
            <div className='flex flex-col sm:flex-row space-y-2 sm:space-y-0 sm:space-x-3'>
              <Button
                variant='secondary'
                onClick={handleReanalyze}
                loading={reanalyzing}
                className='w-full sm:w-auto'
              >
                {t('urlDetails.reanalyze')}
              </Button>
              <Link to={url.url} target='_blank' rel='noopener noreferrer'>
                <Button variant='primary' className='w-full sm:w-auto'>
                  {t('urlDetails.visitSite')}
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </div>

      <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8'>
        <div className='grid grid-cols-1 lg:grid-cols-3 gap-4 lg:gap-6 mb-6 lg:mb-8'>
          <div className='bg-white p-4 sm:p-6 rounded-lg shadow'>
            <div className='flex items-center justify-between'>
              <div>
                <p className='text-sm font-medium text-gray-600'>
                  {t('urlDetails.summary.status')}
                </p>
                <div className='mt-1'>
                  <StatusBadge status={url.status} />
                </div>
              </div>
              <div className='text-right'>
                <p className='text-xs sm:text-sm text-gray-500'>
                  {url.updated_at
                    ? t('urlDetails.summary.analyzedAt', {
                        date: new Date(url.updated_at).toLocaleString(),
                      })
                    : t('urlDetails.summary.neverAnalyzed')}
                </p>
              </div>
            </div>
          </div>

          <div className='bg-white p-4 sm:p-6 rounded-lg shadow'>
            <div>
              <p className='text-sm font-medium text-gray-600'>
                {t('urlDetails.summary.basicInfo')}
              </p>
              <div className='mt-2 space-y-1'>
                <p className='text-sm text-gray-900'>
                  <span className='font-medium'>
                    {t('urlDetails.summary.title')}:
                  </span>{' '}
                  <span className='break-words'>{url.title || 'N/A'}</span>
                </p>
                <p className='text-sm text-gray-900'>
                  <span className='font-medium'>
                    {t('urlDetails.summary.htmlVersion')}:
                  </span>{' '}
                  {url.html_version || 'N/A'}
                </p>
                <p className='text-sm text-gray-900'>
                  <span className='font-medium'>
                    {t('urlDetails.summary.loginForm')}:
                  </span>{' '}
                  {url.has_login_form ? t('common.yes') : t('common.no')}
                </p>
              </div>
            </div>
          </div>

          <div className='bg-white p-4 sm:p-6 rounded-lg shadow'>
            <div>
              <p className='text-sm font-medium text-gray-600'>
                {t('urlDetails.summary.linksSummary')}
              </p>
              <div className='mt-2 space-y-1'>
                <p className='text-sm text-gray-900'>
                  <span className='font-medium'>
                    {t('urlDetails.summary.total')}:
                  </span>{' '}
                  {(url.internal_links_count || 0) +
                    (url.external_links_count || 0)}
                </p>
                <p className='text-sm text-green-600'>
                  <span className='font-medium'>
                    {t('urlDetails.summary.internal')}:
                  </span>{' '}
                  {url.internal_links_count || 0}
                </p>
                <p className='text-sm text-blue-600'>
                  <span className='font-medium'>
                    {t('urlDetails.summary.external')}:
                  </span>{' '}
                  {url.external_links_count || 0}
                </p>
                {url.broken_links_count && url.broken_links_count > 0 && (
                  <p className='text-sm text-red-600'>
                    <span className='font-medium'>
                      {t('urlDetails.summary.broken')}:
                    </span>{' '}
                    {url.broken_links_count}
                  </p>
                )}
              </div>
            </div>
          </div>
        </div>

        <div className='grid grid-cols-1 xl:grid-cols-2 gap-6 lg:gap-8 mb-6 lg:mb-8'>
          <div className='bg-white p-4 sm:p-6 rounded-lg shadow'>
            <h3 className='text-lg font-medium text-gray-900 mb-4'>
              {t('urlDetails.charts.linksDistribution')}
            </h3>
            <div className='h-64 sm:h-80'>
              <LinksChart 
                internalLinks={url.internal_links_count}
                externalLinks={url.external_links_count}
                chartType="donut"
              />
            </div>
          </div>

          <div className='bg-white p-4 sm:p-6 rounded-lg shadow'>
            <h3 className='text-lg font-medium text-gray-900 mb-4'>
              {t('urlDetails.charts.headingsDistribution')}
            </h3>
            <div className='h-64 sm:h-80'>
              <HeadingsChart 
                h1Count={url.h1_count}
                h2Count={url.h2_count}
                h3Count={url.h3_count}
                h4Count={url.h4_count}
                h5Count={url.h5_count}
                h6Count={url.h6_count}
              />
            </div>
          </div>
        </div>

        {brokenLinks.length > 0 && (
          <div className='bg-white rounded-lg shadow overflow-hidden'>
            <div className='px-4 sm:px-6 py-4 border-b border-gray-200'>
              <h3 className='text-lg font-medium text-gray-900'>
                {t('urlDetails.brokenLinks.title', {
                  count: brokenLinks.length,
                })}
              </h3>
            </div>
            
            <div className='hidden sm:block overflow-x-auto'>
              <table className='min-w-full divide-y divide-gray-200'>
                <thead className='bg-gray-50'>
                  <tr>
                    <th className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'>
                      {t('urlDetails.brokenLinks.url')}
                    </th>
                    <th className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'>
                      {t('urlDetails.brokenLinks.statusCode')}
                    </th>
                    <th className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'>
                      {t('urlDetails.brokenLinks.type')}
                    </th>
                    <th className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'>
                      {t('common.actions')}
                    </th>
                  </tr>
                </thead>
                <tbody className='bg-white divide-y divide-gray-200'>
                  {brokenLinks.map((link, index) => (
                    <tr key={index} className='hover:bg-gray-50'>
                      <td className='px-6 py-4 whitespace-nowrap'>
                        <div className='text-sm text-gray-900 truncate max-w-xs'>
                          {link.link_url}
                        </div>
                      </td>
                      <td className='px-6 py-4 whitespace-nowrap'>
                        <span
                          className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            link.status_code >= 400 && link.status_code < 500
                              ? 'bg-yellow-100 text-yellow-800'
                              : 'bg-red-100 text-red-800'
                          }`}
                        >
                          {link.status_code}
                        </span>
                      </td>
                      <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-900'>
                        <span
                          className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800`}
                        >
                          {t('urlDetails.brokenLinks.link')}
                        </span>
                      </td>
                      <td className='px-6 py-4 whitespace-nowrap text-sm font-medium'>
                        <a
                          href={link.link_url}
                          target='_blank'
                          rel='noopener noreferrer'
                          className='text-blue-600 hover:text-blue-900'
                        >
                          {t('urlDetails.brokenLinks.test')}
                        </a>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className='sm:hidden'>
              {brokenLinks.map((link, index) => (
                <div key={index} className='border-b border-gray-200 last:border-b-0 px-4 py-4'>
                  <div className='space-y-2'>
                    <div className='flex items-center justify-between'>
                      <span
                        className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                          link.status_code >= 400 && link.status_code < 500
                            ? 'bg-yellow-100 text-yellow-800'
                            : 'bg-red-100 text-red-800'
                        }`}
                      >
                        {link.status_code}
                      </span>
                      <span className='inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800'>
                        {t('urlDetails.brokenLinks.link')}
                      </span>
                    </div>
                    <div className='text-sm text-gray-900 break-all'>
                      {link.link_url}
                    </div>
                    <div className='text-right'>
                      <a
                        href={link.link_url}
                        target='_blank'
                        rel='noopener noreferrer'
                        className='text-blue-600 hover:text-blue-900 text-sm font-medium'
                      >
                        {t('urlDetails.brokenLinks.test')}
                      </a>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {brokenLinks.length === 0 && url.status === 'completed' && (
          <div className='bg-white rounded-lg shadow p-12 text-center'>
            <NoResultsIcon className='mx-auto h-12 w-12 text-green-400' />
            <h3 className='mt-2 text-lg font-medium text-gray-900'>
              {t('urlDetails.noBrokenLinks.title')}
            </h3>
            <p className='mt-1 text-sm text-gray-500'>
              {t('urlDetails.noBrokenLinks.message')}
            </p>
          </div>
        )}
      </div>
    </div>
  );
};

export default URLDetailsPage;
