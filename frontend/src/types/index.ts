import React from 'react';

export type URLStatus = 'queued' | 'processing' | 'completed' | 'error';

export interface HeadingCounts {
  h1: number;
  h2: number;
  h3: number;
  h4: number;
  h5: number;
  h6: number;
}

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

export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface ErrorResponse {
  error: string;
}

export interface SuccessResponse {
  message: string;
  data?: any;
}

export interface ChartData {
  name: string;
  value: number;
  fill?: string;
}

export interface HeadingChartData {
  name: string;
  count: number;
}

export interface ButtonProps {
  variant?: 'primary' | 'secondary' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
  disabled?: boolean;
  children: React.ReactNode;
  onClick?: () => void;
  type?: 'button' | 'submit' | 'reset';
  className?: string;
}

export interface InputProps {
  label?: string;
  error?: string;
  placeholder?: string;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  type?: string;
  required?: boolean;
  className?: string;
}

export interface TableColumn {
  key: string;
  label: string;
  sortable?: boolean;
  render?: (value: any, row: any) => React.ReactNode;
}

export interface PaginationProps {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

export interface UseURLsOptions {
  page?: number;
  limit?: number;
  search?: string;
  sortField?: keyof URLAnalysis;
  sortDirection?: 'asc' | 'desc';
}

export interface UseURLsReturn {
  urls: URLAnalysis[];
  loading: boolean;
  error: string | null;
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
  refetch: () => void;
  addURL: (url: string) => Promise<void>;
  deleteURL: (id: number) => Promise<void>;
  reanalyzeURL: (id: number) => Promise<void>;
  bulkAnalyze: (ids: number[]) => Promise<void>;
  bulkDelete: (ids: number[]) => Promise<void>;
} 