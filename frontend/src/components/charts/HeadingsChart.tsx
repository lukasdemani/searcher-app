import React from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

interface HeadingsChartProps {
  h1Count: number;
  h2Count: number;
  h3Count: number;
  h4Count: number;
  h5Count: number;
  h6Count: number;
}

const HeadingsChart: React.FC<HeadingsChartProps> = ({ 
  h1Count, 
  h2Count, 
  h3Count, 
  h4Count, 
  h5Count, 
  h6Count 
}) => {

  const data = [
    { level: 'H1', count: h1Count },
    { level: 'H2', count: h2Count },
    { level: 'H3', count: h3Count },
    { level: 'H4', count: h4Count },
    { level: 'H5', count: h5Count },
    { level: 'H6', count: h6Count },
  ];

  return (
    <div className="w-full h-full">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart
          data={data}
          margin={{
            top: 20,
            right: 30,
            left: 20,
            bottom: 5,
          }}
        >
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis 
            dataKey="level" 
            tick={{ fontSize: 12 }}
          />
          <YAxis tick={{ fontSize: 12 }} />
          <Tooltip />
          <Bar dataKey="count" fill="#3B82F6" />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
};

export default HeadingsChart;
