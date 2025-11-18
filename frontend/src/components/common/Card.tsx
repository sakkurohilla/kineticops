import React, { ReactNode } from 'react';

interface CardProps {
  children: ReactNode;
  className?: string;
  hover?: boolean;
  padding?: 'none' | 'sm' | 'md' | 'lg';
  variant?: 'default' | 'glass' | 'elevated' | 'gradient';
}

const Card: React.FC<CardProps> = ({ 
  children, 
  className = '', 
  hover = false,
  padding = 'md',
  variant = 'default'
}) => {
  const paddingClasses = {
    none: '',
    sm: 'p-4',
    md: 'p-6',
    lg: 'p-8',
  };

  const variantClasses = {
    default: 'bg-white shadow-md border border-gray-100',
    glass: 'bg-white/70 backdrop-blur-xl shadow-xl border border-white/20',
    elevated: 'bg-white shadow-2xl border border-gray-50',
    gradient: 'bg-gradient-to-br from-white via-blue-50/30 to-purple-50/30 shadow-xl border border-white/50'
  };

  const hoverClass = hover 
    ? 'hover:shadow-2xl hover:scale-[1.02] hover:-translate-y-1 transition-all duration-300 ease-out cursor-pointer' 
    : 'transition-shadow duration-200';

  return (
    <div 
      className={`
        rounded-xl 
        ${variantClasses[variant]} 
        ${paddingClasses[padding]} 
        ${hoverClass} 
        ${className}
      `}
      style={{
        transform: 'translateZ(0)', // Enable 3D acceleration
        backfaceVisibility: 'hidden', // Improve rendering performance
      }}
    >
      {children}
    </div>
  );
};

export default Card;
