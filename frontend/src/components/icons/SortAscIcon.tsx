import React from 'react';

interface SortAscIconProps {
  className?: string;
}

const SortAscIcon: React.FC<SortAscIconProps> = ({ className = "w-4 h-4" }) => {
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
        d="M3 4h13M3 8h9m-9 4h6m4 0l4-4m0 0l4 4m-4-4v12" 
      />
    </svg>
  );
};

export default SortAscIcon; 