import React, { ReactNode } from 'react';
import Sidebar from './Sidebar';
import Header from './Header';

interface MainLayoutProps {
  children: ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  return (
    <div className="flex min-h-screen bg-gray-50">
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
        <footer className="bg-white border-t border-gray-200 py-4 px-6">
          <div className="flex flex-col md:flex-row justify-between items-center text-sm text-gray-600">
            <p>© 2025 KineticOps. All rights reserved.</p>
            <p>Version 1.0.0</p>
          </div>
        </footer>
      </div>
    </div>
  );
};

export default MainLayout;
