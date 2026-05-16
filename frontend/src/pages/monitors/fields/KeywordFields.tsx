import { Form, Input, Switch } from 'antd';

export function KeywordFields() {
  return (
    <>
      <Form.Item
        name="keyword"
        label="Keyword"
        rules={[{ required: true, message: 'Please enter a keyword to search for' }]}
        extra="Case-sensitive text to look for in the response body."
      >
        <Input placeholder="alive" />
      </Form.Item>

      <Form.Item
        name="invertKeyword"
        label="Alert When Found"
        valuePropName="checked"
        extra="By default, the monitor alerts when the keyword is MISSING. Enable this to alert when the keyword IS found instead."
      >
        <Switch />
      </Form.Item>
    </>
  );
}
