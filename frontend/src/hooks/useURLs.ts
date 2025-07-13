import { useEffect, useRef, useState } from 'react';
import { toast } from 'react-hot-toast';
import { APIService } from '../services/api';
import { URLAnalysis } from '../types';

interface UseURLsOptions {
  page?: number;
  limit?: number;
  search?: string;
  status?: string;
  autoRefresh?: boolean;
  refreshInterval?: number;
}

interface UseURLsReturn {
  urls: URLAnalysis[];
  loading: boolean;
  error: string | null;
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
  selectedURLs: number[];
  refetch: () => void;
  addURL: (url: string) => Promise<void>;
  deleteURL: (id: number) => Promise<void>;
  bulkDelete: (ids: number[]) => Promise<void>;
  bulkAnalyze: (ids: number[]) => Promise<void>;
  toggleSelect: (id: number) => void;
  selectAll: () => void;
  clearSelection: () => void;
  exportData: (format?: 'csv' | 'json') => Promise<void>;
  importData: (file: File) => Promise<void>;
}

export const useURLs = (options: UseURLsOptions = {}): UseURLsReturn => {
  const {
    page = 1,
    limit = 10,
    search = '',
    status = '',
    autoRefresh = true,
    refreshInterval = 5000, // 5 seconds
  } = options;

  const [urls, setUrls] = useState<URLAnalysis[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedURLs, setSelectedURLs] = useState<number[]>([]);
  const [pagination, setPagination] = useState({
    page: 1,
    limit: 10,
    total: 0,
    totalPages: 0,
  });

  const intervalRef = useRef<NodeJS.Timeout | null>(null);
  const isMountedRef = useRef(true);

  const fetchURLs = async () => {
    try {
      setLoading(true);
      setError(null);

      const params = {
        page,
        limit,
        ...(search && { search }),
        ...(status && status !== 'all' && { status }),
      };

      const response = await APIService.getURLs(params);

      if (isMountedRef.current) {
        setUrls(response.data);
        setPagination({
          page: response.page,
          limit: response.limit,
          total: response.total,
          totalPages: response.total_pages,
        });
      }
    } catch (err) {
      if (isMountedRef.current) {
        const errorMessage =
          err instanceof Error ? err.message : 'Erro ao carregar URLs';
        setError(errorMessage);
        toast.error(errorMessage);
      }
    } finally {
      if (isMountedRef.current) {
        setLoading(false);
      }
    }
  };

  const stopPolling = () => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
  };

  useEffect(() => {
    isMountedRef.current = true;
    fetchURLs();
  }, []);

  useEffect(() => {
    return () => {
      isMountedRef.current = false;
      stopPolling();
    };
  }, []);

  const refetch = async () => {
    await fetchURLs();
  };

  const addURL = async (url: string) => {
    try {
      const newURL = await APIService.addURL({ url });

      setUrls((prevUrls) => [newURL, ...prevUrls]);

      toast.success('URL adicionada com sucesso!');
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'Erro ao adicionar URL';
      toast.error(errorMessage);
      throw err;
    }
  };

  const deleteURL = async (id: number) => {
    try {
      await APIService.deleteURL(id);
      toast.success('URL removida com sucesso!');
      setUrls((prevUrls) => prevUrls.filter((url) => url.id !== id));
      setSelectedURLs((prev) => prev.filter((selectedId) => selectedId !== id));
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'Erro ao remover URL';
      toast.error(errorMessage);
      throw err;
    }
  };

  const toggleSelect = (id: number) => {
    setSelectedURLs((prev) =>
      prev.includes(id)
        ? prev.filter((selectedId) => selectedId !== id)
        : [...prev, id]
    );
  };

  const selectAll = () => {
    setSelectedURLs(urls.map((url) => url.id));
  };

  const clearSelection = () => {
    setSelectedURLs([]);
  };

  return {
    urls,
    loading,
    error,
    pagination,
    selectedURLs,
    refetch,
    addURL,
    deleteURL,
    toggleSelect,
    selectAll,
    clearSelection,
  };
};
