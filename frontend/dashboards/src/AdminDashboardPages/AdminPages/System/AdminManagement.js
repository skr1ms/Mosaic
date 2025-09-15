import React, { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Table, Modal, ModalHeader, ModalBody, ModalFooter,
    Form, FormGroup, Label, Input, Alert, Spinner
} from 'reactstrap';
import api from '../../../api/api';

const AdminManagement = () => {
    const { t } = useTranslation();
    const [admins, setAdmins] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const [currentUserRole, setCurrentUserRole] = useState('');
    
    // Модальные окна
    const [createModal, setCreateModal] = useState(false);
    const [deleteModal, setDeleteModal] = useState(false);
    const [passwordModal, setPasswordModal] = useState(false);
    const [emailModal, setEmailModal] = useState(false);
    
    // Данные для форм
    const [createForm, setCreateForm] = useState({ login: '', email: '', password: '' });
    const [passwordForm, setPasswordForm] = useState({ password: '' });
    const [emailForm, setEmailForm] = useState({ email: '' });
    const [selectedAdmin, setSelectedAdmin] = useState(null);
    
    // Загрузка списка администраторов
    const fetchAdmins = useCallback(async () => {
        try {
            setLoading(true);
            setError('');
            const response = await api.get('/admin/main-admin/admins');
            setAdmins(response.data);
        } catch (err) {
            console.error('Failed to fetch admins:', err);
            setError(t('admin.failed_to_load_admins'));
        } finally {
            setLoading(false);
        }
    }, [t]);
    
    
    const fetchCurrentUserRole = async () => {
        try {
            
            const roleFromStorage = localStorage.getItem('userRole');
            
            if (roleFromStorage) {
                setCurrentUserRole(roleFromStorage);
                return;
            }
            
            
            const response = await api.get('/admin/dashboard');
            
            if (response.data && response.data.current_user) {
                const roleFromAPI = response.data.current_user.role || 'admin';
                setCurrentUserRole(roleFromAPI);
            } else {
                
                setCurrentUserRole('admin');
            }
        } catch (err) {
            console.error('Failed to fetch current user role:', err);
            
            const roleFromStorage = localStorage.getItem('userRole');
            setCurrentUserRole(roleFromStorage || 'admin');
        }
    };

    useEffect(() => {
        fetchAdmins();
        fetchCurrentUserRole();
    }, [fetchAdmins]);
    
    
    const handleCreateAdmin = async () => {
        try {
            setError('');
            setSuccess('');
            
            // Проверяем права доступа
            if (currentUserRole !== 'main_admin') {
                setError(t('admin.only_main_admin_create'));
                return;
            }
            
            if (!createForm.login || !createForm.email || !createForm.password) {
                setError(t('forms.required_field'));
                return;
            }
            
            if (createForm.password.length < 6) {
                setError(t('admin.password_min_length'));
                return;
            }
            
            await api.post('/admin/main-admin/admins', createForm);
            setSuccess(t('admin.admin_created_success'));
            setCreateModal(false);
            setCreateForm({ login: '', email: '', password: '' });
            fetchAdmins();
        } catch (err) {
            console.error('Failed to create admin:', err);
            setError(err.response?.data?.error || t('admin.failed_to_create_admin'));
        }
    };
    
    
    const handleDeleteAdmin = async () => {
        try {
            setError('');
            setSuccess('');
            
            // Проверяем права доступа
            if (currentUserRole !== 'main_admin') {
                setError(t('admin.only_main_admin_delete'));
                return;
            }
            
            await api.delete(`/admin/main-admin/admins/${selectedAdmin.id}`);
            setSuccess(t('admin.admin_deleted_success'));
            setDeleteModal(false);
            setSelectedAdmin(null);
            fetchAdmins();
        } catch (err) {
            console.error('Failed to delete admin:', err);
            setError(err.response?.data?.error || t('admin.failed_to_delete_admin'));
        }
    };
    
    
    const handleUpdatePassword = async () => {
        try {
            setError('');
            setSuccess('');
            
            if (!passwordForm.password || passwordForm.password.length < 6) {
                setError(t('admin.password_min_length'));
                return;
            }
            
            await api.patch(`/admin/admins/${selectedAdmin.id}/password`, passwordForm);
            setSuccess(t('admin.password_updated_success'));
            setPasswordModal(false);
            setSelectedAdmin(null);
            setPasswordForm({ password: '' });
        } catch (err) {
            console.error('Failed to update password:', err);
            setError(err.response?.data?.error || t('admin.failed_to_update_password'));
        }
    };
    
    
    const handleUpdateEmail = async () => {
        try {
            setError('');
            setSuccess('');
            
            if (!emailForm.email) {
                setError(t('admin.email_required'));
                return;
            }
            
            await api.patch(`/admin/admins/${selectedAdmin.id}/email`, emailForm);
            setSuccess(t('admin.email_updated_success'));
            setEmailModal(false);
            setSelectedAdmin(null);
            setEmailForm({ email: '' });
            fetchAdmins();
        } catch (err) {
            console.error('Failed to update email:', err);
            setError(err.response?.data?.error || t('admin.failed_to_update_email'));
        }
    };
    
    
    const openDeleteModal = (admin) => {
        if (currentUserRole !== 'main_admin') {
            setError(t('admin.only_main_admin_delete'));
            return;
        }
        setSelectedAdmin(admin);
        setDeleteModal(true);
    };
    
    const openPasswordModal = (admin) => {
        setSelectedAdmin(admin);
        setPasswordModal(true);
    };
    
    const openEmailModal = (admin) => {
        setSelectedAdmin(admin);
        setEmailForm({ email: admin.email });
        setEmailModal(true);
    };
    
    
    const clearMessages = () => {
        setError('');
        setSuccess('');
    };
    
    if (loading) {
        return (
            <div className="d-flex justify-content-center align-items-center min-vh-100">
                <Spinner color="primary" />
            </div>
        );
    }

    
    return (
        <div className="admin-management">
            <Row>
                <Col>
                    <Card>
                        <CardHeader className="d-flex justify-content-between align-items-center">
                            <CardTitle tag="h5">
                                {t('admin.management')}
                                {currentUserRole === 'main_admin' && (
                                    <span className="ms-2 badge bg-danger">{t('admin.main_admin')}</span>
                                )}
                            </CardTitle>
                            {currentUserRole === 'main_admin' && (
                                <Button 
                                    color="primary" 
                                    onClick={() => {
                                        if (currentUserRole === 'main_admin') {
                                            setCreateModal(true);
                                        } else {
                                            setError(t('admin.only_main_admin_create'));
                                        }
                                    }}
                                >
                                    {t('admin.create_administrator')}
                                </Button>
                            )}

                        </CardHeader>
                        <CardBody>
                            {error && (
                                <Alert color="danger" onDismiss={clearMessages}>
                                    {error}
                                </Alert>
                            )}
                            {success && (
                                <Alert color="success" onDismiss={clearMessages}>
                                    {success}
                                </Alert>
                            )}
                            
                            {currentUserRole !== 'main_admin' && (
                                <Alert color="info" onDismiss={clearMessages}>
                                    <strong>{t('admin.info_permissions')}:</strong> {t('admin.can_change_own_data')}
                                    <br />
                                    <small>{t('admin.your_role')}: {currentUserRole === 'admin' ? t('admin.regular_admin') : currentUserRole}</small>
                                </Alert>
                            )}
                            
                            <Table responsive>
                                <thead>
                                    <tr>
                                        <th>{t('admin.login')}</th>
                                        <th>{t('common.email')}</th>
                                        <th>{t('admin.role')}</th>
                                        <th>{t('admin.last_login')}</th>
                                        <th>{t('admin.created_at')}</th>
                                        <th>{t('common.actions')}</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {admins.map((admin) => (
                                        <tr key={admin.id}>
                                            <td>{admin.login}</td>
                                            <td>{admin.email}</td>
                                            <td>
                                                <span className={`badge ${admin.role === 'main_admin' ? 'bg-danger' : 'bg-primary'}`}>
                                                    {admin.role === 'main_admin' ? t('admin.main_admin') : t('admin.regular_admin')}
                                                </span>
                                            </td>
                                            <td>
                                                {admin.last_login 
                                                    ? new Date(admin.last_login).toLocaleString('ru-RU')
                                                    : t('admin.never')
                                                }
                                            </td>
                                            <td>
                                                {new Date(admin.created_at).toLocaleString('ru-RU')}
                                            </td>
                                            <td>
                                                <Button
                                                    color="warning"
                                                    size="sm"
                                                    className="me-2"
                                                    onClick={() => openPasswordModal(admin)}
                                                >
                                                    {t('admin.change_password')}
                                                </Button>
                                                <Button
                                                    color="info"
                                                    size="sm"
                                                    className="me-2"
                                                    onClick={() => openEmailModal(admin)}
                                                >
                                                    {t('admin.change_email')}
                                                </Button>
                                                {currentUserRole === 'main_admin' && admin.role !== 'main_admin' && (
                                                    <Button
                                                        color="danger"
                                                        size="sm"
                                                        onClick={() => openDeleteModal(admin)}
                                                    >
                                                        {t('admin.delete_admin')}
                                                    </Button>
                                                )}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </Table>
                        </CardBody>
                    </Card>
                </Col>
            </Row>
            
            {}
            <Modal isOpen={createModal} toggle={() => setCreateModal(false)}>
                <ModalHeader toggle={() => setCreateModal(false)}>
                    {t('admin.create_administrator')}
                </ModalHeader>
                <ModalBody>
                    <Form>
                        <FormGroup>
                            <Label for="login">{t('admin.login')}</Label>
                            <Input
                                id="login"
                                type="text"
                                value={createForm.login}
                                onChange={(e) => setCreateForm({...createForm, login: e.target.value})}
                                placeholder={t('user.enter_login')}
                            />
                        </FormGroup>
                        <FormGroup>
                            <Label for="email">{t('common.email')}</Label>
                            <Input
                                id="email"
                                type="email"
                                value={createForm.email}
                                onChange={(e) => setCreateForm({...createForm, email: e.target.value})}
                                placeholder={t('user.enter_email')}
                            />
                        </FormGroup>
                        <FormGroup>
                            <Label for="password">{t('user.new_password')}</Label>
                            <Input
                                id="password"
                                type="password"
                                value={createForm.password}
                                onChange={(e) => setCreateForm({...createForm, password: e.target.value})}
                                placeholder={t('user.enter_password')}
                            />
                        </FormGroup>
                    </Form>
                </ModalBody>
                <ModalFooter>
                    <Button color="primary" onClick={handleCreateAdmin}>
                        {t('common.create')}
                    </Button>
                    <Button color="secondary" onClick={() => setCreateModal(false)}>
                        {t('common.cancel')}
                    </Button>
                </ModalFooter>
            </Modal>
            
            {}
            <Modal isOpen={deleteModal} toggle={() => setDeleteModal(false)}>
                <ModalHeader toggle={() => setDeleteModal(false)}>
                    {t('admin.confirm_delete')}
                </ModalHeader>
                <ModalBody>
                    {t('admin.delete_admin_warning')} <strong>{selectedAdmin?.login}</strong>?
                    <br />
                    {t('admin.delete_irreversible')}
                </ModalBody>
                <ModalFooter>
                    <Button color="danger" onClick={handleDeleteAdmin}>
                        {t('common.delete')}
                    </Button>
                    <Button color="secondary" onClick={() => setDeleteModal(false)}>
                        {t('common.cancel')}
                    </Button>
                </ModalFooter>
            </Modal>
            
            {}
            <Modal isOpen={passwordModal} toggle={() => setPasswordModal(false)}>
                <ModalHeader toggle={() => setPasswordModal(false)}>
                    {t('admin.update_password')} {selectedAdmin?.login}
                </ModalHeader>
                <ModalBody>
                    <Form>
                        <FormGroup>
                            <Label for="newPassword">{t('admin.new_password')}</Label>
                            <Input
                                id="newPassword"
                                type="password"
                                value={passwordForm.password}
                                onChange={(e) => setPasswordForm({password: e.target.value})}
                                placeholder={t('user.enter_password')}
                            />
                        </FormGroup>
                    </Form>
                </ModalBody>
                <ModalFooter>
                    <Button color="primary" onClick={handleUpdatePassword}>
                        {t('common.update')}
                    </Button>
                    <Button color="secondary" onClick={() => setPasswordModal(false)}>
                        {t('common.cancel')}
                    </Button>
                </ModalFooter>
            </Modal>
            
            {}
            <Modal isOpen={emailModal} toggle={() => setEmailModal(false)}>
                <ModalHeader toggle={() => setEmailModal(false)}>
                    {t('admin.update_email')} {selectedAdmin?.login}
                </ModalHeader>
                <ModalBody>
                    <Form>
                        <FormGroup>
                            <Label for="newEmail">{t('admin.new_email')}</Label>
                            <Input
                                id="newEmail"
                                type="email"
                                value={emailForm.email}
                                onChange={(e) => setEmailForm({email: e.target.value})}
                                placeholder={t('user.enter_email')}
                            />
                        </FormGroup>
                    </Form>
                </ModalBody>
                <ModalFooter>
                    <Button color="primary" onClick={handleUpdateEmail}>
                        {t('common.update')}
                    </Button>
                    <Button color="secondary" onClick={() => setEmailModal(false)}>
                        {t('common.cancel')}
                    </Button>
                </ModalFooter>
            </Modal>
        </div>
    );
};

export default AdminManagement;
