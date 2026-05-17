import { useState } from 'react';
import { Select, Space, Input, Button, ColorPicker, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useTags, useCreateTag, useAddMonitorTag, useRemoveMonitorTag } from '../hooks/useTags';
import type { MonitorTag } from '../hooks/useTags';
import { TagBadge } from './TagBadge';

interface TagSelectorProps {
  monitorId: string;
  assignedTags: MonitorTag[];
}

export function TagSelector({ monitorId, assignedTags }: TagSelectorProps) {
  const { data: allTags = [] } = useTags();
  const createTag = useCreateTag();
  const addTag = useAddMonitorTag();
  const removeTag = useRemoveMonitorTag();
  const [selectedTagId, setSelectedTagId] = useState<string | undefined>();
  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState('');
  const [newColor, setNewColor] = useState('#597ef7');

  const assignedIds = new Set(assignedTags.map(t => t.tagId));
  const available = allTags.filter(t => !assignedIds.has(t.id));

  function handleAdd(tagId: string) {
    addTag.mutate(
      { monitorId, tagId },
      {
        onSuccess: () => setSelectedTagId(undefined),
        onError: () => message.error('Failed to add tag'),
      },
    );
  }

  function handleRemove(tagId: string) {
    removeTag.mutate(
      { monitorId, tagId },
      { onError: () => message.error('Failed to remove tag') },
    );
  }

  function handleCreate() {
    const name = newName.trim();
    if (!name) return;
    createTag.mutate(
      { name, color: newColor },
      {
        onSuccess: (tag) => {
          addTag.mutate({ monitorId, tagId: tag.id });
          setNewName('');
          setCreating(false);
        },
        onError: () => message.error('Failed to create tag'),
      },
    );
  }

  return (
    <div>
      <div style={{ marginBottom: 8, display: 'flex', flexWrap: 'wrap', gap: 4 }}>
        {assignedTags.map(t => (
          <TagBadge
            key={t.tagId}
            name={t.name ?? ''}
            color={t.color ?? ''}
            onClose={() => handleRemove(t.tagId!)}
          />
        ))}
        {assignedTags.length === 0 && (
          <span style={{ color: 'var(--ant-color-text-tertiary)', fontSize: 12 }}>No tags assigned</span>
        )}
      </div>

      {!creating ? (
        <Space.Compact style={{ width: '100%' }}>
          <Select
            value={selectedTagId}
            onChange={(val) => { setSelectedTagId(val); if (val) handleAdd(val); }}
            placeholder="Select a tag to add"
            options={available.map(t => ({ value: t.id, label: t.name }))}
            style={{ flex: 1, minWidth: 160 }}
            allowClear
            showSearch
            optionFilterProp="label"
          />
          <Button icon={<PlusOutlined />} onClick={() => setCreating(true)}>
            New Tag
          </Button>
        </Space.Compact>
      ) : (
        <Space.Compact style={{ width: '100%' }}>
          <ColorPicker
            value={newColor}
            onChange={(_, hex) => setNewColor(hex)}
            size="middle"
          />
          <Input
            placeholder="Tag name"
            value={newName}
            onChange={e => setNewName(e.target.value)}
            onPressEnter={handleCreate}
            style={{ flex: 1 }}
            autoFocus
          />
          <Button type="primary" onClick={handleCreate} loading={createTag.isPending} disabled={!newName.trim()}>
            Add
          </Button>
          <Button onClick={() => { setCreating(false); setNewName(''); }}>
            Cancel
          </Button>
        </Space.Compact>
      )}
    </div>
  );
}
