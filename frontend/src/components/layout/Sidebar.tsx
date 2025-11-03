import React, { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Home, Server, BarChart3, FileText, Bell, GitBranch, LogOut, Zap, Globe } from 'lucide-react';
import { useAuth } from '../../hooks/useAuth';

const Sidebar: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { logout } = useAuth();
  const [isExpanded, setIsExpanded] = useState(false);

  const menuItems = [
    { id: 'dashboard', icon: Home, label: 'Dashboard', path: '/dashboard' },
    { id: 'hosts', icon: Server, label: 'Hosts', path: '/hosts' },
    { id: 'apm', icon: Zap, label: 'APM', path: '/apm' },
    { id: 'synthetics', icon: Globe, label: 'Synthetics', path: '/synthetics' },
    { id: 'metrics', icon: BarChart3, label: 'Metrics', path: '/metrics' },
    { id: 'logs', icon: FileText, label: 'Logs', path: '/logs' },
    { id: 'alerts', icon: Bell, label: 'Alerts', path: '/alerts' },
    { id: 'workflow', icon: GitBranch, label: 'Workflow', path: '/workflow' },
  ];

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div
      className={`fixed left-0 top-0 h-full bg-gradient-to-b from-slate-900 via-purple-900 to-slate-900 border-r border-purple-500/20 shadow-2xl transition-all duration-300 ease-in-out z-50 ${
        isExpanded ? 'w-64' : 'w-20'
      }`}
      onMouseEnter={() => setIsExpanded(true)}
      onMouseLeave={() => setIsExpanded(false)}
    >
      {/* Logo Section */}
      <div className="flex items-center h-20 px-4 border-b border-purple-500/20">
        <button 
          onClick={() => navigate('/dashboard')}
          className="flex items-center gap-3 hover:bg-white/10 rounded-2xl p-3 transition-all duration-300 w-full group"
        >
          <div className="w-12 h-12 bg-gradient-to-br from-blue-400 via-purple-500 to-pink-500 rounded-2xl flex items-center justify-center shadow-xl group-hover:shadow-2xl group-hover:shadow-purple-500/50 group-hover:scale-110 transition-all duration-300">
            <span className="text-white font-bold text-xl">🚀</span>
          </div>
          <div
            className={`transition-all duration-300 ${
              isExpanded ? 'opacity-100 w-auto' : 'opacity-0 w-0 overflow-hidden'
            }`}
          >
            <span className="text-white font-bold text-xl whitespace-nowrap bg-gradient-to-r from-white to-purple-200 bg-clip-text text-transparent">
              KineticOps Pro
            </span>
          </div>
        </button>
      </div>

      {/* Menu Items */}
      <nav className="flex-1 py-6">
        <ul className="space-y-3 px-3">
          {menuItems.map((item) => {
            const Icon = item.icon;
            const isActive = location.pathname === item.path;

            return (
              <li key={item.id}>
                <button
                  onClick={() => navigate(item.path)}
                  className={`relative flex items-center w-full px-4 py-3 rounded-2xl transition-all duration-300 group hover:scale-105 ${
                    isActive
                      ? 'bg-gradient-to-r from-blue-500 to-purple-600 text-white shadow-xl shadow-purple-500/30'
                      : 'text-gray-300 hover:bg-white/10 hover:text-white backdrop-blur-sm'
                  }`}
                >
                  <Icon className="w-6 h-6 flex-shrink-0" />
                  <span
                    className={`ml-4 font-semibold transition-all duration-300 whitespace-nowrap ${
                      isExpanded ? 'opacity-100' : 'opacity-0 w-0 overflow-hidden'
                    }`}
                  >
                    {item.label}
                  </span>

                  {/* Glowing dot for active item */}
                  {isActive && (
                    <div className="absolute -right-1 top-1/2 transform -translate-y-1/2 w-2 h-2 bg-white rounded-full animate-pulse"></div>
                  )}

                  {/* Tooltip for collapsed state */}
                  {!isExpanded && (
                    <div className="absolute left-full ml-4 px-4 py-3 bg-gradient-to-r from-gray-900 to-gray-800 text-white text-sm rounded-xl opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-300 whitespace-nowrap z-50 shadow-xl">
                      {item.label}
                      <div className="absolute left-0 top-1/2 transform -translate-y-1/2 -translate-x-2 w-3 h-3 bg-gray-900 rotate-45"></div>
                    </div>
                  )}
                </button>
              </li>
            );
          })}
        </ul>
      </nav>

      {/* Bottom Section - Logout */}
      <div className="border-t border-purple-500/20 py-6 px-3">
        <button
          onClick={handleLogout}
          className="relative flex items-center w-full px-4 py-3 rounded-2xl text-gray-300 hover:bg-gradient-to-r hover:from-red-500 hover:to-pink-600 hover:text-white transition-all duration-300 group hover:scale-105 backdrop-blur-sm"
        >
          <LogOut className="w-6 h-6 flex-shrink-0" />
          <span
            className={`ml-4 font-semibold transition-all duration-300 whitespace-nowrap ${
              isExpanded ? 'opacity-100' : 'opacity-0 w-0 overflow-hidden'
            }`}
          >
            Logout
          </span>

          {/* Tooltip for collapsed state */}
          {!isExpanded && (
            <div className="absolute left-full ml-4 px-4 py-3 bg-gradient-to-r from-gray-900 to-gray-800 text-white text-sm rounded-xl opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-300 whitespace-nowrap z-50 shadow-xl">
              Logout
              <div className="absolute left-0 top-1/2 transform -translate-y-1/2 -translate-x-2 w-3 h-3 bg-gray-900 rotate-45"></div>
            </div>
          )}
        </button>
      </div>
    </div>
  );
};

export default Sidebar;
