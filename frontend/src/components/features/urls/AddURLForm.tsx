import React, { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  CheckCircleIcon,
  ExclamationTriangleIcon,
  LinkIcon,
  PlusIcon,
  SpinnerIcon,
} from '../../icons';
import Button from '../../ui/Button';
import Input from '../../ui/Input';
import Modal from '../../ui/Modal';

interface AddURLFormProps {
  onAddURL: (url: string) => Promise<void>;
}

type ValidationState = 'idle' | 'validating' | 'valid' | 'invalid';

const AddURLForm: React.FC<AddURLFormProps> = ({ onAddURL }) => {
  const { t } = useTranslation();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [url, setUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [validationState, setValidationState] =
    useState<ValidationState>('idle');
  const [validationMessage, setValidationMessage] = useState('');

  const cleanupPastedURL = (input: string): string => {
    let cleaned = input.trim();

    if (cleaned.startsWith('https://')) {
      cleaned = cleaned.substring(8);
    } else if (cleaned.startsWith('http://')) {
      cleaned = cleaned.substring(7);
    }

    if (cleaned.startsWith('www.')) {
      cleaned = cleaned.substring(4);
    }

    const pathIndex = cleaned.indexOf('/');
    if (pathIndex > 0) {
      cleaned = cleaned.substring(0, pathIndex);
    }

    const queryIndex = cleaned.indexOf('?');
    if (queryIndex > 0) {
      cleaned = cleaned.substring(0, queryIndex);
    }

    const fragmentIndex = cleaned.indexOf('#');
    if (fragmentIndex > 0) {
      cleaned = cleaned.substring(0, fragmentIndex);
    }

    return cleaned;
  };

  const normalizeURL = (input: string): string => {
    let normalizedUrl = input.trim();

    if (
      !normalizedUrl.startsWith('http://') &&
      !normalizedUrl.startsWith('https://')
    ) {
      normalizedUrl = 'https://' + normalizedUrl;
    }

    return normalizedUrl;
  };

  const validateURL = useCallback((input: string): boolean => {
    if (!input.trim()) return false;

    try {
      const normalizedUrl = normalizeURL(input);
      const urlObj = new URL(normalizedUrl);

      const isValidProtocol =
        urlObj.protocol === 'http:' || urlObj.protocol === 'https:';
      const hasHostname = urlObj.hostname.length > 0;
      const hasValidHostname = urlObj.hostname.includes('.');

      return isValidProtocol && hasHostname && hasValidHostname;
    } catch {
      return false;
    }
  }, []);

  const performValidation = useCallback(
    (input: string) => {
      if (!input.trim()) {
        setValidationState('idle');
        setValidationMessage('');
        return;
      }

      setValidationState('validating');

      setTimeout(() => {
        if (validateURL(input)) {
          setValidationState('valid');
          setValidationMessage(t('addUrl.validation.valid'));
        } else {
          setValidationState('invalid');
          setValidationMessage(t('addUrl.validation.invalid'));
        }
      }, 300);
    },
    [t, validateURL]
  );

  useEffect(() => {
    const timer = setTimeout(() => {
      performValidation(url);
    }, 500);

    return () => clearTimeout(timer);
  }, [url, performValidation]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!url.trim()) {
      setError(t('addUrl.enterUrl'));
      return;
    }

    if (!validateURL(url)) {
      setError(t('addUrl.invalidUrl'));
      return;
    }

    setLoading(true);
    try {
      const normalizedUrl = normalizeURL(url);
      await onAddURL(normalizedUrl);
      setUrl('');
      setIsModalOpen(false);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : t('messages.errorAddingUrl')
      );
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (!loading) {
      setIsModalOpen(false);
      setUrl('');
      setError('');
      setValidationState('idle');
      setValidationMessage('');
    }
  };

  const handleURLChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setUrl(value);
    if (error) setError('');
  };

  const handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault();
    const pastedText = e.clipboardData.getData('text');
    const cleaned = cleanupPastedURL(pastedText);
    setUrl(cleaned);
    if (error) setError('');
  };

  const getPreviewURL = () => {
    if (!url.trim()) return '';
    return normalizeURL(url);
  };

  const getValidationIcon = () => {
    switch (validationState) {
      case 'validating':
        return <SpinnerIcon className='animate-spin h-5 w-5 text-gray-400' />;
      case 'valid':
        return <CheckCircleIcon className='h-5 w-5 text-green-500' />;
      case 'invalid':
        return <ExclamationTriangleIcon className='h-5 w-5 text-red-500' />;
      default:
        return <LinkIcon className='h-5 w-5 text-gray-400' />;
    }
  };

  return (
    <>
      <Button onClick={() => setIsModalOpen(true)} className='shrink-0 w-full sm:w-auto'>
        <PlusIcon className='h-4 w-4 mr-2' />
        {t('dashboard.addUrl')}
      </Button>

      <Modal
        isOpen={isModalOpen}
        onClose={handleClose}
        title={t('addUrl.title')}
        size='md'
      >
        <form onSubmit={handleSubmit} className='space-y-4'>
          <div>
            <Input
              label={t('addUrl.urlLabel')}
              type='text'
              placeholder={t('addUrl.placeholder')}
              value={url}
              onChange={handleURLChange}
              onPaste={handlePaste}
              error={error}
              helperText={t('addUrl.helper')}
              leftIcon={getValidationIcon()}
              disabled={loading}
              autoFocus
            />

            {validationMessage && (
              <div
                className={`mt-2 text-sm transition-all duration-200 ${
                  validationState === 'valid'
                    ? 'text-green-600'
                    : validationState === 'invalid'
                    ? 'text-red-600'
                    : 'text-gray-600'
                }`}
              >
                {validationMessage}
              </div>
            )}

            {url.trim() && validationState === 'valid' && (
              <div className='mt-3 p-3 bg-green-50 rounded-lg border border-green-200 transition-all duration-300 ease-in-out'>
                <div className='flex items-center mb-2'>
                  <CheckCircleIcon className='h-4 w-4 text-green-500 mr-2' />
                  <div className='text-sm font-medium text-green-800'>
                    {t('addUrl.preview')}
                  </div>
                </div>
                <div className='text-sm font-mono text-green-700 break-all bg-white px-2 py-1 rounded border'>
                  {getPreviewURL()}
                </div>
              </div>
            )}
          </div>

          <div className='bg-blue-50 border border-blue-200 rounded-md p-4'>
            <h4 className='text-sm font-medium text-blue-800 mb-2'>
              {t('addUrl.analysisInfo.title')}
            </h4>
            <ul className='text-sm text-blue-700 space-y-1'>
              <li className='flex items-center'>
                <span className='w-2 h-2 bg-blue-400 rounded-full mr-2'></span>
                {t('addUrl.analysisInfo.htmlVersion')}
              </li>
              <li className='flex items-center'>
                <span className='w-2 h-2 bg-blue-400 rounded-full mr-2'></span>
                {t('addUrl.analysisInfo.pageTitle')}
              </li>
              <li className='flex items-center'>
                <span className='w-2 h-2 bg-blue-400 rounded-full mr-2'></span>
                {t('addUrl.analysisInfo.headings')}
              </li>
              <li className='flex items-center'>
                <span className='w-2 h-2 bg-blue-400 rounded-full mr-2'></span>
                {t('addUrl.analysisInfo.links')}
              </li>
              <li className='flex items-center'>
                <span className='w-2 h-2 bg-blue-400 rounded-full mr-2'></span>
                {t('addUrl.analysisInfo.brokenLinks')}
              </li>
              <li className='flex items-center'>
                <span className='w-2 h-2 bg-blue-400 rounded-full mr-2'></span>
                {t('addUrl.analysisInfo.loginForm')}
              </li>
            </ul>
          </div>

          <div className='flex justify-end space-x-3 pt-4'>
            <Button
              type='button'
              variant='secondary'
              onClick={handleClose}
              disabled={loading}
            >
              {t('common.cancel')}
            </Button>
            <Button
              type='submit'
              loading={loading}
              disabled={validationState !== 'valid'}
            >
              {t('addUrl.addAndAnalyze')}
            </Button>
          </div>
        </form>
      </Modal>
    </>
  );
};

export default AddURLForm;
