import React from 'react';
import { Log } from '../../types';

interface LogDetailsProps {
  log: Log;
  onClose: () => void;
}

const LogDetails: React.FC<LogDetailsProps> = ({ log, onClose }) => {
  const copyToClipboard = () => {
    const logText = JSON.stringify(log, null, 2);
    navigator.clipboard.writeText(logText);
    // Could add a toast notification here
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b">
          <h3 className="text-xl font-semibold text-gray-900">Log Details</h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {/* Metadata Section */}
          <div className="mb-6">
            <h4 className="text-sm font-medium text-gray-700 mb-3">Metadata</h4>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <span className="text-xs text-gray-500">Timestamp</span>
                <p className="text-sm text-gray-900 font-mono">{log.timestamp}</p>
              </div>
              <div>
                <span className="text-xs text-gray-500">Level</span>
                <p className="text-sm">
                  <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                    log.level === 'error' ? 'bg-red-100 text-red-800' :
                    log.level === 'warn' ? 'bg-yellow-100 text-yellow-800' :
                    log.level === 'info' ? 'bg-blue-100 text-blue-800' :
                    'bg-gray-100 text-gray-800'
                  }`}>
                    {log.level.toUpperCase()}
                  </span>
                </p>
              </div>
              <div>
                <span className="text-xs text-gray-500">Host</span>
                <p className="text-sm text-gray-900">{log.host_id}</p>
              </div>
              <div>
                <span className="text-xs text-gray-500">Source</span>
                <p className="text-sm text-gray-900 font-mono">{log.source}</p>
              </div>
            </div>
          </div>

          {/* Message Section */}
          <div className="mb-6">
            <h4 className="text-sm font-medium text-gray-700 mb-3">Message</h4>
            <div className="bg-gray-50 rounded-lg p-4 border border-gray-200">
              <p className="text-sm text-gray-900 whitespace-pre-wrap break-words">{log.message}</p>
            </div>
          </div>

          {/* Raw JSON Section */}
          <div>
            <div className="flex items-center justify-between mb-3">
              <h4 className="text-sm font-medium text-gray-700">Raw JSON</h4>
              <button
                onClick={copyToClipboard}
                className="text-sm text-blue-600 hover:text-blue-700 font-medium flex items-center gap-1"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
                Copy to Clipboard
              </button>
            </div>
            <pre className="bg-gray-900 text-gray-100 rounded-lg p-4 overflow-x-auto text-xs font-mono">
              <code>{JSON.stringify(log, null, 2)}</code>
            </pre>
          </div>
        </div>

        {/* Footer */}
        <div className="p-6 border-t bg-gray-50">
          <button
            onClick={onClose}
            className="w-full bg-gray-900 text-white px-4 py-2 rounded-lg hover:bg-gray-800 transition-colors font-medium"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};

export default LogDetails;
