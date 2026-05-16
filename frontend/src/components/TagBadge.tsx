import { Tag } from 'antd';

interface TagBadgeProps {
  name: string;
  color: string;
  value?: string;
  onClose?: () => void;
}

export function TagBadge({ name, color, value, onClose }: TagBadgeProps) {
  const label = value ? `${name}: ${value}` : name;
  return (
    <Tag color={color} closable={!!onClose} onClose={onClose}>
      {label}
    </Tag>
  );
}
