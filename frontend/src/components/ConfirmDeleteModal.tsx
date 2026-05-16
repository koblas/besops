import { Modal } from 'antd';

interface ConfirmDeleteModalProps {
  open: boolean;
  title?: string;
  description?: string;
  onConfirm: () => void;
  onCancel: () => void;
  loading?: boolean;
}

export function ConfirmDeleteModal({
  open,
  title = 'Confirm Delete',
  description = 'Are you sure you want to delete this item? This action cannot be undone.',
  onConfirm,
  onCancel,
  loading,
}: ConfirmDeleteModalProps) {
  return (
    <Modal
      open={open}
      title={title}
      onOk={onConfirm}
      onCancel={onCancel}
      okText="Delete"
      okButtonProps={{ danger: true, loading }}
      cancelButtonProps={{ disabled: loading }}
    >
      <p>{description}</p>
    </Modal>
  );
}
