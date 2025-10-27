import React, { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Home, Server, BarChart3, FileText, Bell, Settings, LogOut, HelpCircle } from 'lucide-react';
import { useAuth } from '../../hooks/useAuth';

const Sidebar: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { logout } = useAuth();
  const [isExpanded, setIsExpanded] = useState(false);

  const menuItems = [
    { id: 'dashboard', icon: Home, label: 'Dashboard', path: '/dashboard' },
    { id: 'hosts', icon: Server, label: 'Hosts', path: '/hosts' },
    { id: 'metrics', icon: BarChart3, label: 'Metrics', path: '/metrics' },
    { id: 'logs', icon: FileText, label: 'Logs', path: '/logs' },
    { id: 'alerts', icon: Bell, label: 'Alerts', path: '/alerts' },
    { id: 'settings', icon: Settings, label: 'Settings', path: '/settings' },
    { id: 'help', icon: HelpCircle, label: 'Help', path: '/help' },
  ];

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div
      className={`fixed left-0 top-0 h-full bg-gradient-to-b from-slate-900 to-slate-800 shadow-2xl transition-all duration-300 ease-in-out z-50 ${
        isExpanded ? 'w-64' : 'w-20'
      }`}
      onMouseEnter={() => setIsExpanded(true)}
      onMouseLeave={() => setIsExpanded(false)}
    >
      {/* Logo Section */}
      <div className="flex items-center h-20 px-6 border-b border-slate-700">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center shadow-lg">
            <span className="text-white font-bold text-xl">K</span>
          </div>
          <div
            className={`transition-all duration-300 ${
              isExpanded ? 'opacity-100 w-auto' : 'opacity-0 w-0 overflow-hidden'
            }`}
          >
            <span className="text-white font-semibold text-xl whitespace-nowrap">
              KineticOps
            </span>
            <p className="text-xs text-gray-400 whitespace-nowrap">Infrastructure Monitoring</p>
          </div>
        </div>
      </div>

      {/* Menu Items */}
      <nav className="flex-1 py-6">
        <ul className="space-y-2 px-3">
          {menuItems.map((item) => {
            const Icon = item.icon;
            const isActive = location.pathname === item.path;

            return (
              <li key={item.id}>
                <button
                  onClick={() => navigate(item.path)}
                  className={`relative flex items-center w-full px-4 py-3 rounded-lg transition-all duration-200 group ${
                    isActive
                      ? 'bg-blue-600 text-white shadow-lg'
                      : 'text-gray-300 hover:bg-slate-700 hover:text-white'
                  }`}
                >
                  <Icon className="w-6 h-6 flex-shrink-0" />
                  <span
                    className={`ml-4 font-medium transition-all duration-300 whitespace-nowrap ${
                      isExpanded ? 'opacity-100' : 'opacity-0 w-0 overflow-hidden'
                    }`}
                  >
                    {item.label}
                  </span>

                  {/* Tooltip for collapsed state */}
                  {!isExpanded && (
                    <div className="absolute left-full ml-6 px-3 py-2 bg-gray-900 text-white text-sm rounded-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 whitespace-nowrap shadow-xl z-50">
                      {item.label}
                      <div className="absolute right-full top-1/2 -translate-y-1/2 border-8 border-transparent border-r-gray-900"></div>
                    </div>
                  )}
                </button>
              </li>
            );
          })}
        </ul>
      </nav>

      {/* Bottom Section - Logout */}
      <div className="border-t border-slate-700 py-4 px-3">
        <button
          onClick={handleLogout}
          className="relative flex items-center w-full px-4 py-3 rounded-lg text-gray-300 hover:bg-red-600 hover:text-white transition-all duration-200 group"
        >
          <LogOut className="w-6 h-6 flex-shrink-0" />
          <span
            className={`ml-4 font-medium transition-all duration-300 whitespace-nowrap ${
              isExpanded ? 'opacity-100' : 'opacity-0 w-0 overflow-hidden'
            }`}
          >
            Logout
          </span>

          {/* Tooltip for collapsed state */}
          {!isExpanded && (
            <div className="absolute left-full ml-6 px-3 py-2 bg-gray-900 text-white text-sm rounded-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 whitespace-nowrap shadow-xl z-50">
              Logout
              <div className="absolute right-full top-1/2 -translate-y-1/2 border-8 border-transparent border-r-gray-900"></div>
            </div>
          )}
        </button>

        {/* Version Info */}
        <div
          className={`mt-4 px-4 text-xs text-gray-500 transition-all duration-300 ${
            isExpanded ? 'opacity-100' : 'opacity-0'
          }`}
        >
          <p>Version 1.0.0</p>
          <p className="mt-1">© 2025 KineticOps</p>
        </div>
      </div>
    </div>
  );
};

export default Sidebar;
