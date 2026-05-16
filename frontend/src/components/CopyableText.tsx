import { Typography, message } from 'antd';
import { CopyOutlined } from '@ant-design/icons';

const { Text } = Typography;

interface CopyableTextProps {
  text: string;
  displayText?: string;
}

export function CopyableText({ text, displayText }: CopyableTextProps) {
  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(text);
      message.success('Copied to clipboard');
    } catch {
      message.error('Failed to copy — try selecting the text manually');
    }
  }

  return (
    <Text
      code
      role="button"
      tabIndex={0}
      aria-label={`Copy: ${displayText ?? text}`}
      style={{ cursor: 'pointer' }}
      onClick={handleCopy}
      onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); handleCopy(); } }}
    >
      {displayText ?? text} <CopyOutlined style={{ fontSize: 12 }} />
    </Text>
  );
}
