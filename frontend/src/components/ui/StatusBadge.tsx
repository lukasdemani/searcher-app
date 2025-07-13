import React from 'react';
import { URLStatus } from '../../types';
import {
  ClockIcon,
  ArrowPathIcon,
  CheckCircleIcon,
  XCircleIcon,
} from '@heroicons/react/24/outline';

interface StatusBadgeProps {
  status: URLStatus;
  showIcon?: boolean;
  size?: 'sm' | 'md';
}

const StatusBadge: React.FC<StatusBadgeProps> = ({
  status,
  showIcon = true,
  size = 'md',
}) => {
  const statusConfig = {
    queued: {
      color: 'bg-gray-100 text-gray-800',
      icon: ClockIcon,
      text: 'Queued',
    },
    processing: {
      color: 'bg-blue-100 text-blue-800',
      icon: ArrowPathIcon,
      text: 'Processing',
    },
    completed: {
      color: 'bg-green-100 text-green-800',
      icon: CheckCircleIcon,
      text: 'Completed',
    },
    error: {
      color: 'bg-red-100 text-red-800',
      icon: XCircleIcon,
      text: 'Error',
    },
  };

  const config = statusConfig[status];
  const Icon = config.icon;

  const sizeClasses = {
    sm: 'px-2 py-1 text-xs',
    md: 'px-2.5 py-0.5 text-xs',
  };

  const iconSizeClasses = {
    sm: 'w-3 h-3',
    md: 'w-4 h-4',
  };

  return (
    <span
      className={`inline-flex items-center rounded-full font-medium ${config.color} ${sizeClasses[size]}`}
    >
      {showIcon && (
        <Icon
          className={`${iconSizeClasses[size]} mr-1 ${
            status === 'processing' ? 'animate-spin' : ''
          }`}
        />
      )}
      {config.text}
    </span>
  );
};

export default StatusBadge; 