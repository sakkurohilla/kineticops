import React from 'react';
import { calculatePasswordStrength } from '../../utils/validation';

interface PasswordStrengthProps {
  password: string;
}

const PasswordStrength: React.FC<PasswordStrengthProps> = ({ password }) => {
  if (!password) return null;

  const { strength, score, color } = calculatePasswordStrength(password);
  const percentage = (score / 7) * 100;

  const colorClasses = {
    error: 'bg-error',
    warning: 'bg-warning',
    success: 'bg-success',
    primary: 'bg-primary-600',
  };

  const strengthLabels = {
    'weak': 'Weak',
    'medium': 'Medium',
    'strong': 'Strong',
    'very-strong': 'Very Strong',
  };

  return (
    <div className="mt-2">
      <div className="flex items-center justify-between mb-1">
        <span className="text-xs text-gray-600">Password strength:</span>
        <span className={`text-xs font-medium text-${color}`}>
          {strengthLabels[strength]}
        </span>
      </div>
      <div className="w-full bg-gray-200 rounded-full h-2">
        <div
          className={`${colorClasses[color]} h-2 rounded-full transition-all duration-300`}
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  );
};

export default PasswordStrength;
