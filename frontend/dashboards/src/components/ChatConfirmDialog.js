import React from 'react';

const ChatConfirmDialog = ({
  isOpen,
  title,
  message,
  confirmText = 'Подтвердить',
  cancelText = 'Отмена',
  onConfirm,
  onCancel,
  confirmColor = 'danger'
}) => {
  if (!isOpen) return null;
  return (
    <div
      className="chat-confirm-overlay"
      style={{
        position: 'absolute',
        inset: 0,
        background: 'rgba(0,0,0,0.4)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 2000,
      }}
      onClick={onCancel}
    >
      <div
        className="chat-confirm-dialog"
        style={{
          background: '#fff',
          borderRadius: 12,
          minWidth: 360,
          maxWidth: '85%',
          boxShadow: '0 12px 32px rgba(0,0,0,0.2)',
          overflow: 'hidden',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <div style={{ padding: '12px 16px', borderBottom: '1px solid #e9ecef', fontWeight: 600 }}>
          {title || 'Подтверждение'}
        </div>
        <div style={{ padding: 16 }}>
          <p style={{ margin: 0 }}>{message}</p>
        </div>
        <div style={{ padding: 12, borderTop: '1px solid #e9ecef', display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
          <button className="btn btn-sm btn-secondary" onClick={onCancel}>{cancelText}</button>
          <button
            className={`btn btn-sm btn-${confirmColor}`}
            onClick={onConfirm}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ChatConfirmDialog;


