import React from 'react';
import { SortAscIcon, SortDescIcon, SortIcon as SortIconBase } from '../icons';

interface SortIconProps {
  field: string;
  currentSortField: string;
  sortDirection: 'asc' | 'desc';
  className?: string;
}

const SortIcon: React.FC<SortIconProps> = ({
  field,
  currentSortField,
  sortDirection,
  className = 'w-4 h-4',
}) => {
  if (currentSortField !== field) {
    return <SortIconBase className={`${className} text-gray-400`} />;
  }

  return sortDirection === 'asc' ? (
    <SortAscIcon className={`${className} text-blue-600`} />
  ) : (
    <SortDescIcon className={`${className} text-blue-600`} />
  );
};

export default SortIcon;
