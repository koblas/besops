import { Form, InputNumber, Select } from 'antd';
import { useTags } from '../../../hooks/useTags';

export function GroupFields() {
  const { data: tags = [] } = useTags();

  const tagOptions = tags.map(t => ({ value: t.id, label: t.name }));

  return (
    <>
      <Form.Item
        name="groupTagIds"
        label="Member Tags"
        extra="Any monitor with at least one of these tags is a member of this group."
      >
        <Select
          mode="multiple"
          allowClear
          placeholder="Select tags..."
          options={tagOptions}
          optionFilterProp="label"
        />
      </Form.Item>

      <Form.Item
        name="interval"
        label="Status Check Interval (seconds)"
        extra="How often to re-evaluate member status."
      >
        <InputNumber min={20} style={{ width: '100%' }} />
      </Form.Item>
    </>
  );
}
