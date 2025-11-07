import React, { ReactNode } from 'react';
import Sidebar from './Sidebar';
import Header from './Header';
import WsOverlay from '../common/WsOverlay';

interface MainLayoutProps {
  children: ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  return (
    <div className="flex min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50">
      {/* Sidebar - Fixed position, handled internally */}
      <Sidebar />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col ml-20 transition-all duration-300">
        {/* Header - Sticky */}
        <Header />

        {/* Page Content */}
        <main className="flex-1 overflow-x-hidden">
          {children}
        </main>

        {/* Footer (Optional) */}
        <footer className="bg-white/80 backdrop-blur-sm border-t border-purple-200/50 py-6 px-8">
          <div className="flex flex-col md:flex-row justify-between items-center text-sm">
            <p className="text-gray-600 font-medium">© 2025 KineticOps Pro. Enterprise Monitoring Platform.</p>
            <div className="flex items-center space-x-4 mt-2 md:mt-0">
              <span className="px-3 py-1 bg-gradient-to-r from-green-500 to-emerald-600 text-white text-xs font-semibold rounded-full">v1.0.0</span>
              <span className="text-gray-500">🚀 Production Ready</span>
            </div>
          </div>
        </footer>
        {/* WebSocket diagnostic overlay */}
        <WsOverlay />
      </div>
    </div>
  );
};

export default MainLayout;
