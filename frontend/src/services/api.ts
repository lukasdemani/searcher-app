import axios from 'axios';
import toast from 'react-hot-toast';
import i18n from '../i18n';

const API_BASE_URL = 'http://localhost:8080/api';
console.log(API_BASE_URL);

export interface BrokenLink {
  id: number;
  url_id: number;
  link_url: string;
  status_code: number;
  error_message?: string;
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

export interface URLRequest {
  url: string;
}

export interface SuccessResponse {
  message: string;
  data?: any;
}

export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const message = error.response?.data?.error || i18n.t('common.error');
    toast.error(message);

    return Promise.reject(error);
  }
);

// API Functions
export const getURLs = async (params?: {
  page?: number;
  limit?: number;
  search?: string;
  status?: string;
}) => {
  console.log('urls');
  try {
    const response = await api.get('/urls', { params });
    console.log(response);
    return response.data;
  } catch (error) {}
};

export const getURL = async (id: number) => {
  console.log('url');
  try {
    const response = await api.get(`/urls/${id}`);
    return response.data;
  } catch (error) {}
};

export const addURL = async (urlData: URLRequest): Promise<URLAnalysis> => {
  const response = await api.post('/urls', urlData);

  toast.success(i18n.t('messages.urlAdded'));
  return response.data.data;
};

export const analyzeURL = async (id: number): Promise<void> => {
  await api.put(`/urls/${id}/analyze`);
  toast.success(i18n.t('messages.analysisStarted'));
};

export const deleteURL = async (id: number): Promise<void> => {
  await api.delete(`/urls/${id}`);
  toast.success(i18n.t('messages.urlDeleted'));
};

export const bulkAnalyze = async (ids: number[]): Promise<SuccessResponse> => {
  try {
    const response = await api.post('/urls/bulk-analyze', { ids });
    return response.data;
  } catch (error) {
    throw new Error(i18n.t('messages.errorBulkAnalyze'));
  }
};

export const bulkDelete = async (ids: number[]): Promise<SuccessResponse> => {
  try {
    const response = await api.post('/urls/bulk-delete', { ids });
    return response.data;
  } catch (error) {
    throw new Error(i18n.t('messages.errorBulkDelete'));
  }
};

export const getBrokenLinks = async (urlId: number): Promise<BrokenLink[]> => {
  return [];
};

export const healthCheck = async (): Promise<{ status: string }> => {
  const response = await axios.get(
    `${API_BASE_URL.replace('/api', '')}/health`
  );
  return response.data;
};

// Default export for backward compatibility
const APIService = {
  getURLs,
  getURL,
  addURL,
  analyzeURL,
  deleteURL,
  bulkAnalyze,
  bulkDelete,
  getBrokenLinks,
  healthCheck,
};

export default APIService;
