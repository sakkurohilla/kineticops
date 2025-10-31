import React, { useState } from 'react';
import { Clock, Calendar } from 'lucide-react';
import CustomTimeRangeSelector from './CustomTimeRangeSelector';

export type TimeRange = '1h' | '6h' | '24h' | '7d' | '30d' | 'custom';

interface TimeRangeSelectorProps {
  selected: TimeRange;
  onChange: (range: TimeRange, customStart?: string, customEnd?: string) => void;
}

const timeRanges: { value: TimeRange; label: string }[] = [
  { value: '1h', label: 'Last Hour' },
  { value: '6h', label: '6 Hours' },
  { value: '24h', label: '24 Hours' },
  { value: '7d', label: '7 Days' },
  { value: '30d', label: '30 Days' },
  { value: 'custom', label: 'Custom Range' },
];

const TimeRangeSelector: React.FC<TimeRangeSelectorProps> = ({ selected, onChange }) => {
  const [showCustomSelector, setShowCustomSelector] = useState(false);
  
  const handleRangeChange = (range: TimeRange) => {
    console.log('Time range changed to:', range);
    if (range === 'custom') {
      setShowCustomSelector(true);
    } else {
      onChange(range);
    }
  };
  
  const handleCustomApply = (startTime: string, endTime: string) => {
    onChange('custom', startTime, endTime);
  };

  return (
    <div className="flex items-center gap-2 bg-white rounded-lg shadow-sm p-1 border border-gray-200">
      <div className="flex items-center gap-2 px-3">
        <Clock className="w-4 h-4 text-gray-500" />
        <span className="text-sm font-medium text-gray-700">Time Range:</span>
      </div>
      {timeRanges.map((range) => (
        <button
          key={range.value}
          onClick={() => handleRangeChange(range.value)}
          className={`px-4 py-2 rounded-md text-sm font-medium transition-all duration-200 flex items-center gap-1 ${
            selected === range.value
              ? 'bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-md transform scale-105'
              : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
          }`}
        >
          {range.value === 'custom' && <Calendar className="w-3 h-3" />}
          {range.label}
        </button>
      ))}
      
      {/* Custom Time Range Modal */}
      {showCustomSelector && (
        <CustomTimeRangeSelector
          onApply={handleCustomApply}
          onClose={() => setShowCustomSelector(false)}
        />
      )}
    </div>
  );
};

export default TimeRangeSelector;