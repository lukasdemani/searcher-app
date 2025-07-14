import React from 'react';

interface PlayIconProps {
  className?: string;
}

const PlayIcon: React.FC<PlayIconProps> = ({ className = "h-4 w-4" }) => {
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
        d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1.586a1 1 0 01.707.293l2.414 2.414a1 1 0 00.707.293H15M13 16h1.586a1 1 0 01.707.293l2.414 2.414a1 1 0 00.707.293H19M6 20l2.5-2.5M18 4l-2.5 2.5M8 8l2.5-2.5M16 16l2.5 2.5" 
      />
    </svg>
  );
};

export default PlayIcon; 