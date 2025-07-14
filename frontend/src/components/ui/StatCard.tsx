import React from 'react';

interface StatCardProps {
  icon: React.ComponentType<{ className?: string }>;
  bgColor: string;
  label: string;
  value: number | string;
  className?: string;
}

const StatCard: React.FC<StatCardProps> = React.memo(({ 
  icon: Icon, 
  bgColor, 
  label, 
  value, 
  className = '' 
}) => {
  return (
    <div className={`bg-white overflow-hidden shadow rounded-lg ${className}`}>
      <div className="p-5">
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <div className={`w-8 h-8 ${bgColor} rounded-md flex items-center justify-center`}>
              <Icon className="w-5 h-5 text-white" />
            </div>
          </div>
          <div className="ml-5 w-0 flex-1">
            <dl>
              <dt className="text-sm font-medium text-gray-500 truncate">{label}</dt>
              <dd className="text-lg font-medium text-gray-900">{value}</dd>
            </dl>
          </div>
        </div>
      </div>
    </div>
  );
});

StatCard.displayName = 'StatCard';

export default StatCard; 