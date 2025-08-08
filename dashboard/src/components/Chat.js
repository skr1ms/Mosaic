import React, { useState, useEffect, useRef } from 'react';
import api from '../api/api';
import './Chat.css';

const Chat = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [users, setUsers] = useState([]);
  const [selectedUser, setSelectedUser] = useState(null);
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [currentUser, setCurrentUser] = useState(null);
  const messagesEndRef = useRef(null);

  useEffect(() => {
    // Получаем информацию о текущем пользователе
    const userRole = localStorage.getItem('userRole') || 'admin';
    const userId = localStorage.getItem('userId');
    const userEmail = localStorage.getItem('userEmail');
    const userName = userRole === 'admin' ? 'Администратор' : 'Партнер';
    
    setCurrentUser({
      id: userId,
      role: userRole,
      email: userEmail,
      name: userName
    });
  }, []);

  useEffect(() => {
    if (isOpen) {
      fetchUsers();
    }
  }, [isOpen, currentUser]);

  useEffect(() => {
    if (selectedUser) {
      fetchMessages(selectedUser.id);
      // Устанавливаем интервал для обновления сообщений каждые 5 секунд
      const interval = setInterval(() => {
        fetchMessages(selectedUser.id);
      }, 5000);
      
      return () => clearInterval(interval);
    }
  }, [selectedUser]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const fetchUsers = async () => {
    if (!currentUser) return;
    
    try {
      setLoading(true);
      // Получаем список пользователей в зависимости от роли
      const response = await api.get(`/chat/users?role=${currentUser.role}`);
      setUsers(response.data.users || []);
    } catch (error) {
      console.error('Ошибка загрузки пользователей:', error);
      // Если API недоступен, создаем тестовые данные
      setUsers(getMockUsers());
    } finally {
      setLoading(false);
    }
  };

  const getMockUsers = () => {
    const userRole = currentUser?.role || 'admin';
    
    if (userRole === 'admin') {
      // Для админа показываем партнеров
      return [
        { id: '1', name: 'Партнер 1', email: 'partner1@example.com', role: 'partner', isOnline: true },
        { id: '2', name: 'Партнер 2', email: 'partner2@example.com', role: 'partner', isOnline: false },
        { id: '3', name: 'Партнер 3', email: 'partner3@example.com', role: 'partner', isOnline: true },
      ];
    } else {
      // Для партнера показываем админов
      return [
        { id: 'admin1', name: 'Администратор 1', email: 'admin1@example.com', role: 'admin', isOnline: true },
        { id: 'admin2', name: 'Администратор 2', email: 'admin2@example.com', role: 'admin', isOnline: true },
      ];
    }
  };

  const fetchMessages = async (targetUserId) => {
    if (!currentUser || !targetUserId) return;
    
    try {
      const response = await api.get(`/chat/messages?targetUserId=${targetUserId}`);
      setMessages(response.data.messages || []);
    } catch (error) {
      console.error('Ошибка загрузки сообщений:', error);
      // Если API недоступен, создаем тестовые сообщения
      setMessages(getMockMessages(targetUserId));
    }
  };

  const getMockMessages = (targetUserId) => {
    return [
      {
        id: '1',
        content: 'Привет! Как дела?',
        senderId: targetUserId,
        timestamp: new Date(Date.now() - 60000).toISOString()
      },
      {
        id: '2',
        content: 'Все хорошо, спасибо!',
        senderId: currentUser?.id,
        timestamp: new Date(Date.now() - 30000).toISOString()
      }
    ];
  };

  const sendMessage = async () => {
    if (!newMessage.trim() || !selectedUser || !currentUser) return;

    try {
      const messageData = {
        targetUserId: selectedUser.id,
        content: newMessage.trim()
      };

      const response = await api.post('/chat/messages', messageData);
      const message = response.data.message || {
        id: Date.now().toString(),
        content: newMessage.trim(),
        senderId: currentUser.id,
        timestamp: new Date().toISOString()
      };

      setMessages(prev => [...prev, message]);
      setNewMessage('');
    } catch (error) {
      console.error('Ошибка отправки сообщения:', error);
      // Если API недоступен, добавляем сообщение локально
      const message = {
        id: Date.now().toString(),
        content: newMessage.trim(),
        senderId: currentUser.id,
        timestamp: new Date().toISOString()
      };
      setMessages(prev => [...prev, message]);
      setNewMessage('');
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  if (!currentUser) return null;

  return (
    <>
      {/* Кнопка чата */}
      <div className="chat-button" style={{
        position: 'fixed',
        bottom: '20px',
        right: '20px',
        zIndex: 1000
      }}>
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="btn btn-primary rounded-circle chat-toggle-btn"
          style={{
            width: '60px',
            height: '60px',
            boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            transition: 'all 0.3s ease'
          }}
        >
          {isOpen ? (
            <i className="pe-7s-close" style={{ fontSize: '20px' }}></i>
          ) : (
            <i className="pe-7s-chat" style={{ fontSize: '20px' }}></i>
          )}
        </button>
      </div>

      {/* Окно чата */}
      {isOpen && (
        <div className="chat-window" style={{
          position: 'fixed',
          bottom: '100px',
          right: '20px',
          width: '400px',
          height: '500px',
          backgroundColor: 'white',
          borderRadius: '12px',
          boxShadow: '0 8px 32px rgba(0,0,0,0.15)',
          border: '1px solid #e0e0e0',
          display: 'flex',
          flexDirection: 'column',
          zIndex: 1000,
          overflow: 'hidden'
        }}>
          {/* Заголовок */}
          <div className="chat-header" style={{
            backgroundColor: '#007bff',
            color: 'white',
            padding: '15px',
            borderRadius: '12px 12px 0 0',
            borderBottom: '1px solid #e0e0e0'
          }}>
            <div className="d-flex justify-content-between align-items-center">
              <h6 className="mb-0">
                {selectedUser ? `Чат с ${selectedUser.name}` : 'Чат'}
              </h6>
              <button
                onClick={() => setIsOpen(false)}
                className="btn btn-sm"
                style={{ 
                  padding: '2px 8px',
                  backgroundColor: 'white',
                  border: '1px solid white',
                  color: '#007bff'
                }}
              >
                <i className="pe-7s-close" style={{ fontWeight: 'bold' }}></i>
              </button>
            </div>
          </div>

          {/* Список пользователей */}
          {!selectedUser && (
            <div className="chat-users" style={{ flex: 1, overflowY: 'auto', padding: '15px' }}>
              <h6 className="mb-3">
                {currentUser.role === 'admin' ? 'Партнеры' : 'Администраторы'}
              </h6>
              {loading ? (
                <div className="text-center text-muted">
                  <div className="spinner-border spinner-border-sm" role="status">
                    <span className="sr-only">Загрузка...</span>
                  </div>
                  <div className="mt-2">Загрузка...</div>
                </div>
              ) : users.length === 0 ? (
                <div className="text-center text-muted">
                  <i className="pe-7s-users" style={{ fontSize: '48px', opacity: 0.3 }}></i>
                  <div className="mt-2">Нет доступных пользователей</div>
                </div>
              ) : (
                <div className="list-group list-group-flush">
                  {users.map((user) => (
                    <div
                      key={user.id}
                      onClick={() => setSelectedUser(user)}
                      className="list-group-item list-group-item-action d-flex align-items-center chat-user-item"
                      style={{ 
                        cursor: 'pointer',
                        border: 'none',
                        borderBottom: '1px solid #f0f0f0',
                        padding: '12px 0',
                        transition: 'background-color 0.2s ease'
                      }}
                    >
                      <div className={`rounded-circle mr-3 ${user.isOnline ? 'bg-success' : 'bg-secondary'}`} 
                           style={{ width: '12px', height: '12px' }}></div>
                      <div style={{ flex: 1 }}>
                        <div className="font-weight-bold">{user.name}</div>
                        <small className="text-muted">{user.email}</small>
                      </div>
                      <i className="pe-7s-angle-right text-muted"></i>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Чат с выбранным пользователем */}
          {selectedUser && (
            <>
              {/* Заголовок чата */}
              <div className="chat-user-header" style={{
                backgroundColor: '#f8f9fa',
                padding: '15px',
                borderBottom: '1px solid #e0e0e0',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center'
              }}>
                <div className="d-flex align-items-center">
                  <div className={`rounded-circle mr-3 ${selectedUser.isOnline ? 'bg-success' : 'bg-secondary'}`} 
                       style={{ width: '12px', height: '12px' }}></div>
                  <div>
                    <div className="font-weight-bold">{selectedUser.name}</div>
                    <small className="text-muted">{selectedUser.email}</small>
                  </div>
                </div>
                <button
                  onClick={() => setSelectedUser(null)}
                  className="btn btn-sm btn-outline-secondary"
                >
                  <i className="pe-7s-back"></i>
                </button>
              </div>

              {/* Сообщения */}
              <div className="chat-messages" style={{ 
                flex: 1, 
                overflowY: 'auto', 
                padding: '15px',
                backgroundColor: '#f8f9fa'
              }}>
                {messages.length === 0 ? (
                  <div className="text-center text-muted mt-4">
                    <i className="pe-7s-chat" style={{ fontSize: '48px', opacity: 0.3 }}></i>
                    <div className="mt-2">Начните переписку</div>
                  </div>
                ) : (
                  messages.map((message) => (
                    <div
                      key={message.id}
                      className={`mb-3 ${message.senderId === currentUser.id ? 'text-right' : 'text-left'}`}
                    >
                      <div className={`d-inline-block p-3 rounded ${
                        message.senderId === currentUser.id
                          ? 'bg-primary text-white'
                          : 'bg-white'
                      }`} style={{ 
                        maxWidth: '80%',
                        boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
                      }}>
                        <div>{message.content}</div>
                        <small className={`${message.senderId === currentUser.id ? 'text-white-50' : 'text-muted'}`}>
                          {new Date(message.timestamp).toLocaleTimeString()}
                        </small>
                      </div>
                    </div>
                  ))
                )}
                <div ref={messagesEndRef} />
              </div>

              {/* Поле ввода */}
              <div className="chat-input" style={{ 
                padding: '15px', 
                borderTop: '1px solid #e0e0e0',
                backgroundColor: 'white'
              }}>
                <div className="input-group">
                  <input
                    type="text"
                    value={newMessage}
                    onChange={(e) => setNewMessage(e.target.value)}
                    onKeyPress={handleKeyPress}
                    placeholder="Введите сообщение..."
                    className="form-control"
                    style={{ borderRight: 'none' }}
                  />
                  <div className="input-group-append">
                    <button
                      onClick={sendMessage}
                      disabled={!newMessage.trim()}
                      className="btn btn-primary"
                      style={{ borderLeft: 'none' }}
                    >
                      <i className="pe-7s-paper-plane"></i>
                    </button>
                  </div>
                </div>
              </div>
            </>
          )}
        </div>
      )}
    </>
  );
};

export default Chat; 