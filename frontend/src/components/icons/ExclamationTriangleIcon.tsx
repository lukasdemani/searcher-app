import React from 'react';

interface ExclamationTriangleIconProps {
  className?: string;
}

const ExclamationTriangleIcon: React.FC<ExclamationTriangleIconProps> = ({ className = "h-6 w-6" }) => {
  return (
    <svg 
      className={className} 
      fill="none" 
      stroke="currentColor" 
      viewBox="0 0 24 24"
    >
      <path 
        strokeLinecap="round" 
        strokeLinejoin="round" 
        strokeWidth={2} 
        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.732 16.5c-.77.833-.192 2.5 1.732 2.5z"
      />
    </svg>
  );
};

export default ExclamationTriangleIcon; 