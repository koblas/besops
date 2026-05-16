import { Form, Input } from 'antd';

const connectionHelp: Record<string, string> = {
  sqlserver: 'Format: Server=host;Database=name;User Id=user;Password=pass;',
  postgres: 'Format: postgresql://user:pass@host:5432/dbname',
  mysql: 'Format: mysql://user:pass@host:3306/dbname',
  mongodb: 'Format: mongodb://user:pass@host:27017/dbname',
  redis: 'Format: redis://host:6379',
};

const placeholders: Record<string, string> = {
  sqlserver: 'Server=localhost;Database=mydb;User Id=sa;Password=pass;',
  postgres: 'postgresql://user:pass@localhost:5432/mydb',
  mysql: 'mysql://user:pass@localhost:3306/mydb',
  mongodb: 'mongodb://user:pass@localhost:27017/mydb',
  redis: 'redis://localhost:6379',
};

export function DatabaseFields({ type }: { type: string }) {
  return (
    <>
      <Form.Item
        name="hostname"
        label="Connection String"
        rules={[{ required: true, message: 'Please enter a connection string' }]}
        extra={connectionHelp[type] ?? 'Enter the full connection URI for your database.'}
      >
        <Input.Password
          placeholder={placeholders[type] ?? ''}
          visibilityToggle
        />
      </Form.Item>

      <Form.Item
        name="databaseQuery"
        label="Query"
        extra="A simple query to test the connection. It must execute successfully for the monitor to report UP."
      >
        <Input.TextArea
          rows={3}
          placeholder="SELECT 1"
          style={{ fontFamily: 'monospace' }}
        />
      </Form.Item>
    </>
  );
}
