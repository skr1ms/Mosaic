import React from 'react';
import { Modal, ModalHeader, ModalBody, ModalFooter, Button } from 'reactstrap';

const ConfirmModal = ({ isOpen, title, message, confirmText = 'Подтвердить', cancelText = 'Отмена', onConfirm, onCancel, confirmColor = 'danger', containerSelector }) => {
  return (
    <Modal
      isOpen={isOpen}
      toggle={onCancel}
      backdrop
      container={containerSelector ? document.querySelector(containerSelector) : undefined}
      modalTransition={{ timeout: 80 }}
      backdropTransition={{ timeout: 80 }}
    >
      <ModalHeader toggle={onCancel}>{title}</ModalHeader>
      <ModalBody>
        <p className="mb-0">{message}</p>
      </ModalBody>
      <ModalFooter>
        <Button color="secondary" onClick={onCancel}>{cancelText}</Button>
        <Button color={confirmColor} onClick={onConfirm}>{confirmText}</Button>
      </ModalFooter>
    </Modal>
  );
};

export default ConfirmModal;


