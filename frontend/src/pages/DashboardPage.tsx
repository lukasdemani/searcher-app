import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import Button from '../components/ui/Button';
import Input from '../components/ui/Input';
import StatusBadge from '../components/ui/StatusBadge';
import AddURLForm from '../components/features/urls/AddURLForm';
import { useURLs } from '../hooks/useURLs';

export const DashboardPage: React.FC = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  
  const { 
    urls, 
    loading, 
    selectedURLs,
    addURL, 
    deleteURL, 
    bulkDelete,
    bulkAnalyze,
    toggleSelect,
    selectAll,
    clearSelection,
    exportData,
    importData,
    refetch 
  } = useURLs({
    search: searchTerm,
    status: statusFilter,
    autoRefresh: true,
    refreshInterval: 5000
  });

  const filteredUrls = urls.filter(url => {
    const matchesSearch = url.url.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         url.title?.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = statusFilter === 'all' || url.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const stats = {
    total: urls.length,
    processing: urls.filter(url => url.status === 'processing').length,
    completed: urls.filter(url => url.status === 'completed').length,
    failed: urls.filter(url => url.status === 'error').length,
  };

  const handleSelectAll = () => {
    if (selectedURLs.length === filteredUrls.length) {
      clearSelection();
    } else {
      filteredUrls.forEach(url => {
        if (!selectedURLs.includes(url.id)) {
          toggleSelect(url.id);
        }
      });
    }
  };

  const handleBulkDelete = async () => {
    await bulkDelete(selectedURLs);
  };

  const handleBulkAnalyze = async () => {
    await bulkAnalyze(selectedURLs);
  };

  const isAllSelected = filteredUrls.length > 0 && selectedURLs.length === filteredUrls.length;
  const isSomeSelected = selectedURLs.length > 0 && selectedURLs.length < filteredUrls.length;

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">Website Analyzer</h1>

        <div className="bg-white shadow rounded-lg mb-6">
          <div className="px-6 py-4 border-b border-gray-200">
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-3 sm:space-y-0">
              <div className="flex-1 max-w-lg">
                <Input
                  placeholder="Buscar por URL ou título..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full"
                />
              </div>
              <div className="flex space-x-3">
                <select
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                  className="border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="all">Todos os status</option>
                  <option value="queued">Pendente</option>
                  <option value="processing">Processando</option>
                  <option value="completed">Concluído</option>
                  <option value="error">Com erro</option>
                </select>
              </div>
            </div>
          </div>
        </div>
        <div className="bg-white shadow rounded-lg overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg leading-6 font-medium text-gray-900">
              URLs ({filteredUrls.length})
            </h3>
          </div>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    <input
                      type="checkbox"
                      checked={isAllSelected}
                      ref={(input) => {
                        if (input) input.indeterminate = isSomeSelected;
                      }}
                      onChange={handleSelectAll}
                      className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                    />
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    URL / Título
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Links
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Análise
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Ações
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {filteredUrls.map((url) => (
                  <tr 
                    key={url.id} 
                    className={`hover:bg-gray-50 ${selectedURLs.includes(url.id) ? 'bg-blue-50' : ''}`}
                  >
                    <td className="px-6 py-4 whitespace-nowrap">
                      <input
                        type="checkbox"
                        checked={selectedURLs.includes(url.id)}
                        onChange={() => toggleSelect(url.id)}
                        className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                      />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div>
                        <div className="text-sm font-medium text-gray-900 truncate max-w-xs">
                          {url.title || url.url}
                        </div>
                        <div className="text-sm text-gray-500 truncate max-w-xs">
                          {url.url}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <StatusBadge status={url.status} />
                      {url.error_message && (
                        <div className="text-xs text-red-600 mt-1 truncate max-w-xs">
                          {url.error_message}
                        </div>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <div className="space-y-1">
                        <div>Internos: {url.internal_links_count || 0}</div>
                        <div>Externos: {url.external_links_count || 0}</div>
                        {url.broken_links_count && url.broken_links_count > 0 && (
                          <div className="text-red-600">Quebrados: {url.broken_links_count}</div>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {url.updated_at ? (
                        new Date(url.updated_at).toLocaleDateString('pt-BR')
                      ) : (
                        'Nunca'
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
                      <Link
                        to={`/url/${url.id}`}
                        className="text-blue-600 hover:text-blue-900"
                      >
                        Ver detalhes
                      </Link>
                      <button
                        onClick={() => deleteURL(url.id)}
                        className="text-red-600 hover:text-red-900 ml-2"
                      >
                        Excluir
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {filteredUrls.length === 0 && (
            <div className="text-center py-12">
              <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.172 16.172a4 4 0 015.656 0M9 12h6m-6-4h6m2 5.291A7.962 7.962 0 0118 12a8 8 0 01-8 8 8 8 0 01-8-8 8 8 0 018-8c2.027 0 3.9.753 5.334 2.009"/>
              </svg>
              <h3 className="mt-2 text-sm font-medium text-gray-900">Nenhuma URL encontrada</h3>
              <p className="mt-1 text-sm text-gray-500">
                {searchTerm || statusFilter !== 'all' 
                  ? 'Tente ajustar os filtros ou adicionar uma nova URL.'
                  : 'Comece adicionando uma URL para análise.'
                }
              </p>
              {!searchTerm && statusFilter === 'all' && (
                <div className="mt-6">
                  <AddURLForm onAddURL={addURL} />
                </div>
              )}
            </div>
          )}
        </div>
      </div>


    </div>
  );
}; 