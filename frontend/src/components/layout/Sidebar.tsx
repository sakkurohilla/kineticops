import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';

const Sidebar: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const menuItems = [
    { icon: '📊', label: 'Dashboard', path: '/dashboard' },
    { icon: '🖥️', label: 'Hosts', path: '/hosts' },
    { icon: '📈', label: 'Metrics', path: '/metrics' },
    { icon: '📝', label: 'Logs', path: '/logs' },
    { icon: '🚨', label: 'Alerts', path: '/alerts' },
    { icon: '⚙️', label: 'Settings', path: '/settings' },
  ];

  return (
    <aside className="w-64 bg-dark min-h-screen text-white">
      <div className="p-6">
        <h1 className="text-2xl font-bold text-primary-400">KineticOps</h1>
        <p className="text-xs text-gray-400 mt-1">Infrastructure Monitoring</p>
      </div>

      <nav className="mt-6">
        {menuItems.map((item) => (
          <button
            key={item.path}
            onClick={() => navigate(item.path)}
            className={`w-full flex items-center px-6 py-3 text-left transition-colors ${
              location.pathname === item.path
                ? 'bg-primary-600 text-white border-l-4 border-primary-400'
                : 'text-gray-300 hover:bg-gray-800 hover:text-white'
            }`}
          >
            <span className="text-xl mr-3">{item.icon}</span>
            <span className="font-medium">{item.label}</span>
          </button>
        ))}
      </nav>

      <div className="absolute bottom-0 w-64 p-6 border-t border-gray-700">
        <div className="text-xs text-gray-500">
          <p>Version 1.0.0</p>
          <p className="mt-1">© 2025 KineticOps</p>
        </div>
      </div>
    </aside>
  );
};

export default Sidebar;
