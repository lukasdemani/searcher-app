import React, { useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import Input from './Input';
import Select from './Select';

interface ColumnFilterProps {
  column: string;
  value: string;
  onChange: (value: string) => void;
  type: 'text' | 'select' | 'number' | 'boolean';
  options?: { value: string; label: string }[];
  placeholder?: string;
  className?: string;
}

const ColumnFilter: React.FC<ColumnFilterProps> = ({
  column,
  value,
  onChange,
  type,
  options,
  placeholder,
  className = '',
}) => {
  const { t } = useTranslation();

  const handleClear = useCallback(() => {
    onChange('');
  }, [onChange]);

  const getPlaceholder = () => {
    if (placeholder) return placeholder;
    if (type === 'text') {
      return t('dashboard.filters.filterBy', { field: t(`dashboard.table.${column}`) });
    }
    return t('dashboard.filters.all');
  };

  const renderFilter = () => {
    switch (type) {
      case 'select':
        return (
          <Select
            value={value}
            onChange={(val) => onChange(String(val))}
            options={options || []}
            className={`w-full text-sm ${className}`}
            placeholder={getPlaceholder()}
          />
        );
      
      case 'boolean':
        return (
          <Select
            value={value}
            onChange={(val) => onChange(String(val))}
            options={[
              { value: '', label: t('dashboard.filters.all') },
              { value: 'true', label: t('common.yes') },
              { value: 'false', label: t('common.no') }
            ]}
            className={`w-full text-sm ${className}`}
            placeholder={getPlaceholder()}
          />
        );
      
      case 'number':
        return (
          <Input
            type="text"
            placeholder={getPlaceholder()}
            value={value}
            onChange={(e) => {
              const val = e.target.value;
              if (val === '' || /^[0-9>=<\s]*$/.test(val)) {
                onChange(val);
              }
            }}
            className={`w-full text-sm ${className}`}
          />
        );
      
      default:
        return (
          <Input
            type="text"
            placeholder={getPlaceholder()}
            value={value}
            onChange={(e) => onChange(e.target.value)}
            className={`w-full text-sm ${className}`}
          />
        );
    }
  };

  return (
    <div className="relative">
      {renderFilter()}
      {value && (
        <button
          onClick={handleClear}
          className="absolute right-2 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 text-sm"
          title={t('dashboard.filters.clearFilter')}
        >
          ×
        </button>
      )}
    </div>
  );
};

export default ColumnFilter;
