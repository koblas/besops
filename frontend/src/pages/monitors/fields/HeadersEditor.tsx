import { Button, AutoComplete, Input, Space } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { useState } from 'react';

const commonHeaders = [
  'Accept',
  'Authorization',
  'Cache-Control',
  'Cookie',
  'If-Modified-Since',
  'If-None-Match',
  'Origin',
  'Referer',
  'User-Agent',
  'X-API-Key',
  'X-Forwarded-For',
  'X-Request-ID',
];

interface HeaderEntry {
  key: string;
  value: string;
}

interface HeadersEditorProps {
  value?: string;
  onChange?: (value: string) => void;
}

function parseInitial(value: string | undefined): HeaderEntry[] {
  if (!value) return [];
  try {
    const parsed = JSON.parse(value);
    return Object.entries(parsed).map(([key, val]) => ({ key, value: String(val) }));
  } catch {
    return [];
  }
}

export function HeadersEditor({ value, onChange }: HeadersEditorProps) {
  const [entries, setEntries] = useState<HeaderEntry[]>(() => parseInitial(value));

  function emit(updated: HeaderEntry[]) {
    setEntries(updated);
    const obj: Record<string, string> = {};
    for (const entry of updated) {
      if (entry.key.trim()) {
        obj[entry.key.trim()] = entry.value;
      }
    }
    const hasEntries = Object.keys(obj).length > 0;
    onChange?.(hasEntries ? JSON.stringify(obj) : '');
  }

  function addEntry() {
    emit([...entries, { key: '', value: '' }]);
  }

  function removeEntry(index: number) {
    emit(entries.filter((_, i) => i !== index));
  }

  function updateKey(index: number, key: string) {
    const updated = [...entries];
    updated[index] = { ...updated[index], key };
    emit(updated);
  }

  function updateValue(index: number, val: string) {
    const updated = [...entries];
    updated[index] = { ...updated[index], value: val };
    emit(updated);
  }

  const usedKeys = new Set(entries.map(e => e.key.toLowerCase()));
  const suggestions = commonHeaders
    .filter(h => !usedKeys.has(h.toLowerCase()))
    .map(h => ({ value: h }));

  return (
    <div>
      {entries.map((entry, i) => (
        <Space key={i} style={{ display: 'flex', marginBottom: 8 }} align="start">
          <AutoComplete
            value={entry.key}
            onChange={(val) => updateKey(i, val)}
            options={suggestions}
            placeholder="Header name"
            style={{ width: 200 }}
            filterOption={(input, option) =>
              (option?.value as string).toLowerCase().includes(input.toLowerCase())
            }
          />
          <Input
            value={entry.value}
            onChange={(e) => updateValue(i, e.target.value)}
            placeholder="Value"
            style={{ width: 280 }}
          />
          <Button
            type="text"
            icon={<DeleteOutlined />}
            onClick={() => removeEntry(i)}
            aria-label="Remove header"
          />
        </Space>
      ))}
      <Button type="dashed" onClick={addEntry} icon={<PlusOutlined />} size="small">
        Add Header
      </Button>
    </div>
  );
}
