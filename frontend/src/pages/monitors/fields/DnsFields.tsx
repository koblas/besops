import { Form, Input, Select } from 'antd';

const dnsTypes = ['A', 'AAAA', 'CAA', 'CNAME', 'MX', 'NS', 'PTR', 'SOA', 'SRV', 'TXT'];

export function DnsFields() {
  return (
    <>
      <Form.Item
        name="hostname"
        label="Hostname to Resolve"
        rules={[{ required: true, message: 'Please enter a hostname' }]}
        extra="The domain name to look up."
      >
        <Input placeholder="example.com" />
      </Form.Item>

      <Form.Item name="dnsResolveType" label="Record Type" initialValue="A" extra="The DNS record type to query.">
        <Select options={dnsTypes.map(t => ({ value: t, label: t }))} />
      </Form.Item>

      <Form.Item
        name="dnsResolveServer"
        label="DNS Server"
        extra="Leave empty to use your system's default resolver."
      >
        <Input placeholder="8.8.8.8" />
      </Form.Item>
    </>
  );
}
