import React from 'react';
import { CheckCircle2, AlertCircle, Info, X } from 'lucide-react';

export default function Toast({ message, type = 'success', onClose }) {
  if (!message) return null;

  const icons = {
    success: <CheckCircle2 size={18} color="#10b981" />,
    error: <AlertCircle size={18} color="#ef4444" />,
    info: <Info size={18} color="#06b6d4" />,
  };

  return (
    <div className="toast-container">
      <div className="toast">
        {icons[type] || icons.info}
        <span>{message}</span>
        {onClose && (
          <button 
            onClick={onClose} 
            style={{ background: 'transparent', color: 'var(--text-dim)', marginLeft: 'auto', display: 'flex', alignItems: 'center' }}
          >
            <X size={16} />
          </button>
        )}
      </div>
    </div>
  );
}
