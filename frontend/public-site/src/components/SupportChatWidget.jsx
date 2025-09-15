import React from 'react';
import { useTranslation } from 'react-i18next';
import { MosaicAPI } from '../api/client';
import './SupportChat.css';

const SupportChatWidget = () => {
  const { t } = useTranslation();
  const [open, setOpen] = React.useState(false);
  const [messages, setMessages] = React.useState([]);
  const [value, setValue] = React.useState('');
  const [chat, setChat] = React.useState(null); // { chat_id, access_token }
  const [selectedFiles, setSelectedFiles] = React.useState([]);
  const [editingId, setEditingId] = React.useState(null);
  const [menu, setMenu] = React.useState({
    visible: false,
    x: 0,
    y: 0,
    msg: null,
  });
  const fileInputRef = React.useRef(null);
  const textareaRef = React.useRef(null);
  const messagesEndRef = React.useRef(null);
  const refreshTimerRef = React.useRef(null);
  const messageInputRef = React.useRef(null);

  const upsertMessage = React.useCallback(incoming => {
    if (!incoming) return;
    setMessages(prev => {
      const id = String(incoming.id);
      const idx = prev.findIndex(m => String(m.id) === id);
      if (idx >= 0) {
        const copy = prev.slice();
        copy[idx] = { ...prev[idx], ...incoming };
        return copy;
      }
      return [...prev, incoming];
    });
  }, []);

  const fetchMessages = async c => {
    const current = c;
    if (!current) return;
    try {
      const list = await MosaicAPI.getSupportMessages(
        current.chat_id,
        current.access_token
      );
      setMessages(list);
    } catch (err) {
      if (err?.status === 404) {
        // Чат не найден — очищаем локальное состояние, НЕ создаём новый автоматически
        try {
          sessionStorage.removeItem('support:chat');
        } catch {}
        setChat(null);
        setMessages([]);
      } else if (err?.status === 401) {
        try {
          sessionStorage.removeItem('support:chat');
        } catch {}
        setChat(null);
        setMessages([]);
      }
    }
  };

  React.useEffect(() => {
    if (!open) return;
    const init = async () => {
      try {
        let chatData = null;
        try {
          const stored = sessionStorage.getItem('support:chat');
          if (stored) chatData = JSON.parse(stored);
        } catch {}
        if (chatData) {
          setChat(chatData);
          await fetchMessages(chatData);
        } else {
          setChat(null);
          setMessages([]);
        }
      } catch (e) {
        console.error('Support chat init failed', e);
      }
    };
    init();
  }, [open]);

  React.useEffect(() => {
    if (!open || !chat) return;
    if (refreshTimerRef.current) clearInterval(refreshTimerRef.current);
    refreshTimerRef.current = setInterval(() => fetchMessages(chat), 5000);
    return () => {
      if (refreshTimerRef.current) clearInterval(refreshTimerRef.current);
      refreshTimerRef.current = null;
    };
  }, [open, chat]);

  React.useEffect(() => {
    const el = messageInputRef.current;
    if (!el) return;
    const maxH = 200;
    el.style.height = 'auto';
    const h = Math.min(el.scrollHeight, maxH);
    el.style.height = h + 'px';
    el.style.overflowY = el.scrollHeight > maxH ? 'auto' : 'hidden';
  }, [value]);

  React.useEffect(() => {
    if (!open || !chat) return;
    const token = chat.access_token;
    let ws;
    try {
      const apiBase = import.meta.env.VITE_API_BASE_URL || '/api';
      let origin;
      try {
        origin = new URL(apiBase).origin;
      } catch {
        origin = window.location.origin;
      }
      const wsProto = origin.startsWith('https') ? 'wss' : 'ws';
      const wsUrl = `${wsProto}://${origin.replace(/^https?:\/\//, '')}/api/ws/chat?token=${encodeURIComponent(token)}`;
      ws = new WebSocket(wsUrl);
      ws.onmessage = ev => {
        try {
          const envelope = JSON.parse(ev.data);
          if (envelope?.type === 'support_new_message' && envelope.data) {
            const m = envelope.data;
            if (String(m.chat_id || chat.chat_id) === String(chat.chat_id)) {
              upsertMessage(m);
            } else {
            }
          }
          if (envelope?.type === 'support_message_update' && envelope.data) {
            const m = envelope.data;
            if (String(m.chat_id || chat.chat_id) === String(chat.chat_id)) {
              upsertMessage(m);
            }
          }
          if (envelope?.type === 'support_message_delete' && envelope.data) {
            const payload = envelope.data;
            if (
              String(payload.chat_id || chat.chat_id) === String(chat.chat_id)
            ) {
              const delId = String(payload.id);
              setMessages(prev => prev.filter(m => String(m.id) !== delId));
            }
          }
          if (envelope?.type === 'support_messages_read' && envelope.data) {
            const payload = envelope.data;
            if (String(payload.chat_id) === String(chat.chat_id)) {
              setMessages(prev => prev.map(m => ({ ...m, read: true })));
            }
          }
        } catch {}
      };
    } catch {}
    return () => {
      try {
        ws && ws.close();
      } catch {}
    };
  }, [open, chat, upsertMessage]);

  React.useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const cancelEdit = () => {
    setEditingId(null);
    setValue('');
    if (textareaRef.current) textareaRef.current.style.height = 'auto';
  };

  const reinitChat = async () => {
    try {
      try {
        sessionStorage.removeItem('support:chat');
      } catch {}

      setChat(null);
      setMessages([]);
      return null;
    } catch (e) {
      console.error('Failed to reinit support chat', e);
      return null;
    }
  };

  const send = async () => {
    let currentChat = chat;
    if (!currentChat) {
      try {
        const fresh = await MosaicAPI.startSupportChat('');
        try {
          sessionStorage.setItem('support:chat', JSON.stringify(fresh));
        } catch {}
        setChat(fresh);
        currentChat = fresh;
      } catch (e) {
        console.error('failed to start support chat', e);
        return;
      }
    }
    const text = String(value || '').trim();
    const hasFiles = selectedFiles.length > 0;
    if (!text && !hasFiles) return;
    try {
      if (editingId && text) {
        await MosaicAPI.updateSupportMessage(
          editingId,
          text,
          currentChat.access_token
        );
        setMessages(prev =>
          prev.map(m =>
            m.id === editingId ? { ...m, content: text, edited: true } : m
          )
        );
        cancelEdit();
        // Refresh to sync with backend
        setTimeout(() => fetchMessages(currentChat), 150);
        return;
      }
      if (hasFiles) {
        let baseId = null;
        try {
          // For attachment-only messages, send empty content to avoid duplication
          const contentText = text || '';
          const created = await MosaicAPI.sendSupportMessage(
            currentChat.chat_id,
            contentText,
            currentChat.access_token
          );
          baseId = created?.id;
          setValue('');
          if (textareaRef.current) textareaRef.current.style.height = 'auto';
        } catch (e) {
          if (e?.status === 404) {
            const fresh = await reinitChat();
            if (fresh) {
              try {
                const contentText = text || '';
                const created2 = await MosaicAPI.sendSupportMessage(
                  fresh.chat_id,
                  contentText,
                  fresh.access_token
                );
                baseId = created2?.id;
                currentChat = fresh;
              } catch (_) {}
            }
          } else if (e?.status === 401) {
            try {
              sessionStorage.removeItem('support:chat');
            } catch {}
            setChat(null);
            setMessages([]);
            return;
          }
        }
        if (baseId) {
          for (const f of selectedFiles) {
            try {
              const fd = new FormData();
              fd.append('file', f);
              const token =
                currentChat?.access_token ||
                (sessionStorage.getItem('support:chat')
                  ? JSON.parse(sessionStorage.getItem('support:chat') || '{}')
                      .access_token || ''
                  : '');
              console.log(
                'Uploading attachment:',
                f.name,
                'to baseId:',
                baseId,
                'with token:',
                token?.substring(0, 10) + '...'
              );
              console.log('File size:', f.size, 'bytes, type:', f.type);

              const response = await fetch(
                `${import.meta.env.VITE_API_BASE_URL || '/api'}/public/support/messages/${baseId}/attachments`,
                {
                  method: 'POST',
                  body: fd,
                  headers: {
                    Authorization: `Bearer ${token}`,
                  },
                }
              );

              let result;
              try {
                result = await response.json();
              } catch (e) {
                result = await response.text();
              }

              console.log('Attachment upload result:', response.status, result);
              if (!response.ok) {
                console.error('Attachment upload failed:', {
                  status: response.status,
                  statusText: response.statusText,
                  result: result,
                  fileName: f.name,
                  fileSize: f.size,
                  baseId: baseId,
                });
              } else {
                console.log('Attachment uploaded successfully:', {
                  fileName: f.name,
                  baseId: baseId,
                  result: result,
                });
              }
            } catch (error) {
              console.error('Attachment upload error:', {
                error: error,
                fileName: f.name,
                baseId: baseId,
                errorMessage: error.message,
              });
            }
          }
          setTimeout(() => fetchMessages(currentChat), 300);
        }
        setSelectedFiles([]);
        if (fileInputRef.current) fileInputRef.current.value = '';
        return;
      }
      let created = null;
      try {
        created = await MosaicAPI.sendSupportMessage(
          currentChat.chat_id,
          text,
          currentChat.access_token
        );
      } catch (e) {
        if (e?.status === 404) {
          // создаём чат только при первой попытке отправить сообщение явно
          const fresh = await MosaicAPI.startSupportChat('');
          try {
            sessionStorage.setItem('support:chat', JSON.stringify(fresh));
          } catch {}
          setChat(fresh);
          created = await MosaicAPI.sendSupportMessage(
            fresh.chat_id,
            text,
            fresh.access_token
          );
          currentChat = fresh;
        } else if (e?.status === 401) {
          try {
            sessionStorage.removeItem('support:chat');
          } catch {}
          setChat(null);
          setMessages([]);
          return;
        } else {
          throw e;
        }
      }
      if (created) upsertMessage(created);
      setValue('');
      if (textareaRef.current) textareaRef.current.style.height = 'auto';
    } catch (e) {
      console.error('send failed', e);
    }
  };

  const handleAttachClick = async () => {
    if (!chat) {
      try {
        const fresh = await MosaicAPI.startSupportChat('');
        try {
          sessionStorage.setItem('support:chat', JSON.stringify(fresh));
        } catch {}
        setChat(fresh);
      } catch (e) {
        console.error('failed to start support chat', e);
        return;
      }
    }
    fileInputRef.current?.click();
  };

  const handleFileSelected = e => {
    const files = Array.from(e.target.files || []);
    if (!files.length) return;
    setSelectedFiles(prev => [...prev, ...files]);
    if (fileInputRef.current) fileInputRef.current.value = '';
  };

  const removeFileAt = idx => {
    setSelectedFiles(prev => prev.filter((_, i) => i !== idx));
  };

  const openContextMenu = (e, m) => {
    e.preventDefault();
    const mine = m.sender_role !== 'admin' && m.sender_role !== 'main_admin';
    if (!mine) return;
    const vw = window.innerWidth;
    const vh = window.innerHeight;
    const menuW = 160;
    const menuH = 80;
    let x = e.clientX;
    let y = e.clientY;
    if (x + menuW > vw) x = vw - menuW - 8;
    if (y + menuH > vh) y = vh - menuH - 8;
    setMenu({ visible: true, x, y, msg: m });
  };

  React.useEffect(() => {
    const onClick = () => setMenu(m => ({ ...m, visible: false }));
    const onKey = ev => {
      if (ev.key === 'Escape') setMenu(m => ({ ...m, visible: false }));
    };
    window.addEventListener('click', onClick);
    window.addEventListener('keydown', onKey);
    return () => {
      window.removeEventListener('click', onClick);
      window.removeEventListener('keydown', onKey);
    };
  }, []);

  return (
    <div>
      {menu.visible && menu.msg && (
        <div
          style={{
            position: 'fixed',
            top: menu.y,
            left: menu.x,
            width: 160,
            zIndex: 2000,
          }}
          className="support-context-menu"
          onClick={e => e.stopPropagation()}
        >
          <button
            onClick={() => {
              setMenu({ visible: false, x: 0, y: 0, msg: null });
              setEditingId(menu.msg.id);
              setValue(menu.msg.content || '');
              setTimeout(() => {
                if (textareaRef.current) {
                  textareaRef.current.focus();
                  textareaRef.current.style.height = 'auto';
                  textareaRef.current.style.height =
                    Math.min(textareaRef.current.scrollHeight, 200) + 'px';
                }
              }, 0);
            }}
          >
            {t('support_chat.edit_message')}
          </button>
          <div className="support-context-menu-divider" />
          <button
            className="danger"
            onClick={async () => {
              const id = menu.msg?.id;
              setMenu({ visible: false, x: 0, y: 0, msg: null });
              if (!id || !chat) return;
              try {
                await MosaicAPI.deleteSupportMessage(id, chat.access_token);
                setMessages(prev => prev.filter(m => m.id !== id));
              } catch {}
            }}
          >
            {t('support_chat.delete_message')}
          </button>
        </div>
      )}
      <button
        onClick={() => setOpen(v => !v)}
        style={{ position: 'fixed', right: 20, bottom: 20, zIndex: 1000 }}
        className="support-chat-button"
        aria-label={t('support_chat.title')}
        title={t('support_chat.button_title')}
      >
        {open ? '×' : '💬'}
      </button>
      {open && (
        <div
          style={{
            position: 'fixed',
            right: 20,
            bottom: 100,
            width: 560,
            height: 560,
            zIndex: 1000,
          }}
          className="support-chat-window bg-white rounded-xl shadow-xl border flex flex-col"
        >
          <div className="px-4 py-3 border-b font-semibold flex items-center justify-between bg-gray-50 rounded-t-xl">
            <span className="text-gray-800">{t('support_chat.title')}</span>
            <button
              className="text-gray-500 hover:text-gray-700 text-xl font-normal"
              onClick={() => setOpen(false)}
              title={t('support_chat.close')}
            >
              ×
            </button>
          </div>
          <div
            className="flex-1 overflow-auto p-4 space-y-3 support-chat-messages"
            onScroll={() => {}}
          >
            {messages.map(m => {
              const mine =
                m.sender_role !== 'admin' && m.sender_role !== 'main_admin';

              let displayName = '';
              const attachmentUrl = m.attachment_url || m.attachmentUrl || '';
              if (attachmentUrl) {
                try {
                  const raw = String(attachmentUrl);
                  const path = raw.split('?')[0].replace(/\\/g, '/');
                  const last = decodeURIComponent(path.split('/').pop() || '');
                  const idx = last.indexOf('_');
                  displayName = idx > -1 ? last.slice(idx + 1) : last;
                } catch (_) {
                  displayName = t('support_chat.view_attachment');
                }
              }

              const contentText = String(m.content || '').trim();
              const isDuplicate =
                !!attachmentUrl &&
                !!displayName &&
                contentText.toLowerCase() === displayName.toLowerCase();

              return (
                <div
                  key={m.id}
                  style={{
                    display: 'flex',
                    justifyContent: mine ? 'flex-end' : 'flex-start',
                  }}
                >
                  <div
                    className={`support-message-bubble ${mine ? 'mine' : 'theirs'}`}
                    onContextMenu={e => openContextMenu(e, m)}
                  >
                    {!isDuplicate && contentText && (
                      <div className="support-message-text">{m.content}</div>
                    )}
                    {attachmentUrl && (
                      <div
                        className={`${contentText && !isDuplicate ? 'mt-2' : ''}`}
                      >
                        <a
                          href={attachmentUrl}
                          target="_blank"
                          rel="noreferrer"
                          className="support-attachment-link"
                          style={{
                            color: mine ? '#ffffff' : '#3b82f6',
                            backgroundColor: mine
                              ? 'rgba(255,255,255,0.1)'
                              : 'rgba(59,130,246,0.1)',
                          }}
                        >
                          📎 {displayName || t('support_chat.view_attachment')}
                        </a>
                      </div>
                    )}
                    <div className="support-message-meta">
                      <span className="support-message-time">
                        {new Date(m.timestamp || Date.now()).toLocaleTimeString(
                          [],
                          { hour: '2-digit', minute: '2-digit', hour12: false }
                        )}
                        {m.edited ? ` · ${t('support_chat.edited')}` : ''}
                      </span>
                      {mine && (
                        <span
                          className="support-message-status"
                          title={
                            m.read
                              ? t('support_chat.message_read', 'Прочитано')
                              : t('support_chat.message_sent', 'Отправлено')
                          }
                        >
                          {m.read ? '✓✓' : '✓'}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
            <div ref={messagesEndRef} />
          </div>
          <div className="p-4 border-t flex flex-col gap-3 bg-gray-50 rounded-b-xl">
            <div className="flex items-start gap-3">
              <button
                onClick={handleAttachClick}
                className="support-attach-button"
                title={t('support_chat.attach_file')}
              >
                📎
              </button>
              <input
                ref={fileInputRef}
                type="file"
                multiple
                className="hidden"
                onChange={handleFileSelected}
              />
              <textarea
                ref={textareaRef}
                value={value}
                placeholder={
                  editingId
                    ? t('support_chat.edit_placeholder')
                    : t('support_chat.message_placeholder')
                }
                className="support-message-input flex-1"
                rows={1}
                style={{
                  resize: 'none',
                  minHeight: 40,
                  maxHeight: 160,
                  overflowY: 'auto',
                }}
                onChange={e => {
                  setValue(e.target.value);
                  const el = e.target;
                  el.style.height = 'auto';
                  el.style.height = Math.min(el.scrollHeight, 160) + 'px';
                }}
                onKeyDown={e => {
                  if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault();
                    send();
                  }
                  if (e.key === 'Escape' && editingId) {
                    e.preventDefault();
                    cancelEdit();
                  }
                }}
              />
              <button
                onClick={send}
                className="support-send-button"
                disabled={!value.trim() && selectedFiles.length === 0}
              >
                {editingId
                  ? t('support_chat.save_button')
                  : t('support_chat.send_button')}
              </button>
            </div>
            {selectedFiles.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {selectedFiles.map((f, idx) => (
                  <div key={idx} className="support-file-preview">
                    <span className="max-w-[220px] truncate" title={f.name}>
                      📎 {f.name}
                    </span>
                    <button
                      className="support-file-remove"
                      onClick={() => removeFileAt(idx)}
                      title={t('support_chat.remove_file')}
                    >
                      ×
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default SupportChatWidget;
