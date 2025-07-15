import { useCallback, useEffect, useState } from 'react';
import { APIService } from '../services/api';
import { PaginatedResponse, URLAnalysis } from '../types';
import { useWebSocket } from './useWebSocket';

interface UseURLsOptions {
  page?: number;
  limit?: number;
  search?: string;
  status?: string;
  autoRefresh?: boolean;
  refreshInterval?: number;
}

export const useURLs = (options: UseURLsOptions = {}) => {
  const [urls, setUrls] = useState<URLAnalysis[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedURLs, setSelectedURLs] = useState<number[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [totalPages, setTotalPages] = useState(0);

  const {
    page = 1,
    limit = 10,
    search = '',
    status = 'all',
    autoRefresh = false,
    refreshInterval = 30000,
  } = options;

  const handleStatusUpdate = useCallback((update: any) => {
    console.log('Status update received:', update);

    setUrls((currentUrls) => {
      const updatedUrls = [...currentUrls];
      const urlIndex = updatedUrls.findIndex((url) => url.id === update.url_id);

      if (urlIndex !== -1) {
        if (update.status === 'deleted') {
          updatedUrls.splice(urlIndex, 1);
        } else if (update.data) {
          updatedUrls[urlIndex] = { ...updatedUrls[urlIndex], ...update.data };
        } else {
          updatedUrls[urlIndex] = {
            ...updatedUrls[urlIndex],
            status: update.status,
          };
        }
      }

      return updatedUrls;
    });
  }, []);

  const { isConnected } = useWebSocket({
    onStatusUpdate: handleStatusUpdate,
  });

  const fetchURLs = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const response: PaginatedResponse<URLAnalysis> = await APIService.getURLs(
        {
          page,
          limit,
          search,
          status: status === 'all' ? undefined : status,
        }
      );

      setUrls(response.data);
      setTotalCount(response.total);
      setTotalPages(response.total_pages);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch URLs');
      console.error('Error fetching URLs:', err);
    } finally {
      setLoading(false);
    }
  }, [page, limit, search, status]);

  const addURL = useCallback(async (url: string) => {
    try {
      const newURL = await APIService.addURL({ url });
      setUrls((prevUrls) => [newURL, ...(prevUrls || [])]);
      return newURL;
    } catch (err) {
      console.error('Error adding URL:', err);
      throw err;
    }
  }, []);

  const deleteURL = useCallback(async (id: number) => {
    try {
      await APIService.deleteURL(id);
      setUrls((prevUrls) => (prevUrls || []).filter((url) => url.id !== id));
      setSelectedURLs((prev) => prev.filter((selectedId) => selectedId !== id));
    } catch (err) {
      console.error('Error deleting URL:', err);
      throw err;
    }
  }, []);

  const bulkDelete = useCallback(async (ids: number[]) => {
    try {
      await APIService.bulkDelete(ids);
      setUrls((prevUrls) =>
        (prevUrls || []).filter((url) => !ids.includes(url.id))
      );
      setSelectedURLs([]);
    } catch (err) {
      console.error('Error bulk deleting URLs:', err);
      throw err;
    }
  }, []);

  const bulkAnalyze = useCallback(async (ids: number[]) => {
    try {
      await APIService.bulkAnalyze(ids);
    } catch (err) {
      console.error('Error bulk analyzing URLs:', err);
      throw err;
    }
  }, []);

  const analyzeURL = useCallback(async (id: number) => {
    try {
      await APIService.analyzeURL(id);
    } catch (err) {
      console.error('Error analyzing URL:', err);
      throw err;
    }
  }, []);

  const toggleSelect = useCallback((id: number) => {
    setSelectedURLs((prev) =>
      prev.includes(id)
        ? prev.filter((selectedId) => selectedId !== id)
        : [...prev, id]
    );
  }, []);

  const clearSelection = useCallback(() => {
    setSelectedURLs([]);
  }, []);

  const selectAll = useCallback(() => {
    setSelectedURLs(urls.map((url) => url.id));
  }, [urls]);

  useEffect(() => {
    fetchURLs();
  }, [fetchURLs]);

  useEffect(() => {
    if (autoRefresh && !isConnected) {
      const interval = setInterval(fetchURLs, refreshInterval);
      return () => clearInterval(interval);
    }
  }, [autoRefresh, refreshInterval, isConnected, fetchURLs]);

  return {
    urls,
    loading,
    error,
    selectedURLs,
    totalCount,
    totalPages,
    isConnected,
    addURL,
    deleteURL,
    bulkDelete,
    bulkAnalyze,
    analyzeURL,
    toggleSelect,
    clearSelection,
    selectAll,
    refetch: fetchURLs,
  };
};
