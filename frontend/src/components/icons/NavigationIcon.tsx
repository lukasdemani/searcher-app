import React from 'react';

interface NavigationIconProps {
  className?: string;
}

const NavigationIcon: React.FC<NavigationIconProps> = ({ className = "h-5 w-5" }) => {
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

export default NavigationIcon; 