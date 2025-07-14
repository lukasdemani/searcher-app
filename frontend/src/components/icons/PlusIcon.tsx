import React from 'react';

interface PlusIconProps {
  className?: string;
}

const PlusIcon: React.FC<PlusIconProps> = ({ className = "h-4 w-4" }) => {
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
        d="M12 6v6m0 0v6m0-6h6m-6 0H6" 
      />
    </svg>
  );
};

export default PlusIcon; 