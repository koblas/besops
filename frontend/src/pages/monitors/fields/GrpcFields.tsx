import { Form, Input, Switch } from 'antd';

export function GrpcFields() {
  return (
    <>
      <Form.Item
        name="grpcUrl"
        label="gRPC Endpoint"
        rules={[{ required: true, message: 'Please enter a gRPC endpoint' }]}
        extra="Host and port without a scheme (e.g. localhost:50051)."
      >
        <Input placeholder="localhost:50051" />
      </Form.Item>

      <Form.Item
        name="grpcServiceName"
        label="Service Name"
        extra="The fully-qualified service name to health-check. Leave empty to check the server overall."
      >
        <Input placeholder="grpc.health.v1.Health" />
      </Form.Item>

      <Form.Item
        name="grpcMethod"
        label="Method"
        extra="The RPC method to call for the health check."
      >
        <Input placeholder="Check" />
      </Form.Item>

      <Form.Item name="grpcEnableTls" label="Enable TLS" valuePropName="checked" extra="Connect using TLS encryption.">
        <Switch />
      </Form.Item>
    </>
  );
}
