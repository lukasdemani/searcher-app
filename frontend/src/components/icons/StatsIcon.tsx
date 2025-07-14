import React from 'react';

interface StatsIconProps {
  className?: string;
}

const StatsIcon: React.FC<StatsIconProps> = ({ className = "w-5 h-5" }) => {
  return (
    <svg 
      className={className} 
      fill="currentColor" 
      viewBox="0 0 20 20"
    >
      <path 
        d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
      />
    </svg>
  );
};

export default StatsIcon; 