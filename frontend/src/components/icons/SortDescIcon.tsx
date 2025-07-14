import React from 'react';

interface SortDescIconProps {
  className?: string;
}

const SortDescIcon: React.FC<SortDescIconProps> = ({ className = "w-4 h-4" }) => {
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
        d="M3 4h13M3 8h9m-9 4h9m5-4v12m0 0l-4-4m4 4l4-4" 
      />
    </svg>
  );
};

export default SortDescIcon; 