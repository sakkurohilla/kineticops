import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';

const Navbar: React.FC = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuth();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <nav className="bg-white shadow-sm border-b border-gray-200">
      <div className="px-8">
        <div className="flex justify-between items-center h-16">
          {/* Page Title / Breadcrumb */}
          <div className="flex items-center">
            <h2 className="text-xl font-semibold text-gray-800">Dashboard</h2>
          </div>

          {/* Right Side: User Info + Logout */}
          <div className="flex items-center space-x-4">
            {user && (
              <>
                <div className="flex items-center space-x-3">
                  <div className="w-10 h-10 rounded-full bg-primary-100 flex items-center justify-center">
                    <span className="text-primary-700 font-bold text-lg">
                      {user.username.charAt(0).toUpperCase()}
                    </span>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-medium text-gray-700">{user.username}</p>
                    <p className="text-xs text-gray-500">{user.email}</p>
                  </div>
                </div>
                <button
                  onClick={handleLogout}
                  className="btn btn-primary px-4 py-2 text-sm"
                >
                  Logout
                </button>
              </>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
