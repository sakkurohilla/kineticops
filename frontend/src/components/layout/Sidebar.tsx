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
      className={`fixed left-0 top-0 h-full bg-white border-r border-gray-200 shadow-lg transition-all duration-300 ease-in-out z-50 ${
        isExpanded ? 'w-64' : 'w-16'
      }`}
      onMouseEnter={() => setIsExpanded(true)}
      onMouseLeave={() => setIsExpanded(false)}
    >
      {/* Logo Section */}
      <div className="flex items-center h-16 px-4 border-b border-gray-100">
        <button 
          onClick={() => navigate('/dashboard')}
          className="flex items-center gap-3 hover:bg-gray-50 rounded-lg p-2 transition-colors w-full"
        >
          <div className="w-8 h-8 bg-gradient-to-br from-blue-600 to-blue-700 rounded-lg flex items-center justify-center">
            <span className="text-white font-bold text-sm">K</span>
          </div>
          <div
            className={`transition-all duration-300 ${
              isExpanded ? 'opacity-100 w-auto' : 'opacity-0 w-0 overflow-hidden'
            }`}
          >
            <span className="text-gray-900 font-semibold text-lg whitespace-nowrap">
              KineticOps Pro
            </span>
          </div>
        </button>
      </div>

      {/* Menu Items */}
      <nav className="flex-1 py-4">
        <ul className="space-y-1 px-2">
          {menuItems.map((item) => {
            const Icon = item.icon;
            const isActive = location.pathname === item.path;

            return (
              <li key={item.id}>
                <button
                  onClick={() => navigate(item.path)}
                  className={`relative flex items-center w-full px-3 py-2.5 rounded-lg transition-all duration-200 group ${
                    isActive
                      ? 'bg-blue-50 text-blue-700 border-r-2 border-blue-700'
                      : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                  }`}
                >
                  <Icon className="w-5 h-5 flex-shrink-0" />
                  <span
                    className={`ml-3 font-medium transition-all duration-300 whitespace-nowrap ${
                      isExpanded ? 'opacity-100' : 'opacity-0 w-0 overflow-hidden'
                    }`}
                  >
                    {item.label}
                  </span>

                  {/* Tooltip for collapsed state */}
                  {!isExpanded && (
                    <div className="absolute left-full ml-4 px-2 py-1 bg-gray-900 text-white text-xs rounded opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 whitespace-nowrap z-50">
                      {item.label}
                    </div>
                  )}
                </button>
              </li>
            );
          })}
        </ul>
      </nav>

      {/* Bottom Section - Logout */}
      <div className="border-t border-gray-100 py-4 px-2">
        <button
          onClick={handleLogout}
          className="relative flex items-center w-full px-3 py-2.5 rounded-lg text-gray-600 hover:bg-red-50 hover:text-red-700 transition-all duration-200 group"
        >
          <LogOut className="w-5 h-5 flex-shrink-0" />
          <span
            className={`ml-3 font-medium transition-all duration-300 whitespace-nowrap ${
              isExpanded ? 'opacity-100' : 'opacity-0 w-0 overflow-hidden'
            }`}
          >
            Logout
          </span>

          {/* Tooltip for collapsed state */}
          {!isExpanded && (
            <div className="absolute left-full ml-4 px-2 py-1 bg-gray-900 text-white text-xs rounded opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 whitespace-nowrap z-50">
              Logout
            </div>
          )}
        </button>
      </div>
    </div>
  );
};

export default Sidebar;
