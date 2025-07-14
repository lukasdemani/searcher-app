import React from 'react';
import { ChevronDownIcon } from '../icons';

interface SelectOption {
  value: string | number;
  label: string;
}

interface SelectProps {
  value: string | number;
  onChange: (value: string | number) => void;
  options: SelectOption[];
  placeholder?: string;
  className?: string;
  disabled?: boolean;
}

const Select: React.FC<SelectProps> = ({
  value,
  onChange,
  options,
  placeholder,
  className = '',
  disabled = false,
}) => {
  return (
    <div className={`relative ${className}`}>
      <select
        value={value}
        onChange={(e) => {
          const newValue = e.target.value;
          // Try to convert to number if the original value was a number
          const convertedValue = typeof value === 'number' ? Number(newValue) : newValue;
          onChange(convertedValue);
        }}
        disabled={disabled}
        className="block w-full pl-3 pr-10 py-2 text-base border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 sm:text-sm rounded-md bg-white disabled:bg-gray-50 disabled:text-gray-500 appearance-none"
      >
        {placeholder && <option value="">{placeholder}</option>}
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      <div className="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
        <ChevronDownIcon className="h-5 w-5 text-gray-400" />
      </div>
    </div>
  );
};

export default Select; 