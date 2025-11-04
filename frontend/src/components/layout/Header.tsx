import React, { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { Bell, Settings, LogOut, User, ChevronDown } from 'lucide-react';
import wsStatus from '../../utils/wsStatus';

const Header: React.FC = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const [showDropdown, setShowDropdown] = useState(false);
  const [wsState, setWsState] = useState<'disconnected'|'connecting'|'connected'|'error'|'reconnecting'>('disconnected');
  const [wsInfo, setWsInfo] = useState<string | undefined>(undefined);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  // Close dropdown when clicking outside
  useEffect(() => {
    const unsub = wsStatus.subscribeWsStatus((s, info) => {
      setWsState(s);
      setWsInfo(info);
    });
    // initialize current state
    const cur = wsStatus.getWsStatus();
    setWsState(cur.state);
    setWsInfo(cur.info);

    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowDropdown(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      unsub();
    };
  }, []);

  // helper for indicator color/title
  const wsTitle = () => {
    switch (wsState) {
      case 'connected':
        return 'Realtime: connected';
      case 'connecting':
        return 'Realtime: connecting';
      case 'reconnecting':
        return `Realtime: reconnecting (${wsInfo || ''})`;
      case 'error':
        return `Realtime: error (${wsInfo || ''})`;
      default:
        return 'Realtime: disconnected';
    }
  };
  

  return (
    <header className="bg-white shadow-sm border-b border-gray-200 sticky top-0 z-40">
      <div className="px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Dashboard Title with Live Status */}
          <div className="flex items-center space-x-6">
            <div>
              <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                🚀 KineticOps Dashboard
              </h1>
              <p className="text-gray-600 mt-1">Real-time infrastructure monitoring • Enterprise scale</p>
            </div>
            
            {/* Live Indicator */}
            <div className="flex items-center space-x-2 px-4 py-2 bg-gradient-to-r from-green-50 to-emerald-50 rounded-full border border-green-200">
              <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse shadow-lg"></div>
              <span className="text-sm font-semibold text-green-700">LIVE</span>
            </div>
            

          </div>

          {/* Right Side */}
          <div className="flex items-center space-x-4">
            {/* Websocket status indicator */}
            <div title={wsTitle()} className="flex items-center mr-2">
              <span
                className={`w-3 h-3 rounded-full mr-2 transition-colors 
                  ${wsState === 'connected' ? 'bg-green-500' : wsState === 'connecting' || wsState === 'reconnecting' ? 'bg-yellow-400' : wsState === 'error' ? 'bg-red-500' : 'bg-gray-300'}`}
              />
              <span className="text-xs text-gray-500 hidden sm:inline">Realtime</span>
            </div>
            {/* Notifications */}
            <button className="relative p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors">
              <Bell className="w-5 h-5" />
              <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full"></span>
            </button>

            {/* Settings */}
            <button 
              onClick={() => navigate('/settings')}
              className="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <Settings className="w-5 h-5" />
            </button>

            {/* User Dropdown */}
            {user && (
              <div className="relative" ref={dropdownRef}>
                <button
                  onClick={() => setShowDropdown(!showDropdown)}
                  className="flex items-center space-x-3 px-3 py-2 rounded-lg hover:bg-gray-100 transition-colors"
                >
                  <div className="w-9 h-9 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-lg">
                    <span className="text-white font-bold text-sm">
                      {user.username.charAt(0).toUpperCase()}
                    </span>
                  </div>
                  <div className="text-left hidden md:block">
                    <p className="text-sm font-medium text-gray-700">{user.username}</p>
                    <p className="text-xs text-gray-500 truncate max-w-[150px]">{user.email}</p>
                  </div>
                  <ChevronDown className={`w-4 h-4 text-gray-500 transition-transform duration-200 ${showDropdown ? 'rotate-180' : ''}`} />
                </button>

                {/* Dropdown Menu */}
                {showDropdown && (
                  <div className="absolute right-0 mt-2 w-56 bg-white rounded-lg shadow-xl border border-gray-200 py-1 z-50">
                    <div className="px-4 py-3 border-b border-gray-200">
                      <p className="text-sm font-medium text-gray-900">{user.username}</p>
                      <p className="text-xs text-gray-500 truncate">{user.email}</p>
                    </div>

                    <button
                      onClick={() => {
                        navigate('/profile');
                        setShowDropdown(false);
                      }}
                      className="w-full flex items-center gap-3 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition-colors"
                    >
                      <User className="w-4 h-4" />
                      Profile
                    </button>

                    <button
                      onClick={() => {
                        navigate('/settings');
                        setShowDropdown(false);
                      }}
                      className="w-full flex items-center gap-3 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition-colors"
                    >
                      <Settings className="w-4 h-4" />
                      Settings
                    </button>

                    <div className="border-t border-gray-200 mt-1 pt-1">
                      <button
                        onClick={handleLogout}
                        className="w-full flex items-center gap-3 px-4 py-2 text-sm text-red-600 hover:bg-red-50 transition-colors"
                      >
                        <LogOut className="w-4 h-4" />
                        Logout
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header;
