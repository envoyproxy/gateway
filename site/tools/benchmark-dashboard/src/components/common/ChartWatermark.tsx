import React from 'react';
import { GitBranch } from 'lucide-react';

interface ChartWatermarkProps {
  text?: string;
  position?: 'bottom-center' | 'bottom-left' | 'bottom-right' | 'top-center' | 'top-left' | 'top-right';
  className?: string;
}

const ChartWatermark: React.FC<ChartWatermarkProps> = ({
  text = "Kind Cluster / GitHub CI",
  position = "top-center",
  className = ""
}) => {
  const getPositionClasses = () => {
    switch (position) {
      case 'bottom-left':
        return 'bottom-8 left-4';
      case 'bottom-right':
        return 'bottom-8 right-4';
      case 'bottom-center':
        return 'bottom-8 left-1/2 transform -translate-x-1/2';
      case 'top-left':
        return 'top-4 left-4';
      case 'top-right':
        return 'top-4 right-4';
      case 'top-center':
        return 'left-1/2 transform -translate-x-1/2 -translate-y-1/2';
      default:
        return 'bottom-8 left-1/2 transform -translate-x-1/2';
    }
  };

  return (
    <div
      className={`absolute ${getPositionClasses()} z-10 flex items-center gap-1 px-2 py-1 bg-indigo-600/40 rounded-md border border-gray-200/30 shadow-sm ${className}`}
    >
      <GitBranch className="h-3 w-3 text-white" />
      <span className="text-xs text-white font-medium">{text}</span>
    </div>
  );
};

export default ChartWatermark;
