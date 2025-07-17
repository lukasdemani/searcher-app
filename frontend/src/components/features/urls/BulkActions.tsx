import React, { useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import toast from 'react-hot-toast';
import { 
  DownloadIcon, 
  PlayIcon, 
  TrashIcon, 
  UploadIcon, 
  XCircleIcon 
} from '../../icons';
import Button from '../../ui/Button';
import Modal from '../../ui/Modal';
import APIService from '../../../services/api';
import type { URLAnalysis } from '../../../types/types';

interface BulkActionsProps {
  selectedURLs: number[];
  onClearSelection: () => void;
  onRefresh: () => void;
  onBulkAnalyze: (ids: number[]) => Promise<void>;
  onBulkDelete: (ids: number[]) => Promise<void>;
}

const BulkActions: React.FC<BulkActionsProps> = ({
  selectedURLs,
  onClearSelection,
  onRefresh,
  onBulkAnalyze,
  onBulkDelete,
}) => {
  const { t } = useTranslation();
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [isImportModalOpen, setIsImportModalOpen] = useState(false);
  const [isExporting, setIsExporting] = useState(false);
  const [isImporting, setIsImporting] = useState(false);
  const [importFile, setImportFile] = useState<File | null>(null);
  const [bulkAnalyzing, setBulkAnalyzing] = useState(false);
  const [bulkDeleting, setBulkDeleting] = useState(false);

  const handleBulkAnalyze = useCallback(async () => {
    try {
      setBulkAnalyzing(true);
      await onBulkAnalyze(selectedURLs);
      toast.success(t('messages.bulkAnalyzeStarted'));
    } catch (error) {
      toast.error(t('messages.errorBulkAnalyze'));
    } finally {
      setBulkAnalyzing(false);
    }
  }, [selectedURLs, onBulkAnalyze, t]);

  const handleBulkDelete = useCallback(async () => {
    try {
      setBulkDeleting(true);
      await onBulkDelete(selectedURLs);
      toast.success(t('messages.bulkDeleteSuccess'));
      setIsDeleteModalOpen(false);
      onClearSelection();
    } catch (error) {
      toast.error(t('messages.errorBulkDelete'));
    } finally {
      setBulkDeleting(false);
    }
  }, [selectedURLs, onBulkDelete, onClearSelection, t]);

  const handleExportCSV = useCallback(async () => {
    try {
      setIsExporting(true);
      
      const response = await APIService.getURLs({ limit: 1000 });
      const urls = response.data;
      
      const csvHeaders = [
        'ID', 'URL', 'Title', 'HTML Version', 'Status', 
        'H1 Count', 'H2 Count', 'H3 Count', 'H4 Count', 'H5 Count', 'H6 Count',
        'Internal Links', 'External Links', 'Broken Links', 'Has Login Form',
        'Created At', 'Updated At'
      ];
      
      const csvRows = urls.map((url: URLAnalysis) => [
        url.id,
        url.url,
        url.title || '',
        url.html_version || '',
        url.status,
        url.h1_count || 0,
        url.h2_count || 0,
        url.h3_count || 0,
        url.h4_count || 0,
        url.h5_count || 0,
        url.h6_count || 0,
        url.internal_links_count || 0,
        url.external_links_count || 0,
        url.broken_links_count || 0,
        url.has_login_form ? 'Yes' : 'No',
        url.created_at,
        url.updated_at
      ]);
      
      const csvContent = [csvHeaders, ...csvRows]
        .map(row => row.map((cell: string | number | boolean) => `"${cell}"`).join(','))
        .join('\n');
      
      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
      const link = document.createElement('a');
      const url = URL.createObjectURL(blob);
      link.setAttribute('href', url);
      link.setAttribute('download', `website-analyzer-export-${new Date().toISOString().split('T')[0]}.csv`);
      link.style.visibility = 'hidden';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      toast.success(t('messages.exportSuccess'));
    } catch (error) {
      toast.error(t('messages.errorExporting'));
    } finally {
      setIsExporting(false);
    }
  }, [t]);

  const handleImportCSV = useCallback(async () => {
    if (!importFile) return;
    
    try {
      setIsImporting(true);
      
      const text = await importFile.text();
      const lines = text.split('\n');
      const headers = lines[0].split(',').map(h => h.replace(/"/g, ''));
      
      const urlColumnIndex = headers.findIndex(h => 
        h.toLowerCase().includes('url') || h.toLowerCase().includes('domain')
      );
      
      if (urlColumnIndex === -1) {
        toast.error(t('messages.importError'));
        return;
      }
      
      const urls = lines.slice(1)
        .filter(line => line.trim())
        .map(line => {
          const columns = line.split(',');
          return columns[urlColumnIndex]?.replace(/"/g, '').trim();
        })
        .filter(url => url && url.length > 0);
      
      if (urls.length === 0) {
        toast.error(t('messages.noUrlsFound'));
        return;
      }
      
      let imported = 0;
      for (const url of urls) {
        try {
          await APIService.addURL({ url });
          imported++;
        } catch (error) {
          console.error(`Failed to import URL: ${url}`, error);
        }
      }
      
      toast.success(t('messages.importSuccess', { count: imported }));
      setIsImportModalOpen(false);
      setImportFile(null);
      onRefresh();
    } catch (error) {
      toast.error(t('messages.errorImporting'));
    } finally {
      setIsImporting(false);
    }
  }, [importFile, t, onRefresh]);

  const handleFileChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file && file.type === 'text/csv') {
      setImportFile(file);
    } else {
      toast.error(t('messages.invalidFileType'));
    }
  }, [t]);

  const selectedCount = selectedURLs.length;
  const isSelected = selectedCount > 0;

  return (
    <>
      <div className="bg-white border-b border-gray-200 px-4 sm:px-6 py-4">
        <div className="flex flex-col space-y-4 sm:flex-row sm:items-center sm:justify-between sm:space-y-0">
          <div className="flex items-center space-x-4">
            {isSelected ? (
              <div className="flex items-center space-x-4">
                <span className="text-sm font-medium text-gray-700">
                  {selectedCount === 1 
                    ? t('bulkActions.selectedUrls', { count: selectedCount })
                    : t('bulkActions.selectedUrlsPlural', { count: selectedCount })
                  }
                </span>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={onClearSelection}
                  className="text-xs"
                >
                  <XCircleIcon className="h-4 w-4 mr-1" />
                  {t('bulkActions.clearSelection')}
                </Button>
              </div>
            ) : (
              <span className="text-sm text-gray-500">
                {t('bulkActions.selectUrls')}
              </span>
            )}
          </div>
          
          <div className="flex flex-col space-y-2 sm:flex-row sm:space-y-0 sm:space-x-2">
            {isSelected && (
              <div className="flex space-x-2">
                <Button
                  variant="primary"
                  size="sm"
                  onClick={handleBulkAnalyze}
                  loading={bulkAnalyzing}
                  disabled={bulkAnalyzing}
                  className="w-full sm:w-auto"
                >
                  <PlayIcon className="h-4 w-4 mr-1" />
                  {t('bulkActions.reanalyze')}
                </Button>
                <Button
                  variant="danger"
                  size="sm"
                  onClick={() => setIsDeleteModalOpen(true)}
                  disabled={bulkDeleting}
                  className="w-full sm:w-auto"
                >
                  <TrashIcon className="h-4 w-4 mr-1" />
                  {t('common.delete')}
                </Button>
              </div>
            )}
            
            <div className="flex space-x-2">
              <Button
                variant="secondary"
                size="sm"
                onClick={handleExportCSV}
                loading={isExporting}
                disabled={isExporting}
                className="w-full sm:w-auto"
              >
                <DownloadIcon className="h-4 w-4 mr-1" />
                {t('bulkActions.exportCsv')}
              </Button>
              
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setIsImportModalOpen(true)}
                className="w-full sm:w-auto"
              >
                <UploadIcon className="h-4 w-4 mr-1" />
                {t('bulkActions.importUrls')}
              </Button>
            </div>
          </div>
        </div>
      </div>

      <Modal
        isOpen={isDeleteModalOpen}
        onClose={() => setIsDeleteModalOpen(false)}
        title={t('bulkActions.confirmDelete')}
      >
        <div className="p-6">
          <p className="text-sm text-gray-600 mb-6">
            {selectedCount === 1 
              ? t('bulkActions.deleteMessage', { count: selectedCount })
              : t('bulkActions.deleteMessagePlural', { count: selectedCount })
            }
          </p>
          <div className="flex justify-end space-x-3">
            <Button
              variant="secondary"
              onClick={() => setIsDeleteModalOpen(false)}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="danger"
              onClick={handleBulkDelete}
              loading={bulkDeleting}
            >
              {t('common.delete')}
            </Button>
          </div>
        </div>
      </Modal>

      <Modal
        isOpen={isImportModalOpen}
        onClose={() => setIsImportModalOpen(false)}
        title={t('bulkActions.import.title')}
      >
        <div className="p-6">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                {t('bulkActions.import.selectFile')}
              </label>
              <input
                type="file"
                accept=".csv"
                onChange={handleFileChange}
                className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
              />
              <p className="text-xs text-gray-500 mt-1">
                {t('bulkActions.import.fileHelper')}
              </p>
            </div>
            
            <div className="bg-gray-50 p-3 rounded-md">
              <p className="text-sm font-medium text-gray-700 mb-2">
                {t('bulkActions.import.expectedFormat')}
              </p>
              <pre className="text-xs text-gray-600">
                url,title{'\n'}
                https://example.com,Example Site{'\n'}
                https://google.com,Google
              </pre>
            </div>
          </div>
          
          <div className="flex justify-end space-x-3 mt-6">
            <Button
              variant="secondary"
              onClick={() => setIsImportModalOpen(false)}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="primary"
              onClick={handleImportCSV}
              loading={isImporting}
              disabled={!importFile || isImporting}
            >
              {t('bulkActions.import.import')}
            </Button>
          </div>
        </div>
      </Modal>
    </>
  );
};

export default BulkActions;
