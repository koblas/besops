import { Button, AutoComplete, Input, Space } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { useState } from 'react';

const commonHeaders = [
  'Accept',
  'Authorization',
  'Cache-Control',
  'Content-Type',
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

export interface HeaderEntry {
  name: string;
  value: string;
}

interface HeadersEditorProps {
  value?: HeaderEntry[];
  onChange?: (value: HeaderEntry[]) => void;
}

export function HeadersEditor({ value = [], onChange }: HeadersEditorProps) {
  const [entries, setEntries] = useState<HeaderEntry[]>(value);

  function emit(updated: HeaderEntry[]) {
    setEntries(updated);
    onChange?.(updated.filter(e => e.name.trim()));
  }

  function addEntry() {
    emit([...entries, { name: '', value: '' }]);
  }

  function removeEntry(index: number) {
    emit(entries.filter((_, i) => i !== index));
  }

  function updateName(index: number, name: string) {
    const updated = [...entries];
    updated[index] = { ...updated[index], name };
    emit(updated);
  }

  function updateValue(index: number, val: string) {
    const updated = [...entries];
    updated[index] = { ...updated[index], value: val };
    emit(updated);
  }

  const usedNames = new Set(entries.map(e => e.name.toLowerCase()));
  const suggestions = commonHeaders
    .filter(h => !usedNames.has(h.toLowerCase()))
    .map(h => ({ value: h }));

  return (
    <div>
      {entries.map((entry, i) => (
        <Space key={i} style={{ display: 'flex', marginBottom: 8 }} align="start">
          <AutoComplete
            value={entry.name}
            onChange={(val) => updateName(i, val)}
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
