import React from 'react';

interface NoResultsIconProps {
  className?: string;
}

const NoResultsIcon: React.FC<NoResultsIconProps> = ({ className = "mx-auto h-12 w-12" }) => {
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
        d="M9 12h6m-6-4h6m2 5.291A7.962 7.962 0 0118 12a8 8 0 01-8 8 8 8 0 01-8-8 8 8 0 018-8c2.027 0 3.9.753 5.334 2.009"
      />
    </svg>
  );
};

export default NoResultsIcon; 