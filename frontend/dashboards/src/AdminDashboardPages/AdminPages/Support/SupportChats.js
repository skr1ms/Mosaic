import React, { useEffect, useState } from 'react'
import { Row, Col, Card, CardBody, CardHeader, CardTitle, Button, Table, Modal, ModalHeader, ModalBody, ModalFooter } from 'reactstrap'
import api from '../../../api/api'

const SupportChats = () => {
  const [chats, setChats] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [selected, setSelected] = useState(null)
  const [messages, setMessages] = useState([])

  const loadChats = async () => {
    try {
      setLoading(true)
      setError('')
      const { data } = await api.get('/admin/support/chats')
      setChats(data?.chats || [])
    } catch (e) {
      setError(e?.message || 'Failed to load support chats')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { loadChats() }, [])

  
  useEffect(() => {
    const id = setInterval(() => {
      loadChats()
    }, 10000)
    return () => clearInterval(id)
  }, [])

  const openChat = async (chat) => {
    setSelected(chat)
    try {
      const { data } = await api.get('/support/messages', { params: { chat_id: chat.id } })
      setMessages(data?.messages || [])
    } catch (_) {
      setMessages([])
    }
  }

  
  useEffect(() => {
    if (!selected) return
    let disposed = false
    const tick = async () => {
      try {
        const { data } = await api.get('/support/messages', { params: { chat_id: selected.id } })
        if (!disposed) setMessages(data?.messages || [])
      } catch (_) {}
    }
    tick()
    const id = setInterval(tick, 5000)
    return () => {
      disposed = true
      clearInterval(id)
    }
  }, [selected])

  const deleteChat = async (id) => {
    if (!window.confirm('Удалить чат поддержки?')) return
    try {
      await api.delete(`/admin/support/chats/${id}`)
      setSelected(null)
      await loadChats()
    } catch (e) {
      alert(e?.message || 'Failed to delete support chat')
    }
  }

  return (
    <Row>
      <Col lg="12">
        <Card className="main-card mb-3">
          <CardHeader>
            <CardTitle>Support чаты</CardTitle>
            <div className="ms-auto">
              <Button size="sm" color="secondary" onClick={loadChats}>Обновить</Button>
            </div>
          </CardHeader>
          <CardBody>
            {loading && <div>Загрузка...</div>}
            {error && <div className="alert alert-danger">{error}</div>}
            {!loading && !error && (
              <div className="table-responsive">
                <Table hover>
                  <thead>
                    <tr>
                      <th>Название</th>
                      <th>Гость</th>
                      <th>Создан</th>
                      <th>Действия</th>
                    </tr>
                  </thead>
                  <tbody>
                    {chats.map(ch => (
                      <tr key={ch.id}>
                        <td>{ch.title}</td>
                        <td><code>{ch.guest_id}</code></td>
                        <td>{new Date(ch.created_at).toLocaleString('ru-RU', { hour12: false })}</td>
                        <td>
                          <div className="btn-group">
                            <Button size="sm" color="info" onClick={() => openChat(ch)}>Открыть</Button>
                            <Button size="sm" color="danger" onClick={() => deleteChat(ch.id)}>Удалить</Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </div>
            )}
          </CardBody>
        </Card>
      </Col>

      <Modal isOpen={!!selected} toggle={() => setSelected(null)} size="lg">
        <ModalHeader toggle={() => setSelected(null)}>
          Чат: {selected?.title}
        </ModalHeader>
        <ModalBody>
          <div style={{ maxHeight: 420, overflowY: 'auto' }}>
            {messages.map(m => (
              <div key={m.id} className={`p-2 mb-2 rounded ${m.sender_role === 'admin' ? 'bg-light' : 'bg-primary text-white'}`}>
                <div>{m.content}</div>
                <div className="text-muted small">{new Date(m.timestamp).toLocaleString('ru-RU', { hour12: false })}</div>
              </div>
            ))}
          </div>
        </ModalBody>
        <ModalFooter>
          <Button color="danger" onClick={() => deleteChat(selected.id)}>Удалить чат</Button>
          <Button color="secondary" onClick={() => setSelected(null)}>Закрыть</Button>
        </ModalFooter>
      </Modal>
    </Row>
  )
}

export default SupportChats


