import { Form, Input, Select, Switch, InputNumber, Collapse, Typography, AutoComplete } from 'antd';
import { useProxies } from '../../../hooks/useProxies';
import { HeadersEditor } from './HeadersEditor';

const { Text } = Typography;

const httpMethods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'];
const methodsWithBody = ['POST', 'PUT', 'PATCH'];

const contentTypes = [
  { value: 'application/json', label: 'application/json' },
  { value: 'application/x-www-form-urlencoded', label: 'application/x-www-form-urlencoded' },
  { value: 'text/plain', label: 'text/plain' },
  { value: 'text/xml', label: 'text/xml' },
  { value: 'application/xml', label: 'application/xml' },
  { value: 'multipart/form-data', label: 'multipart/form-data' },
  { value: 'application/octet-stream', label: 'application/octet-stream' },
];

const bodyPlaceholders: Record<string, string> = {
  'application/json': '{"key": "value"}',
  'application/x-www-form-urlencoded': 'field1=value1&field2=value2',
  'text/plain': 'Your text here...',
  'text/xml': '<root>\n  <element>value</element>\n</root>',
  'application/xml': '<root>\n  <element>value</element>\n</root>',
};

function isJsonContentType(ct: string | undefined): boolean {
  if (!ct) return false;
  return ct === 'application/json' || ct.endsWith('+json');
}

export function HttpFields() {
  const { data: proxies = [] } = useProxies();
  const form = Form.useFormInstance();
  const method = Form.useWatch('method', form);
  const contentType = Form.useWatch('bodyContentType', form);
  const showBody = methodsWithBody.includes(method);
  const expectsJson = isJsonContentType(contentType);

  return (
    <>
      <Form.Item
        name="url"
        label="URL"
        rules={[
          { required: true, message: 'Please enter a URL' },
          { type: 'url', message: 'Enter a valid URL (e.g. https://example.com)' },
        ]}
        extra="The full URL to monitor, including https://"
      >
        <Input placeholder="https://example.com/health" />
      </Form.Item>

      <Form.Item name="method" label="HTTP Method" initialValue="GET">
        <Select options={httpMethods.map(m => ({ value: m, label: m }))} />
      </Form.Item>

      {showBody && (
        <>
          <Form.Item
            name="bodyContentType"
            label="Content Type"
            initialValue="application/json"
            extra="Choose a common type or type your own MIME type (e.g. application/vnd.api+json)."
          >
            <AutoComplete
              options={contentTypes}
              placeholder="application/json"
              filterOption={(input, option) =>
                (option?.value as string).toLowerCase().includes(input.toLowerCase())
              }
            />
          </Form.Item>

          <Form.Item
            name="body"
            label="Request Body"
            validateTrigger="onBlur"
            rules={expectsJson ? [{
              validator(_, val) {
                if (!val || !val.trim()) return Promise.resolve();
                try {
                  JSON.parse(val);
                  return Promise.resolve();
                } catch (e) {
                  const msg = e instanceof SyntaxError ? e.message : 'Invalid JSON';
                  return Promise.reject(new Error(`Invalid JSON — ${msg}`));
                }
              },
            }] : undefined}
          >
            <Input.TextArea
              rows={4}
              placeholder={bodyPlaceholders[contentType] || bodyPlaceholders['application/json']}
              style={{ fontFamily: 'monospace' }}
            />
          </Form.Item>
        </>
      )}

      <Form.Item
        name="acceptedStatusCodes"
        label="Accepted Status Codes"
        extra="HTTP codes that indicate success. Use ranges like 200-299."
        initialValue={['200-299']}
      >
        <Select
          mode="tags"
          placeholder="200-299"
          tokenSeparators={[',']}
        />
      </Form.Item>

      <Form.Item name="maxRedirects" label="Max Redirects" initialValue={10} extra="How many HTTP redirects to follow before failing.">
        <InputNumber min={0} max={50} style={{ width: '100%' }} />
      </Form.Item>

      <Collapse
        ghost
        items={[{
          key: 'auth',
          label: 'Authentication & Headers',
          children: (
            <>
              <Form.Item name="basicAuthUser" label="Username">
                <Input placeholder="Leave empty if no auth needed" autoComplete="off" />
              </Form.Item>

              <Form.Item name="basicAuthPass" label="Password">
                <Input.Password placeholder="Leave empty if no auth needed" autoComplete="new-password" />
              </Form.Item>

              <Form.Item
                name="headers"
                label="Custom Headers"
                extra="Content-Type is set automatically when a request body is configured above."
              >
                <HeadersEditor />
              </Form.Item>
            </>
          ),
        }]}
      />

      <Collapse
        ghost
        items={[{
          key: 'keyword',
          label: 'Keyword Check',
          children: (
            <>
              <Form.Item
                name="keyword"
                label="Keyword"
                extra="Case-sensitive text to look for in the response body. Leave empty to skip."
              >
                <Input placeholder="alive" />
              </Form.Item>

              <Form.Item
                name="invertKeyword"
                label="Alert When Found"
                valuePropName="checked"
                extra="By default, alerts when the keyword is MISSING. Enable to alert when it IS found."
              >
                <Switch />
              </Form.Item>
            </>
          ),
        }]}
      />

      <Collapse
        ghost
        items={[{
          key: 'jsonQuery',
          label: 'JSON Query',
          children: (
            <>
              <Form.Item
                name="jsonPath"
                label="JSON Path"
                extra="Dot-notation path into the JSON response (e.g. data.status or items.0.id). Leave empty to skip."
              >
                <Input placeholder="data.status" style={{ fontFamily: 'monospace' }} />
              </Form.Item>

              <Form.Item
                name="expectedValue"
                label="Expected Value"
                extra="The value the path must return. If empty, the check passes when the path exists."
              >
                <Input placeholder="ok" style={{ fontFamily: 'monospace' }} />
              </Form.Item>
            </>
          ),
        }]}
      />

      <Collapse
        ghost
        items={[{
          key: 'tls',
          label: 'TLS & Proxy',
          children: (
            <>
              <Form.Item name="ignoreTls" label="Ignore TLS/SSL errors" valuePropName="checked" extra="Enable this for self-signed certificates. Not recommended for production.">
                <Switch />
              </Form.Item>

              {proxies.length > 0 ? (
                <Form.Item name="proxyId" label="Proxy" extra="Route requests through a configured proxy.">
                  <Select
                    allowClear
                    placeholder="Direct connection (no proxy)"
                    options={proxies.map(p => ({ value: p.id, label: `${p.protocol}://${p.host}:${p.port}` }))}
                  />
                </Form.Item>
              ) : (
                <Text type="secondary" style={{ display: 'block', marginBottom: 16 }}>
                  No proxies configured. You can add proxies in Settings → Proxies.
                </Text>
              )}
            </>
          ),
        }]}
      />
    </>
  );
}
