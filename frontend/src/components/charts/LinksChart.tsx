import React from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import { useTranslation } from 'react-i18next';

interface LinksChartProps {
  internalLinks: number;
  externalLinks: number;
  chartType?: 'bar' | 'donut';
}

interface TooltipProps {
  active?: boolean;
  payload?: Array<{
    value: number;
    name: string;
  }>;
  label?: string;
}

interface PieTooltipProps {
  active?: boolean;
  payload?: Array<{
    value: number;
    name: string;
  }>;
}

interface LabelProps {
  cx?: number;
  cy?: number;
  midAngle?: number;
  innerRadius?: number;
  outerRadius?: number;
  percent?: number;
}

const LinksChart: React.FC<LinksChartProps> = ({ 
  internalLinks, 
  externalLinks, 
  chartType = 'bar' 
}) => {
  const { t } = useTranslation();

  const data = [
    {
      name: t('urlDetails.charts.internalLinks'),
      value: internalLinks,
      color: '#3B82F6'
    },
    {
      name: t('urlDetails.charts.externalLinks'),
      value: externalLinks,
      color: '#EF4444'
    }
  ];

  const barData = [
    {
      type: t('urlDetails.charts.internalLinks'),
      count: internalLinks,
      fill: '#3B82F6'
    },
    {
      type: t('urlDetails.charts.externalLinks'),
      count: externalLinks,
      fill: '#EF4444'
    }
  ];

  const COLORS = ['#3B82F6', '#EF4444'];

  const CustomTooltip = ({ active, payload, label }: TooltipProps) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-white p-3 border border-gray-300 rounded shadow">
          <p className="text-sm font-medium">{label}</p>
          <p className="text-sm">
            <span className="font-medium">{t('urlDetails.charts.count')}: </span>
            {payload[0].value}
          </p>
        </div>
      );
    }
    return null;
  };

  const CustomPieTooltip = ({ active, payload }: PieTooltipProps) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-white p-3 border border-gray-300 rounded shadow">
          <p className="text-sm font-medium">{payload[0].name}</p>
          <p className="text-sm">
            <span className="font-medium">{t('urlDetails.charts.count')}: </span>
            {payload[0].value}
          </p>
        </div>
      );
    }
    return null;
  };

  const renderCustomizedLabel = ({ cx, cy, midAngle, innerRadius, outerRadius, percent }: LabelProps) => {
    if (!cx || !cy || midAngle === undefined || !innerRadius || !outerRadius || !percent) return null;
    
    const RADIAN = Math.PI / 180;
    const radius = innerRadius + (outerRadius - innerRadius) * 0.5;
    const x = cx + radius * Math.cos(-midAngle * RADIAN);
    const y = cy + radius * Math.sin(-midAngle * RADIAN);

    return (
      <text 
        x={x} 
        y={y} 
        fill="white" 
        textAnchor={x > cx ? 'start' : 'end'} 
        dominantBaseline="central"
        className="text-sm font-medium"
      >
        {`${(percent * 100).toFixed(0)}%`}
      </text>
    );
  };

  if (chartType === 'donut') {
    return (
      <div className="w-full h-full">
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={data}
              cx="50%"
              cy="50%"
              labelLine={false}
              label={renderCustomizedLabel}
              outerRadius="80%"
              innerRadius="40%"
              fill="#8884d8"
              dataKey="value"
            >
              {data.map((_, index) => (
                <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
              ))}
            </Pie>
            <Tooltip content={<CustomPieTooltip />} />
            <Legend />
          </PieChart>
        </ResponsiveContainer>
      </div>
    );
  }

  return (
    <div className="w-full h-full">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart
          data={barData}
          margin={{
            top: 20,
            right: 30,
            left: 20,
            bottom: 5,
          }}
        >
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis 
            dataKey="type" 
            tick={{ fontSize: 12 }}
            interval={0}
            angle={-45}
            textAnchor="end"
            height={60}
          />
          <YAxis tick={{ fontSize: 12 }} />
          <Tooltip content={<CustomTooltip />} />
          <Legend />
          <Bar 
            dataKey="count" 
            fill="#3B82F6"
            name={t('urlDetails.charts.linkCount')}
          />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
};

export default LinksChart;
