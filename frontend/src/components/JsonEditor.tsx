import { Input, Typography } from 'antd';
import { useState } from 'react';

const { Text } = Typography;

interface JsonEditorProps {
  value?: string;
  onChange?: (value: string) => void;
  rows?: number;
  placeholder?: string;
}

export function JsonEditor({ value = '', onChange, rows = 6, placeholder }: JsonEditorProps) {
  const [error, setError] = useState<string | null>(null);

  function handleChange(e: React.ChangeEvent<HTMLTextAreaElement>) {
    const val = e.target.value;
    onChange?.(val);
    if (!val.trim()) {
      setError(null);
      return;
    }
    try {
      JSON.parse(val);
      setError(null);
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <div>
      <Input.TextArea
        value={value}
        onChange={handleChange}
        rows={rows}
        placeholder={placeholder}
        status={error ? 'error' : undefined}
        style={{ fontFamily: 'monospace' }}
      />
      {error && (
        <Text type="danger" style={{ fontSize: 12 }}>
          {error}
        </Text>
      )}
    </div>
  );
}
