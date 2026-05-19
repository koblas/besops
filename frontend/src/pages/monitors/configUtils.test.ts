import { describe, it, expect } from 'vitest';
import { buildConfigFromForm, flattenConfigToForm } from './configUtils';

describe('buildConfigFromForm', () => {
  it('builds HttpMonitorConfig from flat form values', () => {
    const values = {
      name: 'Test HTTP',
      type: 'http',
      active: true,
      interval: 60,
      url: 'https://example.com',
      method: 'POST',
      headers: [{ name: 'X-Custom', value: 'val' }],
      body: '{"ok":true}',
      basicAuthUser: 'admin',
      basicAuthPass: 'secret',
      maxRedirects: 5,
      acceptedStatusCodes: ['200', '201'],
      ignoreTls: true,
      keyword: 'success',
      invertKeyword: true,
      jsonPath: '$.status',
      expectedValue: 'ok',
    };

    const config = buildConfigFromForm(values);

    expect(config).toEqual({
      kind: 'http',
      url: 'https://example.com',
      method: 'POST',
      headers: [{ name: 'X-Custom', value: 'val' }],
      body: '{"ok":true}',
      basicAuthUser: 'admin',
      basicAuthPass: 'secret',
      maxRedirects: 5,
      acceptedStatusCodes: ['200', '201'],
      ignoreTls: true,
      keyword: 'success',
      invertKeyword: true,
      jsonPath: '$.status',
      expectedValue: 'ok',
    });
  });

  it('builds PortMonitorConfig from flat form values', () => {
    const values = {
      name: 'Test Port',
      type: 'port',
      hostname: 'db.local',
      port: 5432,
      ignoreTls: true,
    };

    const config = buildConfigFromForm(values);

    expect(config).toEqual({
      kind: 'port',
      hostname: 'db.local',
      port: 5432,
      ignoreTls: true,
    });
  });

  it('builds PingMonitorConfig from flat form values', () => {
    const values = {
      type: 'ping',
      hostname: 'router.local',
      packetSize: 128,
    };

    const config = buildConfigFromForm(values);

    expect(config).toEqual({
      kind: 'ping',
      hostname: 'router.local',
      packetSize: 128,
    });
  });

  it('builds DnsMonitorConfig from flat form values', () => {
    const values = {
      type: 'dns',
      hostname: 'example.com',
      port: 53,
      dnsResolveType: 'A',
      dnsResolveServer: '8.8.8.8',
    };

    const config = buildConfigFromForm(values);

    expect(config).toEqual({
      kind: 'dns',
      hostname: 'example.com',
      port: 53,
      dnsResolveType: 'A',
      dnsResolveServer: '8.8.8.8',
    });
  });

  it('builds GrpcMonitorConfig from flat form values', () => {
    const values = {
      type: 'grpc-keyword',
      grpcUrl: 'grpc.example.com:443',
      grpcServiceName: 'health.v1.Health',
      grpcMethod: 'Check',
      grpcEnableTls: true,
      ignoreTls: false,
    };

    const config = buildConfigFromForm(values);

    expect(config).toEqual({
      kind: 'grpc-keyword',
      grpcUrl: 'grpc.example.com:443',
      grpcServiceName: 'health.v1.Health',
      grpcMethod: 'Check',
      grpcEnableTls: true,
      ignoreTls: false,
    });
  });

  it('builds MqttMonitorConfig from flat form values', () => {
    const values = {
      type: 'mqtt',
      hostname: 'broker.local',
      port: 1883,
      mqttTopic: 'health/check',
      mqttSuccessMessage: 'alive',
      mqttUsername: 'user',
      mqttPassword: 'pass',
      ignoreTls: false,
    };

    const config = buildConfigFromForm(values);

    expect(config).toEqual({
      kind: 'mqtt',
      hostname: 'broker.local',
      port: 1883,
      mqttTopic: 'health/check',
      mqttSuccessMessage: 'alive',
      mqttUsername: 'user',
      mqttPassword: 'pass',
      ignoreTls: false,
    });
  });

  it('builds RedisMonitorConfig from flat form values', () => {
    const values = {
      type: 'redis',
      hostname: 'redis.local',
      port: 6379,
      databaseQuery: 'PING',
    };

    const config = buildConfigFromForm(values);

    expect(config).toEqual({
      kind: 'redis',
      hostname: 'redis.local',
      port: 6379,
      databaseQuery: 'PING',
    });
  });

  it('builds SmtpMonitorConfig from flat form values', () => {
    const values = {
      type: 'smtp',
      hostname: 'mail.example.com',
      port: 587,
      ignoreTls: false,
    };

    const config = buildConfigFromForm(values);

    expect(config).toEqual({
      kind: 'smtp',
      hostname: 'mail.example.com',
      port: 587,
      ignoreTls: false,
    });
  });

  it('builds TailscalePingMonitorConfig from flat form values', () => {
    const values = {
      type: 'tailscale-ping',
      hostname: 'node.ts.net',
    };

    const config = buildConfigFromForm(values);

    expect(config).toEqual({
      kind: 'tailscale-ping',
      hostname: 'node.ts.net',
    });
  });

  it('group with tagIds produces correct config', () => {
    const values = {
      type: 'group',
      groupTagIds: ['a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'b2c3d4e5-f6a7-8901-bcde-f12345678901'],
    };

    const config = buildConfigFromForm(values);

    expect(config).toEqual({
      kind: 'group',
      tagIds: ['a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'b2c3d4e5-f6a7-8901-bcde-f12345678901'],
    });
  });

  it('group without tagIds produces minimal config', () => {
    const values = { type: 'group' };

    const config = buildConfigFromForm(values);

    expect(config).toEqual({ kind: 'group' });
  });

  it('omits undefined optional fields for http', () => {
    const values = {
      type: 'http',
      url: 'https://example.com',
      method: 'GET',
    };

    const config = buildConfigFromForm(values);

    expect(config.kind).toBe('http');
    expect(config).not.toHaveProperty('body');
    expect(config).not.toHaveProperty('keyword');
  });
});

describe('flattenConfigToForm', () => {
  it('flattens HttpMonitorConfig into form fields', () => {
    const monitor = {
      name: 'My HTTP',
      type: 'http',
      config: {
        kind: 'http' as const,
        url: 'https://example.com',
        method: 'POST' as const,
        headers: [{ name: 'Auth', value: 'Bearer x' }],
        body: '{}',
        maxRedirects: 5,
      },
    };

    const flat = flattenConfigToForm(monitor);

    expect(flat.url).toBe('https://example.com');
    expect(flat.method).toBe('POST');
    expect(flat.headers).toEqual([{ name: 'Auth', value: 'Bearer x' }]);
    expect(flat.body).toBe('{}');
    expect(flat.maxRedirects).toBe(5);
  });

  it('flattens PortMonitorConfig into form fields', () => {
    const monitor = {
      name: 'Port Mon',
      type: 'port',
      config: {
        kind: 'port' as const,
        hostname: 'db.local',
        port: 5432,
        ignoreTls: true,
      },
    };

    const flat = flattenConfigToForm(monitor);

    expect(flat.hostname).toBe('db.local');
    expect(flat.port).toBe(5432);
    expect(flat.ignoreTls).toBe(true);
  });

  it('returns empty object when config is undefined', () => {
    const monitor = { name: 'Old', type: 'http' };

    const flat = flattenConfigToForm(monitor);

    expect(flat).toEqual({});
  });

  it('flattens DnsMonitorConfig into form fields', () => {
    const monitor = {
      name: 'DNS Mon',
      type: 'dns',
      config: {
        kind: 'dns' as const,
        hostname: 'example.com',
        port: 53,
        dnsResolveType: 'A' as const,
        dnsResolveServer: '1.1.1.1',
      },
    };

    const flat = flattenConfigToForm(monitor);

    expect(flat.hostname).toBe('example.com');
    expect(flat.port).toBe(53);
    expect(flat.dnsResolveType).toBe('A');
    expect(flat.dnsResolveServer).toBe('1.1.1.1');
  });

  it('renames tagIds to groupTagIds for group config', () => {
    const monitor = {
      name: 'Group Mon',
      type: 'group',
      config: {
        kind: 'group' as const,
        tagIds: ['a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'b2c3d4e5-f6a7-8901-bcde-f12345678901'],
      },
    };

    const flat = flattenConfigToForm(monitor);

    expect(flat.groupTagIds).toEqual(['a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'b2c3d4e5-f6a7-8901-bcde-f12345678901']);
    expect(flat).not.toHaveProperty('tagIds');
  });

  it('handles group config without tagIds', () => {
    const monitor = {
      name: 'Group Mon',
      type: 'group',
      config: {
        kind: 'group' as const,
      },
    };

    const flat = flattenConfigToForm(monitor);

    expect(flat).toEqual({});
    expect(flat).not.toHaveProperty('tagIds');
    expect(flat).not.toHaveProperty('groupTagIds');
  });
});
