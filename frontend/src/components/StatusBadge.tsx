import { Tag } from 'antd';
import { STATUS_COLORS, STATUS_LABELS, type StatusValue } from '../lib/constants';

interface StatusBadgeProps {
  status: StatusValue;
}

export function StatusBadge({ status }: StatusBadgeProps) {
  return <Tag color={STATUS_COLORS[status]}>{STATUS_LABELS[status]}</Tag>;
}
