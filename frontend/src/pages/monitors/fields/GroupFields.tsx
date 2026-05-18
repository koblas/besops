import { Form, Select, Typography } from 'antd';
import { useTags } from '../../../hooks/useTags';

const { Text } = Typography;

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

      <Text type="secondary" style={{ fontSize: 12 }}>
        Group status is re-evaluated every 60 seconds.
      </Text>
    </>
  );
}
