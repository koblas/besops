import { useState } from 'react';
import { Select, Space, Input, Button, ColorPicker, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useTags, useCreateTag } from '../hooks/useTags';
import type { Tag } from '../hooks/useTags';
import { TagBadge } from './TagBadge';

export interface TagSelectorProps {
  value: string[];
  onChange: (tagIds: string[]) => void;
}

export function TagSelector({ value, onChange }: TagSelectorProps) {
  const { data: allTags = [] } = useTags();
  const createTag = useCreateTag();
  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState('');
  const [newColor, setNewColor] = useState('#597ef7');

  const assignedSet = new Set(value);
  const available = allTags.filter(t => !assignedSet.has(t.id));
  const assigned: Tag[] = value
    .map(id => allTags.find(t => t.id === id))
    .filter((t): t is Tag => t != null);

  function handleSelect(tagId: string) {
    onChange([...value, tagId]);
  }

  function handleRemove(tagId: string) {
    onChange(value.filter(id => id !== tagId));
  }

  function handleCreate() {
    const name = newName.trim();
    if (!name) return;
    createTag.mutate(
      { name, color: newColor },
      {
        onSuccess: (tag) => {
          onChange([...value, tag.id]);
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
        {assigned.map(t => (
          <TagBadge
            key={t.id}
            name={t.name}
            color={t.color}
            onClose={() => handleRemove(t.id)}
          />
        ))}
        {assigned.length === 0 && (
          <span style={{ color: 'var(--ant-color-text-tertiary)', fontSize: 12 }}>No tags assigned</span>
        )}
      </div>

      {!creating ? (
        <Space.Compact style={{ width: '100%' }}>
          <Select
            value={undefined}
            onChange={(val) => { if (val) handleSelect(val); }}
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
