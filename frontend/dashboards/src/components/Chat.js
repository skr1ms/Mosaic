import React, { useState, useEffect, useRef, useCallback } from "react";
import { useTranslation } from "react-i18next";
import api from "../api/api";
import "./Chat.css";
import ConfirmModal from "./ConfirmModal";
import ChatConfirmDialog from "./ChatConfirmDialog";

const Chat = () => {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);
  const [users, setUsers] = useState([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [allUsers, setAllUsers] = useState([]);
  const [selectedUser, setSelectedUser] = useState(null);
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");
  const [selectedFiles, setSelectedFiles] = useState([]);
  const fileInputRef = useRef(null);
  const [loading, setLoading] = useState(false);
  const [currentUser, setCurrentUser] = useState(null);
  const [unreadCount, setUnreadCount] = useState(0);
  const [partnerUnreadCount, setPartnerUnreadCount] = useState(0);
  const [supportUnreadCount, setSupportUnreadCount] = useState(0);
  const wsRef = useRef(null);
  const reconnectRef = useRef(null);
  const reconnectAttemptsRef = useRef(0);
  const messagesEndRef = useRef(null);
  const messagesContainerRef = useRef(null);
  const selectingRef = useRef(false);
  const lastMouseYRef = useRef(0);
  const messageInputRef = useRef(null);
  const [editingId, setEditingId] = useState(null);
  const [menu, setMenu] = useState({ visible: false, x: 0, y: 0, msg: null });
  const [confirm, setConfirm] = useState({ open: false, messageId: null });
  const [confirmChat, setConfirmChat] = useState({ open: false, chatId: null });
  const [confirmBlock, setConfirmBlock] = useState({
    open: false,
    mode: "block",
  });
  const [templatesOpen, setTemplatesOpen] = useState(false);
  const [unreadBySender, setUnreadBySender] = useState({});
  const [lastActiveMap, setLastActiveMap] = useState({});

  const selectedUserIdRef = useRef(null);
  const currentUserIdRef = useRef(null);

  useEffect(() => {
    const userRole = localStorage.getItem("userRole") || "admin";
    const userId = localStorage.getItem("userId");
    const userEmail = localStorage.getItem("userEmail");
    const userName =
      userRole === "admin" || userRole === "main_admin"
        ? "Administrator"
        : "Partner";

    setCurrentUser({
      id: userId,
      role: userRole,
      email: userEmail,
      name: userName,
    });
  }, []);

  useEffect(() => {
    selectedUserIdRef.current = selectedUser ? selectedUser.id : null;
  }, [selectedUser]);

  useEffect(() => {
    currentUserIdRef.current = currentUser ? currentUser.id : null;
  }, [currentUser]);

  const reorderByActivity = useCallback(
    (list) => {
      try {
        return [...(list || [])].sort((a, b) => {
          const av = Number(lastActiveMap[a.id] || 0);
          const bv = Number(lastActiveMap[b.id] || 0);
          if (bv !== av) return bv - av;

          return 0;
        });
      } catch {
        return list || [];
      }
    },
    [lastActiveMap]
  );

  const bumpUserActivity = useCallback(
    (userId) => {
      if (!userId) return;
      setLastActiveMap((prev) => ({ ...prev, [userId]: Date.now() }));

      setUsers((prev) => reorderByActivity(prev));
      setAllUsers((prev) => reorderByActivity(prev));
    },
    [reorderByActivity]
  );

  const fetchUsers = useCallback(async () => {
    if (!currentUser) return;
    try {
      setLoading(true);
      const qs = searchTerm ? `&search=${encodeURIComponent(searchTerm)}` : "";
      const response = await api.get(
        `/chat/users?role=${currentUser.role}${qs}`
      );
      const usersApi = response.data.users || [];
      const mapped = usersApi.map((u) => ({
        id: u.id,
        name: u.name,
        email: u.email,
        role: u.role,
        isOnline: u.is_online,
        status: u.status || "",
        partnerCode: u.partner_code || "",
        login: u.login || "",
        is_blocked_in_chat: !!u.is_blocked_in_chat,
      }));
      // Append support chats for admin as pseudo-users
      let supportUsers = [];
      if (currentUser.role === "admin" || currentUser.role === "main_admin") {
        try {
          const sres = await api.get("/admin/support/chats");
          const chats = sres.data?.chats || [];

          const uniqueByGuest = [];
          const seenGuests = new Set();
          for (const ch of chats) {
            if (seenGuests.has(ch.guest_id)) continue;
            seenGuests.add(ch.guest_id);
            uniqueByGuest.push(ch);
          }
          supportUsers = uniqueByGuest.map((ch) => ({
            id: ch.id,
            name: ch.title || "Support",
            email: "",
            role: "support",
            isOnline: false,
            status: "",
            partnerCode: "",
            login: "",
            is_blocked_in_chat: false,
          }));
        } catch (_) { }
      }
      const isEmptySearch = String(searchTerm).trim() === "";
      const combined = [...mapped, ...supportUsers];
      const sorted = reorderByActivity(combined);
      setUsers(sorted);
      if (isEmptySearch) setAllUsers(sorted);
    } catch (error) {
      console.error(t("chat.error_loading_users"), error);
      setUsers([]);
    } finally {
      setLoading(false);
    }
  }, [currentUser, searchTerm, reorderByActivity, t]);

  
  const refreshUnreadCount = useCallback(async () => {
    try {
      const res = await api.get("/chat/unread-count");
      const count = res.data?.data?.count ?? 0;
      const dm = Number(count);
      setPartnerUnreadCount(dm);
    } catch (e) {
      
    }
  }, []);

  const refreshSupportUnreadCount = useCallback(async () => {
    const userRole = localStorage.getItem("userRole");
    if (userRole !== "admin" && userRole !== "main_admin") {
      setSupportUnreadCount(0);
      return;
    }

    try {
      const res = await api.get("/chat/support-unread-count");
      const count = res.data?.data?.count ?? 0;
      const sm = Number(count);
      setSupportUnreadCount(sm);
    } catch (e) {
      setSupportUnreadCount(0);
    }
  }, []);

  const fetchMessages = useCallback(
    async (targetUserId) => {
      if (!currentUser || !targetUserId) return;
      try {
        if (selectedUser && selectedUser.role === "support") {
          const response = await api.get("/support/messages", {
            params: { chat_id: targetUserId },
          });
          const msgs = response.data?.messages || [];
          const mapped = msgs.map((m) => ({
            id: String(m.id),
            content: m.content,
            senderId: m.sender_id,
            targetId: String(targetUserId),
            timestamp: m.timestamp,
            read: true,
            attachmentUrl: m.attachment_url || "",
            edited: !!m.edited,
            senderRole: m.sender_role,
          }));
          setMessages(mapped);
        } else {
          const response = await api.get(
            `/chat/messages?targetUserId=${targetUserId}`
          );
          const msgs = response.data.messages || [];
          const mapped = msgs.map((m) => ({
            id: String(m.id),
            content: m.content,
            senderId: m.sender_id,
            targetId: m.target_id,
            timestamp: m.timestamp,
            read: m.read,
            attachmentUrl: m.attachment_url || "",
            edited: !!m.edited,
          }));
          setMessages(mapped);
          refreshUnreadCount();
          refreshSupportUnreadCount();
        }
      } catch (error) {
        console.error(t("chat.error_loading_messages"), error);
        setMessages([]);
      }
    },
    [
      currentUser,
      selectedUser,
      t,
      refreshUnreadCount,
      refreshSupportUnreadCount,
    ]
  );

  const scheduleReconnectRef = useRef();

  const scheduleReconnect = useCallback(() => {
    if (!currentUser) return;
    const attempt = Math.min(reconnectAttemptsRef.current + 1, 6);
    reconnectAttemptsRef.current = attempt;
    const delay = Math.pow(2, attempt) * 1000;
    if (reconnectRef.current) clearTimeout(reconnectRef.current);
    reconnectRef.current = setTimeout(() => {
      if (!wsRef.current && scheduleReconnectRef.current?.connectWebSocket) {
        scheduleReconnectRef.current.connectWebSocket();
      }
    }, delay);
  }, [currentUser]);

  const connectWebSocket = useCallback(() => {
    if (wsRef.current || !currentUser) return;
    const token = localStorage.getItem("token") || "";
    try {
      const apiBase = process.env.REACT_APP_API_URL || window.location.origin;
      const apiUrl = new URL(apiBase);
      const wsProto = apiUrl.protocol === "https:" ? "wss:" : "ws:";
      const wsBase = `${wsProto}//${apiUrl.host}`;
      const url = `${wsBase}/api/ws/chat?token=${encodeURIComponent(token)}`;
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        reconnectAttemptsRef.current = 0;
      };

      ws.onmessage = (event) => {
        try {
          const envelope = JSON.parse(event.data);
          if (envelope?.type === "support_new_message" && envelope.data) {
            const m = envelope.data;
            const curId = currentUserIdRef.current;

            if (m.sender_id !== curId) {
              setSupportUnreadCount((c) => c + 1);
            }
          }
          if (envelope?.type === "message" && envelope.data) {
            const m = envelope.data;
            const msg = {
              id: String(m.id || Date.now()),
              content: m.content,
              senderId: m.sender_id,
              targetId: m.target_id,
              timestamp: m.timestamp || new Date().toISOString(),
              read: m.read,
              attachmentUrl: m.attachment_url || "",
              edited: !!m.edited,
            };
            const selId = selectedUserIdRef.current;
            const curId = currentUserIdRef.current;
            const isForCurrentDialog =
              !!selId && (msg.senderId === selId || msg.targetId === selId);
            const otherId =
              msg.senderId === curId ? msg.targetId : msg.senderId;
            if (otherId) bumpUserActivity(otherId);
            if (isForCurrentDialog) {
              setMessages((prev) => {
                const exists = prev.some((p) => p.id === msg.id);
                return exists
                  ? prev.map((p) => (p.id === msg.id ? { ...p, ...msg } : p))
                  : [...prev, msg];
              });
              refreshUnreadCount();
              refreshSupportUnreadCount();
            } else if (msg.senderId !== curId) {
              setPartnerUnreadCount((c) => c + 1);
            }
          }
          if (envelope?.type === "support_new_message" && envelope.data) {
            const m = envelope.data;
            const selId = selectedUserIdRef.current;

            if (selId && String(selId) === String(m.chat_id || "")) {
              setMessages((prev) => {
                const exists = prev.some((p) => String(p.id) === String(m.id));
                if (exists)
                  return prev.map((p) =>
                    String(p.id) === String(m.id)
                      ? {
                        ...p,
                        content: m.content,
                        attachmentUrl: m.attachment_url || "",
                        edited: !!m.edited,
                        senderRole: m.sender_role,
                      }
                      : p
                  );
                return [
                  ...prev,
                  {
                    id: String(m.id),
                    content: m.content,
                    senderId: m.sender_id,
                    targetId: String(selId),
                    timestamp: m.timestamp,
                    read: true,
                    attachmentUrl: m.attachment_url || "",
                    edited: !!m.edited,
                    senderRole: m.sender_role,
                  },
                ];
              });
            } else {
              setSupportUnreadCount((c) => c + 1);
            }
          }
          if (envelope?.type === "support_message_update" && envelope.data) {
            const m = envelope.data;
            const selId = selectedUserIdRef.current;

            if (selId && String(selId) === String(m.chat_id || "")) {
              setMessages((prev) =>
                prev.map((p) =>
                  String(p.id) === String(m.id)
                    ? {
                      ...p,
                      content: m.content,
                      attachmentUrl: m.attachment_url || "",
                      edited: !!m.edited,
                      senderRole: m.sender_role,
                    }
                    : p
                )
              );
            }
          }
          if (envelope?.type === "read" && envelope.data) {
            const { by_user_id, user_id } = envelope.data;

            const selId = selectedUserIdRef.current;
            const curId = currentUserIdRef.current;
            if (selId && curId && user_id === curId && by_user_id === selId) {
              setMessages((prev) =>
                prev.map((m) =>
                  m.senderId === currentUser.id ? { ...m, read: true } : m
                )
              );
            }
          }

          if (envelope?.type === "support_messages_read" && envelope.data) {
            refreshSupportUnreadCount();
          }
          if (envelope?.type === "message_update" && envelope.data) {
            const { id, content, edited } = envelope.data;
            setMessages((prev) =>
              prev.map((m) =>
                String(m.id) === String(id)
                  ? { ...m, content, edited: !!edited }
                  : m
              )
            );
          }
          if (envelope?.type === "message_delete" && envelope.data) {
            const { id } = envelope.data;
            setMessages((prev) =>
              prev.filter((m) => String(m.id) !== String(id))
            );
          }
          if (envelope?.type === "presence" && envelope.data) {
            const { user_id, online } = envelope.data;
            setUsers((prev) =>
              prev.map((u) =>
                u.id === user_id ? { ...u, isOnline: !!online } : u
              )
            );
          }
        } catch { }
      };

      ws.onclose = () => {
        wsRef.current = null;
        scheduleReconnect();
      };

      ws.onerror = () => {
        try {
          ws.close();
        } catch { }
      };
    } catch (e) {
      scheduleReconnect();
    }
  }, [
    currentUser,
    bumpUserActivity,
    scheduleReconnect,
    refreshUnreadCount,
    refreshSupportUnreadCount,
  ]);

  scheduleReconnectRef.current = { connectWebSocket };

  useEffect(() => {
    if (isOpen) {
      fetchUsers();
    }
  }, [isOpen, currentUser, fetchUsers]);

  useEffect(() => {
    if (!isOpen || !currentUser) return;
    const usersUpdateInterval = setInterval(() => {
      fetchUsers();
    }, 3000);
    return () => clearInterval(usersUpdateInterval);
  }, [isOpen, currentUser, fetchUsers]);

  useEffect(() => {
    if (!currentUser) return;
    connectWebSocket();
    fetchUsers();
    refreshUnreadCount();
    refreshSupportUnreadCount();
    refreshUnreadBySender();

    const unreadInterval = setInterval(refreshUnreadCount, 15000);
    const supportUnreadInterval = setInterval(refreshSupportUnreadCount, 15000);
    const unreadBySenderInterval = setInterval(refreshUnreadBySender, 15000);
    return () => {
      clearInterval(unreadInterval);
      clearInterval(supportUnreadInterval);
      clearInterval(unreadBySenderInterval);
      teardownWebSocket();
    };
  }, [
    selectedUser,
    currentUser,
    fetchUsers,
    connectWebSocket,
    refreshUnreadCount,
    refreshSupportUnreadCount,
  ]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  useEffect(() => {
    setUnreadCount(partnerUnreadCount + supportUnreadCount);
  }, [partnerUnreadCount, supportUnreadCount]);

  useEffect(() => {
    const el = messageInputRef.current;
    if (!el) return;
    const maxH = 200;
    el.style.height = "auto";
    const h = Math.min(el.scrollHeight, maxH);
    el.style.height = h + "px";
    el.style.overflowY = el.scrollHeight > maxH ? "auto" : "hidden";
  }, [newMessage]);

  useEffect(() => {
    const onMouseMove = (e) => {
      lastMouseYRef.current = e.clientY;
    };
    const onMouseUp = () => {
      selectingRef.current = false;
    };
    window.addEventListener("mousemove", onMouseMove);
    window.addEventListener("mouseup", onMouseUp);

    let rafId = 0;
    const tick = () => {
      try {
        const container = messagesContainerRef.current;
        if (container && selectingRef.current) {
          const rect = container.getBoundingClientRect();
          const y = lastMouseYRef.current;
          const margin = 96;
          let delta = 0;
          if (y > rect.bottom - margin && y < rect.bottom + margin) {
            const k = (y - (rect.bottom - margin)) / margin;
            delta = 8 + Math.ceil(32 * k);
          } else if (y < rect.top + margin && y > rect.top - margin) {
            const k = (rect.top + margin - y) / margin;
            delta = -(8 + Math.ceil(32 * k));
          }
          if (delta !== 0) {
            container.scrollTop += delta;
          }
        }
      } catch { }
      rafId = requestAnimationFrame(tick);
    };
    rafId = requestAnimationFrame(tick);

    return () => {
      window.removeEventListener("mousemove", onMouseMove);
      window.removeEventListener("mouseup", onMouseUp);
      if (rafId) cancelAnimationFrame(rafId);
    };
  }, []);

  useEffect(() => {
    if (selectedUser) {
      fetchMessages(selectedUser.id);
    }
  }, [selectedUser, fetchMessages]);

  useEffect(() => {
    if (!isOpen || !selectedUser) return;
    const id = setInterval(() => fetchMessages(selectedUser.id), 3000);
    return () => clearInterval(id);
  }, [isOpen, selectedUser, fetchMessages]);

  const sendMessage = async () => {
    if (!selectedUser || !currentUser) return;
    const text = newMessage.trim();
    const hasFiles = selectedFiles.length > 0;
    if (!text && !hasFiles) return;

    try {
      if (editingId && text) {
        await saveEdit(editingId, text);
        return;
      }
      if (selectedUser.role === "support") {
        if (hasFiles) {
          let baseId = null;
          try {
            const contentText = text || "";
            const createRes = await api.post("/support/messages", {
              chat_id: selectedUser.id,
              content: contentText,
            });
            baseId = String(
              createRes.data?.message?.id || createRes.data?.data?.id || ""
            );
            setNewMessage("");
            if (messageInputRef.current) {
              messageInputRef.current.style.height = "auto";
            }
          } catch (error) {
            console.error("Error creating support message:", error);
          }
          if (baseId) {
            for (const f of selectedFiles) {
              try {
                const fd = new FormData();
                fd.append("file", f);
                await api.post(`/support/messages/${baseId}/attachments`, fd, {
                  headers: { "Content-Type": "multipart/form-data" },
                });
              } catch (error) {
                console.error("Error uploading attachment:", error);
              }
            }
            setTimeout(() => fetchMessages(selectedUser.id), 300);
          }
          setSelectedFiles([]);
          if (fileInputRef.current) fileInputRef.current.value = "";
          return;
        }
        const response = await api.post("/support/messages", {
          chat_id: selectedUser.id,
          content: text,
        });
        const created = response.data?.message || response.data?.data;
        if (created) {
          const message = {
            id: String(created.id),
            content: created.content,
            senderId: created.sender_id,
            targetId: String(selectedUser.id),
            timestamp: created.timestamp,
            read: true,
            attachmentUrl: created.attachment_url || "",
            edited: !!created.edited,
            senderRole: created.sender_role || "admin",
          };
          setMessages((prev) => [...prev, message]);
        }
        setNewMessage("");
        if (messageInputRef.current) {
          messageInputRef.current.style.height = "auto";
        }
        return;
      }
      if (hasFiles) {
        let baseId = null;
        try {
          const contentText = text || "";
          const createRes = await api.post("/chat/messages", {
            target_id: selectedUser.id,
            content: contentText,
          });
          baseId = String(createRes.data?.data?.id || "");
          setNewMessage("");
          if (messageInputRef.current) {
            messageInputRef.current.style.height = "auto";
          }
        } catch { }

        if (baseId) {
          for (const f of selectedFiles) {
            try {
              const fd = new FormData();
              fd.append("file", f);
              console.log(
                "Uploading attachment:",
                f.name,
                "to baseId:",
                baseId
              );
              console.log("File size:", f.size, "bytes, type:", f.type);

              const response = await api.post(
                `/chat/messages/${baseId}/attachments`,
                fd,
                {
                  headers: { "Content-Type": "multipart/form-data" },
                }
              );

              console.log("Attachment uploaded successfully:", {
                fileName: f.name,
                baseId: baseId,
                response: response.data,
              });
            } catch (error) {
              console.error("Attachment upload error:", {
                error: error,
                fileName: f.name,
                baseId: baseId,
                errorMessage: error.message,
              });
            }
          }

          setTimeout(() => fetchMessages(selectedUser.id), 300);
        }
        setSelectedFiles([]);
        if (fileInputRef.current) fileInputRef.current.value = "";
        return;
      }

      const payload = {
        type: "message",
        data: { target_id: selectedUser.id, content: text },
      };
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify(payload));
        setNewMessage("");
        if (messageInputRef.current) {
          messageInputRef.current.style.height = "auto";
        }
        bumpUserActivity(selectedUser.id);
      } else {
        const response = await api.post("/chat/messages", payload.data);
        const created = response.data?.data;
        if (created) {
          const message = {
            id: String(created.id),
            content: created.content,
            senderId: created.sender_id,
            targetId: created.target_id,
            timestamp: created.timestamp,
            read: created.read,
            attachmentUrl: created.attachment_url || "",
            edited: !!created.edited,
          };
          setMessages((prev) => [...prev, message]);
        }
        setNewMessage("");
        if (messageInputRef.current) {
          messageInputRef.current.style.height = "auto";
        }
        bumpUserActivity(selectedUser.id);
      }
    } catch (error) {
      console.error(t("chat.error_sending_message"), error);
    }
  };

  const wsUpdateMessage = (id, content) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({ type: "update", data: { id: Number(id), content } })
      );
      return Promise.resolve();
    }
    return api.patch(`/chat/messages/${id}`, { content });
  };

  const wsDeleteMessage = (id) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({ type: "delete", data: { id: Number(id) } })
      );
      return Promise.resolve();
    }
    return api.delete(`/chat/messages/${id}`);
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  const refreshUnreadBySender = async () => {
    try {
      const res = await api.get("/chat/unread-by-sender");
      const list = res.data?.data || [];
      const map = {};
      list.forEach((row) => {
        map[String(row.sender_id || row.senderId)] = Number(row.count) || 0;
      });
      setUnreadBySender(map);
    } catch (_) { }
  };

  const adminTemplates = [
    "Здравствуйте, я бы хотел стать партнером и получить свою white label страницу, что для этого нужно?",
    "Что такое white label?",
    `Для запуска White Label пришлите, пожалуйста, данные для создания партнёра:
1) Название бренда
2) Домен (ваш домен или желаемый субдомен)
3) Логотип (файл или ссылка)
4) Email для контактов
5) Адрес офиса
6) Телефон
7) Telegram: @username (можно прислать только логин — ссылку сформируем сами)
8) WhatsApp: номер телефона (можно прислать только номер — ссылку сформируем сами)
9) Ссылки на магазины: OZON и/или Wildberries (если есть)
10) Цветовая палитра сайта — до 3 цветов в формате HEX (например: #6D28D9, #0EA5E9, #F59E0B). Если не требуется — оставьте пустым.`,
    `${t("chat.white_label_partnership")}
При подключении партнера по системе White Label в системе создается профиль партнера со следующими параметрами:
    • Доменное имя партнера
    • Название бренда партнера
    • Ссылки на товарные наборы на маркетплейсах (OZON, Wildberries)
    • Контактная информация: email, адрес офиса, телефон
    • Ссылки на мессенджеры (Telegram, WhatsApp)
Когда пользователь переходит по доменному имени партнера, ему отображается основной пользовательский сайт платформы, но с автоматической подстановкой индивидуальных данных партнера:
    • Логотип и название бренда заменяются на партнерские
    • Контактная информация отображается согласно профилю партнера
    • Ссылки на товары ведут на наборы конкретного партнёра
Таким образом, каждый партнер получает персонализированную версию сайта под своим брендом и доменом, при этом используется единая техническая платформа.`,
  ];

  const currentTemplates =
    (currentUser?.role === "admin" || currentUser?.role === "main_admin") &&
      selectedUser?.role === "support"
      ? adminTemplates
      : [];

  const insertTemplate = (text) => {
    setNewMessage((prev) => {
      const base = String(prev || "");
      const sep = base && !base.endsWith("\n") ? "\n" : "";
      return base + sep + text;
    });
    setTemplatesOpen(false);
    setTimeout(() => {
      const el = messageInputRef.current;
      if (el) {
        const maxH = 200;
        el.style.height = "auto";
        const h = Math.min(el.scrollHeight, maxH);
        el.style.height = h + "px";
        el.style.overflowY = el.scrollHeight > maxH ? "auto" : "hidden";
      }
    }, 0);
  };

  const onFileChange = (e) => {
    const files = Array.from(e.target.files || []);
    if (!files.length) return;
    setSelectedFiles((prev) => [...prev, ...files]);
  };
  const removeFileAt = (idx) => {
    setSelectedFiles((prev) => prev.filter((_, i) => i !== idx));
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  const startEdit = (m) => {
    setEditingId(m.id);
    setNewMessage(m.content || "");
    setTimeout(() => {
      if (messageInputRef.current) {
        messageInputRef.current.style.height = "auto";
        messageInputRef.current.style.height =
          Math.min(messageInputRef.current.scrollHeight, 200) + "px";
      }
    }, 0);
    try {
      messageInputRef.current?.focus();
    } catch (_) { }
  };

  const cancelEdit = () => {
    setEditingId(null);
    setNewMessage("");
    if (messageInputRef.current) {
      messageInputRef.current.style.height = "auto";
    }
  };

  const saveEdit = async (id, text) => {
    const content = String(text || "").trim();
    if (!content) return;
    try {
      await wsUpdateMessage(id, content);
      setMessages((prev) =>
        prev.map((m) => (m.id === id ? { ...m, content, edited: true } : m))
      );
      cancelEdit();
    } catch (e) {
      console.error(t("chat.error_editing_message"), e);
    }
  };

  const doDeleteMessage = async (id) => {
    try {
      await wsDeleteMessage(id);
      setMessages((prev) => prev.filter((m) => m.id !== id));
    } catch (e) {
      console.error(t("chat.error_deleting_message"), e);
    }
  };

  const deleteSupportChat = async (chatId) => {
    try {
      await api.delete(`/admin/support/chats/${chatId}`);

      setSelectedUser(null);
      fetchUsers();
    } catch (e) {
      console.error("Error deleting support chat:", e);
    }
  };

  const openContextMenu = (e, message) => {
    e.preventDefault();

    if (message.senderId !== currentUser.id) return;
    const viewportW = window.innerWidth;
    const viewportH = window.innerHeight;
    const menuW = 160;
    const menuH = 84;
    let x = e.clientX;
    let y = e.clientY;
    if (x + menuW > viewportW) x = viewportW - menuW - 8;
    if (y + menuH > viewportH) y = viewportH - menuH - 8;
    setMenu({ visible: true, x, y, msg: message });
  };

  useEffect(() => {
    const hide = () => setMenu((m) => ({ ...m, visible: false }));
    const onClick = () => {
      if (menu.visible) hide();
    };
    const onKey = (ev) => {
      if (ev.key === "Escape") hide();
    };
    window.addEventListener("click", onClick);
    window.addEventListener("keydown", onKey);
    return () => {
      window.removeEventListener("click", onClick);
      window.removeEventListener("keydown", onKey);
    };
  }, [menu.visible]);

  const teardownWebSocket = () => {
    if (reconnectRef.current) {
      clearTimeout(reconnectRef.current);
      reconnectRef.current = null;
    }
    if (wsRef.current) {
      try {
        wsRef.current.close();
      } catch { }
      wsRef.current = null;
    }
  };

  if (!currentUser) return null;

  return (
    <>
      { }
      {menu.visible && menu.msg && (
        <div
          style={{
            position: "fixed",
            top: menu.y,
            left: menu.x,
            width: 160,
            background: "#fff",
            border: "1px solid #e0e0e0",
            borderRadius: 12,
            boxShadow: "0 8px 24px rgba(0,0,0,0.15)",
            zIndex: 2000,
          }}
          onClick={(e) => e.stopPropagation()}
        >
          <button
            className="btn btn-link btn-sm w-100 text-left"
            style={{ padding: "8px 12px" }}
            onClick={() => {
              setMenu({ visible: false, x: 0, y: 0, msg: null });
              startEdit(menu.msg);
            }}
          >
            Изменить
          </button>
          <div style={{ height: 1, background: "#eee" }} />
          <button
            className="btn btn-link btn-sm w-100 text-left text-danger"
            style={{ padding: "8px 12px" }}
            onClick={() => {
              const id = menu.msg?.id;
              setMenu({ visible: false, x: 0, y: 0, msg: null });
              if (id) setConfirm({ open: true, messageId: id });
            }}
          >
            Удалить
          </button>
        </div>
      )}
      { }
      <ConfirmModal
        isOpen={confirm.open}
        title="Подтверждение"
        message="Вы точно уверены, что хотите удалить сообщение?"
        confirmText="Удалить"
        confirmColor="danger"
        onConfirm={() => {
          const id = confirm.messageId;
          setConfirm({ open: false, messageId: null });
          if (id) doDeleteMessage(id);
        }}
        onCancel={() => setConfirm({ open: false, messageId: null })}
        containerSelector=".chat-window"
      />
      <ChatConfirmDialog
        isOpen={!!confirmBlock.open}
        title={
          confirmBlock.mode === "block"
            ? t("chat.block_partner")
            : t("chat.unblock_partner")
        }
        message={
          confirmBlock.mode === "block"
            ? t("chat.partner_blocked_message")
            : t("chat.partner_unblocked_message")
        }
        confirmText={
          confirmBlock.mode === "block" ? "Заблокировать" : "Разблокировать"
        }
        confirmColor={confirmBlock.mode === "block" ? "warning" : "success"}
        onConfirm={async () => {
          const id = selectedUser?.id;
          setConfirmBlock({ open: false, mode: "block" });
          if (!id) return;
          try {
            if (confirmBlock.mode === "block") {
              await api.patch(`/chat/partners/${encodeURIComponent(id)}/block`);

              await api.patch(
                `/admin/partners/${encodeURIComponent(id)}/block`
              );
              setSelectedUser({
                ...selectedUser,
                status: "blocked",
                is_blocked_in_chat: true,
              });
            } else {
              await api.patch(
                `/chat/partners/${encodeURIComponent(id)}/unblock`
              );
              await api.patch(
                `/admin/partners/${encodeURIComponent(id)}/unblock`
              );
              setSelectedUser({
                ...selectedUser,
                status: "active",
                is_blocked_in_chat: false,
              });
            }
            fetchUsers();
          } catch (_) { }
        }}
        onCancel={() => setConfirmBlock({ open: false, mode: "block" })}
      />
      <ChatConfirmDialog
        isOpen={!!confirmChat.open}
        title={t("chat.delete_chat")}
        message={t("chat.delete_chat_message")}
        confirmText={t("common.delete")}
        confirmColor="danger"
        onConfirm={async () => {
          const chatId = confirmChat.chatId;
          setConfirmChat({ open: false, chatId: null });
          if (!chatId) return;
          await deleteSupportChat(chatId);
        }}
        onCancel={() => setConfirmChat({ open: false, chatId: null })}
      />
      <div
        className="chat-button"
        style={{
          position: "fixed",
          bottom: "20px",
          right: "20px",
          zIndex: 1000,
        }}
      >
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="btn btn-primary rounded-circle chat-toggle-btn"
          style={{
            width: "64px",
            height: "64px",
            boxShadow: "0 4px 12px rgba(0,0,0,0.15)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            transition: "all 0.3s ease",
          }}
        >
          {isOpen ? (
            <i className="pe-7s-close" style={{ fontSize: "24px" }}></i>
          ) : (
            <div style={{ position: "relative" }}>
              <i className="pe-7s-chat" style={{ fontSize: "24px" }}></i>
              {unreadCount > 0 && (
                <span
                  style={{
                    position: "absolute",
                    top: -8,
                    right: -10,
                    backgroundColor: "#dc3545",
                    color: "#fff",
                    borderRadius: "10px",
                    padding: "0 6px",
                    fontSize: "11px",
                    lineHeight: "18px",
                    minWidth: 18,
                    textAlign: "center",
                  }}
                >
                  {unreadCount > 99 ? "99+" : unreadCount}
                </span>
              )}
              { }
              {unreadCount > 0 && (
                <span
                  title={`Партнёры: ${partnerUnreadCount} • Support: ${supportUnreadCount}`}
                  style={{
                    position: "absolute",
                    top: 12,
                    right: -10,
                    fontSize: 9,
                    color: "#999",
                  }}
                ></span>
              )}
            </div>
          )}
        </button>
      </div>

      { }
      {isOpen && (
        <div
          className="chat-window"
          style={{
            position: "fixed",
            bottom: "100px",
            right: "20px",
            width: "560px",
            height: "560px",
            backgroundColor: "white",
            borderRadius: "12px",
            boxShadow: "0 8px 32px rgba(0,0,0,0.15)",
            border: "1px solid #e0e0e0",
            display: "flex",
            flexDirection: "column",
            zIndex: 1000,
            overflow: "hidden",
          }}
        >
          { }
          <div
            className="chat-header"
            style={{
              backgroundColor: "#007bff",
              color: "white",
              padding: "15px",
              borderRadius: "12px 12px 0 0",
              borderBottom: "1px solid #e0e0e0",
            }}
          >
            <div className="d-flex justify-content-between align-items-center">
              <h6 className="mb-0">
                {selectedUser
                  ? `${t("chat.chat_with")} ${selectedUser.name}`
                  : t("chat.support_chat")}
              </h6>
              <button
                onClick={() => setIsOpen(false)}
                className="btn btn-sm"
                style={{
                  padding: "2px 8px",
                  backgroundColor: "white",
                  border: "1px solid white",
                  color: "#007bff",
                }}
              >
                <i className="pe-7s-close" style={{ fontWeight: "bold" }}></i>
              </button>
            </div>
          </div>

          { }
          {!selectedUser && (
            <div
              className="chat-users"
              style={{ flex: 1, overflowY: "auto", padding: "15px" }}
            >
              <h6 className="mb-3">
                {currentUser.role === "admin" ||
                  currentUser.role === "main_admin"
                  ? t("chat.partners")
                  : t("chat.administrators")}
              </h6>
              <div className="mb-3">
                <input
                  type="text"
                  className="form-control"
                  placeholder={
                    currentUser.role === "admin" ||
                      currentUser.role === "main_admin"
                      ? t("chat.search_by_code_brand_email")
                      : t("chat.search_by_email_login")
                  }
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
              {loading ? (
                <div className="text-center text-muted">
                  <div
                    className="spinner-border spinner-border-sm"
                    role="status"
                  >
                    <span className="sr-only">{t("chat.loading")}</span>
                  </div>
                  <div className="mt-2">{t("chat.loading")}</div>
                </div>
              ) : (String(searchTerm).trim() === "" ? allUsers : users)
                .length === 0 ? (
                <div className="text-center text-muted">
                  <i
                    className="pe-7s-users"
                    style={{ fontSize: "48px", opacity: 0.3 }}
                  ></i>
                  <div className="mt-2">
                    {currentUser.role === "admin" ||
                      currentUser.role === "main_admin"
                      ? t("chat.no_partners_available")
                      : t("chat.no_administrators_available")}
                  </div>
                </div>
              ) : (
                <div className="list-group list-group-flush">
                  {(String(searchTerm).trim() === "" ? allUsers : users).map(
                    (user) => (
                      <div
                        key={user.id}
                        className="list-group-item list-group-item-action d-flex align-items-center chat-user-item"
                        style={{
                          cursor: "pointer",
                          border: "none",
                          borderBottom: "1px solid #f0f0f0",
                          padding: "12px 0",
                          transition: "background-color 0.2s ease",
                          backgroundColor:
                            (currentUser.role === "admin" ||
                              currentUser.role === "main_admin") &&
                              user.role === "partner" &&
                              user.is_blocked_in_chat
                              ? "#fff3cd"
                              : undefined,
                        }}
                      >
                        <div
                          onClick={() => setSelectedUser(user)}
                          style={{
                            display: "flex",
                            alignItems: "center",
                            flex: 1,
                            cursor: "pointer",
                          }}
                        >
                          <div
                            className={`rounded-circle mr-3 ${user.isOnline ? "bg-success" : "bg-secondary"
                              }`}
                            style={{ width: "12px", height: "12px" }}
                          ></div>
                          <div
                            style={{ flex: 1, marginLeft: 6, lineHeight: 1.1 }}
                          >
                            <div
                              className="font-weight-bold"
                              style={{
                                marginBottom: 0,
                                display: "flex",
                                alignItems: "center",
                                gap: 8,
                              }}
                            >
                              {user.name}
                              {user.role === "partner" &&
                                ((typeof user.status === "string" &&
                                  user.status
                                    .toLowerCase()
                                    .includes("block")) ||
                                  user.is_blocked_in_chat) && (
                                  <span className="badge badge-danger">
                                    {t("chat.blocked")}
                                  </span>
                                )}
                            </div>
                            <small
                              className="text-muted"
                              style={{ display: "block", marginTop: 2 }}
                            >
                              {user.email}
                              {user.partnerCode
                                ? ` · Код: ${user.partnerCode}`
                                : ""}
                              {!user.partnerCode && user.login
                                ? ` · ${user.login}`
                                : ""}
                            </small>
                            {/* превью последних сообщений удалены */}
                          </div>
                        </div>
                        <div
                          style={{
                            display: "flex",
                            alignItems: "center",
                            gap: 8,
                            marginLeft: "auto",
                          }}
                        >
                          {unreadBySender[user.id] > 0 && (
                            <span
                              style={{
                                backgroundColor: "#dc3545",
                                color: "#fff",
                                borderRadius: 12,
                                padding: "0 6px",
                                fontSize: 11,
                                lineHeight: "18px",
                                minWidth: 18,
                                textAlign: "center",
                              }}
                            >
                              {unreadBySender[user.id] > 99
                                ? "99+"
                                : unreadBySender[user.id]}
                            </span>
                          )}
                          {(currentUser.role === "admin" ||
                            currentUser.role === "main_admin") &&
                            user.role === "support" && (
                              <button
                                className="btn btn-sm btn-danger"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  setConfirmChat({
                                    open: true,
                                    chatId: user.id,
                                  });
                                }}
                                title={t("chat.delete_chat")}
                                style={{ padding: "4px 8px", fontSize: "12px" }}
                              >
                                <i className="pe-7s-trash"></i>
                              </button>
                            )}
                        </div>
                      </div>
                    )
                  )}
                </div>
              )}
            </div>
          )}

          { }
          {selectedUser && (
            <>
              { }
              <div
                className="chat-user-header"
                style={{
                  backgroundColor: "#f8f9fa",
                  padding: "15px",
                  borderBottom: "1px solid #e0e0e0",
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <div className="d-flex align-items-center">
                  <div
                    className={`rounded-circle mr-3 ${selectedUser.isOnline ? "bg-success" : "bg-secondary"
                      }`}
                    style={{ width: "12px", height: "12px" }}
                  ></div>
                  <div style={{ marginLeft: 6, lineHeight: 1.1 }}>
                    <div
                      className="font-weight-bold"
                      style={{
                        marginBottom: 0,
                        display: "flex",
                        alignItems: "center",
                        gap: 8,
                      }}
                    >
                      {selectedUser.name}
                      {(currentUser.role === "admin" ||
                        currentUser.role === "main_admin") &&
                        selectedUser.role === "partner" &&
                        (String(selectedUser.status || "")
                          .toLowerCase()
                          .includes("block") ||
                          selectedUser.is_blocked_in_chat) && (
                          <span className="badge badge-danger">
                            заблокирован
                          </span>
                        )}
                    </div>
                    <small
                      className="text-muted"
                      style={{ display: "block", marginTop: 2 }}
                    >
                      {selectedUser.email}
                    </small>
                  </div>
                </div>
                <div className="d-flex align-items-center" style={{ gap: 8 }}>
                  {(currentUser.role === "admin" ||
                    currentUser.role === "main_admin") &&
                    selectedUser.role === "partner" && (
                      <>
                        {selectedUser?.is_blocked_in_chat ||
                          String(selectedUser?.status || "")
                            .toLowerCase()
                            .includes("block") ? (
                          <button
                            className="btn btn-sm btn-success"
                            onClick={() =>
                              setConfirmBlock({ open: true, mode: "unblock" })
                            }
                          >
                            {t("chat.unblock")}
                          </button>
                        ) : (
                          <button
                            className="btn btn-sm btn-warning"
                            onClick={() =>
                              setConfirmBlock({ open: true, mode: "block" })
                            }
                          >
                            {t("chat.block")}
                          </button>
                        )}
                        { }
                      </>
                    )}
                  {(currentUser.role === "admin" ||
                    currentUser.role === "main_admin") &&
                    selectedUser.role === "support" && (
                      <button
                        className="btn btn-sm btn-danger"
                        onClick={() =>
                          setConfirmChat({
                            open: true,
                            chatId: selectedUser.id,
                          })
                        }
                        title={t("chat.delete_chat")}
                      >
                        <i className="pe-7s-trash"></i>
                      </button>
                    )}
                  <button
                    onClick={() => setSelectedUser(null)}
                    className="btn btn-sm btn-outline-secondary"
                  >
                    <i className="pe-7s-back"></i>
                  </button>
                </div>
              </div>

              { }
              <div
                className="chat-messages"
                ref={messagesContainerRef}
                onMouseDown={(e) => {
                  if (e.button !== 0) return;
                  const tag = String(e.target?.tagName || "").toLowerCase();
                  // Не включаем режим автоскролла, если клик по интерактивным элементам
                  if (
                    tag === "textarea" ||
                    tag === "button" ||
                    tag === "input" ||
                    tag === "a"
                  )
                    return;
                  selectingRef.current = true;
                }}
                onMouseLeave={() => {
                  if (selectingRef.current) {
                  }
                }}
                style={{
                  flex: 1,
                  overflowY: "auto",
                  padding: "15px",
                  backgroundColor: "#f8f9fa",
                }}
              >
                {messages.length === 0 ? (
                  <div className="text-center text-muted mt-4">
                    <i
                      className="pe-7s-chat"
                      style={{ fontSize: "48px", opacity: 0.3 }}
                    ></i>
                    <div className="mt-2">{t("chat.start_conversation")}</div>
                  </div>
                ) : (
                  messages.map((message) => {
                    const isMyMessage = message.senderId === currentUser.id;
                    return (
                      <div
                        key={message.id}
                        className="mb-3"
                        style={{
                          display: "flex",
                          justifyContent: isMyMessage
                            ? "flex-end"
                            : "flex-start",
                        }}
                        onContextMenu={(e) => openContextMenu(e, message)}
                      >
                        <div
                          className={`d-inline-block p-3 rounded ${isMyMessage ? "bg-primary text-white" : "bg-white"
                            }`}
                          style={{
                            maxWidth: "80%",
                            boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
                            borderRadius: 12,
                          }}
                        >
                          <>
                            <div
                              style={{
                                whiteSpace: "pre-wrap",
                                wordBreak: "break-word",
                              }}
                            >
                              {(() => {
                                let displayName = "";
                                const attachmentUrl =
                                  message.attachmentUrl ||
                                  message.attachment_url ||
                                  "";
                                try {
                                  const raw = String(attachmentUrl);
                                  const path = raw
                                    .split("?")[0]
                                    .replace(/\\/g, "/");
                                  const last = decodeURIComponent(
                                    path.split("/").pop() || ""
                                  );
                                  const idx = last.indexOf("_");
                                  displayName =
                                    idx > -1 ? last.slice(idx + 1) : last;
                                } catch (_) { }
                                const contentText = String(
                                  message.content || ""
                                ).trim();
                                const isDuplicate =
                                  !!attachmentUrl &&
                                  !!displayName &&
                                  contentText.toLowerCase() ===
                                  displayName.toLowerCase();
                                return (
                                  <>
                                    {!isDuplicate && contentText && (
                                      <span>{message.content}</span>
                                    )}
                                    {attachmentUrl && (
                                      <div
                                        className={
                                          contentText && !isDuplicate
                                            ? "mt-2"
                                            : ""
                                        }
                                      >
                                        <a
                                          href={attachmentUrl}
                                          target="_blank"
                                          rel="noreferrer"
                                          style={{
                                            color: isMyMessage
                                              ? "#ffffff"
                                              : "#0d6efd",
                                          }}
                                          className="chat-attachment-link"
                                        >
                                          {displayName || "Вложение"}
                                        </a>
                                      </div>
                                    )}
                                  </>
                                );
                              })()}
                            </div>
                            <div
                              style={{
                                display: "flex",
                                alignItems: "center",
                                gap: 6,
                                marginTop: 4,
                              }}
                            >
                              <small
                                className={`${isMyMessage ? "text-white-50" : "text-muted"
                                  }`}
                              >
                                {new Date(message.timestamp).toLocaleTimeString(
                                  [],
                                  {
                                    hour: "2-digit",
                                    minute: "2-digit",
                                    hour12: false,
                                  }
                                )}
                                {message.edited ? " · изменено" : ""}
                              </small>
                              {isMyMessage && (
                                <small
                                  className={`${isMyMessage ? "text-white-50" : "text-muted"
                                    }`}
                                  title={
                                    message.read ? "Прочитано" : "Отправлено"
                                  }
                                  style={{ letterSpacing: "-1px" }}
                                >
                                  {message.read ? "✓✓" : "✓"}
                                </small>
                              )}
                            </div>
                          </>
                        </div>
                      </div>
                    );
                  })
                )}
                <div ref={messagesEndRef} />
              </div>

              { }
              <div
                className="chat-input"
                style={{
                  padding: "15px",
                  borderTop: "1px solid #e0e0e0",
                  backgroundColor: "white",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    alignItems: "flex-start",
                    minHeight: "44px",
                  }}
                >
                  <div
                    style={{
                      display: "flex",
                      alignItems: "flex-start",
                      marginRight: 8,
                      marginTop: "2px",
                    }}
                  >
                    <button
                      type="button"
                      className="btn btn-outline-secondary"
                      onClick={() => fileInputRef.current?.click()}
                      title={t("chat.attach_file")}
                      style={{ borderRadius: 12 }}
                    >
                      <i className="pe-7s-paperclip"></i>
                    </button>
                    <input
                      ref={fileInputRef}
                      type="file"
                      multiple
                      style={{ display: "none" }}
                      onChange={onFileChange}
                    />
                  </div>
                  <textarea
                    className="form-control"
                    placeholder={t("chat.type_message")}
                    value={newMessage}
                    onChange={(e) => setNewMessage(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" && !e.shiftKey) {
                        e.preventDefault();
                        sendMessage();
                      }
                    }}
                    ref={messageInputRef}
                    style={{
                      resize: "none",
                      minHeight: "36px",
                      maxHeight: "200px",
                      overflowY: "auto",
                      lineHeight: "1.2",
                    }}
                    rows={1}
                  />
                  <div
                    style={{
                      marginLeft: 8,
                      position: "relative",
                      display: "flex",
                      alignItems: "flex-start",
                      gap: 8,
                      marginTop: "2px",
                    }}
                  >
                    {currentTemplates.length > 0 && (
                      <button
                        type="button"
                        className="btn btn-outline-secondary"
                        title={t("chat.message_templates")}
                        onClick={() => setTemplatesOpen((v) => !v)}
                        style={{ borderRadius: 12 }}
                      >
                        ▾
                      </button>
                    )}
                    <button
                      type="button"
                      className="btn btn-primary"
                      onClick={sendMessage}
                      disabled={
                        (!newMessage.trim() && selectedFiles.length === 0) ||
                        !!selectedUser?.is_blocked_in_chat
                      }
                      style={{ borderRadius: 12 }}
                    >
                      {t("chat.send")}
                    </button>
                    {templatesOpen && (
                      <div
                        className="chat-templates-dropdown"
                        style={{
                          position: "absolute",
                          right: 0,
                          bottom: 48,
                          width: 360,
                          maxHeight: 240,
                          overflowY: "auto",
                          background: "#fff",
                          border: "1px solid #e0e0e0",
                          borderRadius: 12,
                          boxShadow: "0 8px 24px rgba(0,0,0,0.15)",
                          padding: 8,
                          zIndex: 1100,
                        }}
                      >
                        <div
                          style={{
                            fontWeight: 600,
                            padding: "4px 8px",
                            borderBottom: "1px solid #f1f3f5",
                            marginBottom: 6,
                          }}
                        >
                          {t("chat.message_templates")}
                        </div>
                        <div
                          style={{
                            display: "flex",
                            flexDirection: "column",
                            gap: 6,
                          }}
                        >
                          {currentTemplates.map((t, i) => (
                            <button
                              key={i}
                              type="button"
                              className="btn btn-light text-start"
                              style={{
                                whiteSpace: "pre-wrap",
                                textAlign: "left",
                                borderRadius: 10,
                              }}
                              onClick={() => insertTemplate(t)}
                            >
                              {t.length > 120 ? t.slice(0, 120) + "…" : t}
                            </button>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
                {selectedFiles.length > 0 && (
                  <div
                    className="mt-2"
                    style={{ display: "flex", flexWrap: "wrap", gap: 8 }}
                  >
                    {selectedFiles.map((f, idx) => (
                      <div
                        key={idx}
                        className="d-flex align-items-center"
                        style={{
                          gap: 8,
                          background: "#f1f3f5",
                          borderRadius: 8,
                          padding: "4px 8px",
                        }}
                      >
                        <small
                          style={{
                            maxWidth: 260,
                            whiteSpace: "nowrap",
                            overflow: "hidden",
                            textOverflow: "ellipsis",
                          }}
                        >
                          {f.name}
                        </small>
                        <button
                          className="btn btn-sm btn-link text-danger p-0"
                          onClick={() => removeFileAt(idx)}
                          title="Убрать"
                        >
                          ×
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </>
          )}
        </div>
      )}
    </>
  );
};

export default Chat;
