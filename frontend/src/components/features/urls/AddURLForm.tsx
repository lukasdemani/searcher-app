import React, { useState } from 'react';
import { PlusIcon, LinkIcon } from '@heroicons/react/24/outline';
import Button from '../../ui/Button';
import Input from '../../ui/Input';
import Modal from '../../ui/Modal';

interface AddURLFormProps {
  onAddURL: (url: string) => Promise<void>;
}

const AddURLForm: React.FC<AddURLFormProps> = ({ onAddURL }) => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [url, setUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const normalizeURL = (input: string): string => {
    let normalizedUrl = input.trim();
    
    if (!normalizedUrl.startsWith('http://') && !normalizedUrl.startsWith('https://')) {
      normalizedUrl = 'https://' + normalizedUrl;
    }
    
    return normalizedUrl;
  };

  const validateURL = (input: string): boolean => {
    try {
      const normalizedUrl = normalizeURL(input);
      const urlObj = new URL(normalizedUrl);
      return (urlObj.protocol === 'http:' || urlObj.protocol === 'https:') && 
             urlObj.hostname.length > 0;
    } catch {
      return false;
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!url.trim()) {
      setError('Insira uma URL');
      return;
    }

    if (!validateURL(url)) {
      setError('Insira uma URL válida (ex: facebook.com, www.google.com)');
      return;
    }

    setLoading(true);
    try {
      const normalizedUrl = normalizeURL(url);
      await onAddURL(normalizedUrl);
      setUrl('');
      setIsModalOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao adicionar URL');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (!loading) {
      setIsModalOpen(false);
      setUrl('');
      setError('');
    }
  };

  const handleURLChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setUrl(e.target.value);
    if (error) setError(''); // Clear error when user starts typing
  };

  return (
    <>
      <Button onClick={() => setIsModalOpen(true)} className="shrink-0">
        <PlusIcon className="h-4 w-4 mr-2" />
        Adicionar URL
      </Button>

      <Modal
        isOpen={isModalOpen}
        onClose={handleClose}
        title="Adicionar Nova URL para Análise"
        size="md"
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="URL do Website"
            type="url"
            placeholder="facebook.com ou www.google.com"
            value={url}
            onChange={handleURLChange}
            error={error}
            helperText="Digite a URL (o https:// será adicionado automaticamente)"
            leftIcon={<LinkIcon />}
            disabled={loading}
            autoFocus
          />

          <div className="bg-blue-50 border border-blue-200 rounded-md p-4">
            <h4 className="text-sm font-medium text-blue-800 mb-2">
              O que será analisado:
            </h4>
            <ul className="text-sm text-blue-700 space-y-1">
              <li>• Detecção da versão HTML</li>
              <li>• Extração do título da página</li>
              <li>• Contagem de tags de cabeçalho (H1-H6)</li>
              <li>• Links internos vs externos</li>
              <li>• Detecção de links quebrados</li>
              <li>• Presença de formulário de login</li>
            </ul>
          </div>

          <div className="flex justify-end space-x-3 pt-4">
            <Button
              type="button"
              variant="secondary"
              onClick={handleClose}
              disabled={loading}
            >
              Cancelar
            </Button>
            <Button type="submit" loading={loading}>
              Adicionar & Analisar
            </Button>
          </div>
        </form>
      </Modal>
    </>
  );
};

export default AddURLForm; 