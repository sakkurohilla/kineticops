import React from 'react';

interface GlassCardProps {
  children: React.ReactNode;
  className?: string;
  gradient?: string;
}

const GlassCard: React.FC<GlassCardProps> = ({ 
  children, 
  className = '',
  gradient = 'from-blue-500/10 to-purple-500/10'
}) => {
  return (
    <div 
      className={`
        relative overflow-hidden rounded-3xl
        bg-gradient-to-br ${gradient}
        backdrop-blur-xl
        border border-white/20
        shadow-[0_8px_32px_0_rgba(31,38,135,0.15)]
        hover:shadow-[0_12px_48px_0_rgba(31,38,135,0.25)]
        transform hover:-translate-y-1
        transition-all duration-300
        ${className}
      `}
    >
      {/* Glass effect overlay */}
      <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent pointer-events-none" />
      
      {/* Content */}
      <div className="relative z-10">
        {children}
      </div>
    </div>
  );
};

export default GlassCard;
