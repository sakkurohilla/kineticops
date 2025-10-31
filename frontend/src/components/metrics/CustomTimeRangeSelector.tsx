import React, { useState } from 'react';
import { Calendar, Clock } from 'lucide-react';
import Button from '../common/Button';

interface CustomTimeRangeSelectorProps {
  onApply: (startTime: string, endTime: string) => void;
  onClose: () => void;
}

const CustomTimeRangeSelector: React.FC<CustomTimeRangeSelectorProps> = ({ onApply, onClose }) => {
  const [startDate, setStartDate] = useState('');
  const [startTime, setStartTime] = useState('');
  const [endDate, setEndDate] = useState('');
  const [endTime, setEndTime] = useState('');

  const handleApply = () => {
    if (!startDate || !startTime || !endDate || !endTime) {
      alert('Please fill in all fields');
      return;
    }

    const start = new Date(`${startDate}T${startTime}`);
    const end = new Date(`${endDate}T${endTime}`);

    if (start >= end) {
      alert('Start time must be before end time');
      return;
    }

    onApply(start.toISOString(), end.toISOString());
    onClose();
  };

  const setQuickRange = (hours: number) => {
    const now = new Date();
    const start = new Date(now.getTime() - hours * 60 * 60 * 1000);
    
    setEndDate(now.toISOString().split('T')[0]);
    setEndTime(now.toTimeString().slice(0, 5));
    setStartDate(start.toISOString().split('T')[0]);
    setStartTime(start.toTimeString().slice(0, 5));
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl p-6 w-96">
        <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
          <Calendar className="w-5 h-5" />
          Custom Time Range
        </h3>

        {/* Quick Range Buttons */}
        <div className="mb-4">
          <p className="text-sm font-medium text-gray-700 mb-2">Quick Select:</p>
          <div className="flex gap-2 flex-wrap">
            <button onClick={() => setQuickRange(2)} className="px-3 py-1 text-xs bg-gray-100 hover:bg-gray-200 rounded">Last 2h</button>
            <button onClick={() => setQuickRange(12)} className="px-3 py-1 text-xs bg-gray-100 hover:bg-gray-200 rounded">Last 12h</button>
            <button onClick={() => setQuickRange(48)} className="px-3 py-1 text-xs bg-gray-100 hover:bg-gray-200 rounded">Last 2d</button>
          </div>
        </div>

        {/* Start Time */}
        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-700 mb-2">Start Time</label>
          <div className="grid grid-cols-2 gap-2">
            <input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              className="px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <input
              type="time"
              value={startTime}
              onChange={(e) => setStartTime(e.target.value)}
              className="px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
        </div>

        {/* End Time */}
        <div className="mb-6">
          <label className="block text-sm font-medium text-gray-700 mb-2">End Time</label>
          <div className="grid grid-cols-2 gap-2">
            <input
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              className="px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <input
              type="time"
              value={endTime}
              onChange={(e) => setEndTime(e.target.value)}
              className="px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex gap-3">
          <Button variant="outline" fullWidth onClick={onClose}>
            Cancel
          </Button>
          <Button variant="primary" fullWidth onClick={handleApply}>
            <Clock className="w-4 h-4" />
            Apply Range
          </Button>
        </div>
      </div>
    </div>
  );
};

export default CustomTimeRangeSelector;