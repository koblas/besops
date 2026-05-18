import type { components } from '../../api/generated/v1';

type MonitorConfig = components['schemas']['MonitorConfig'];

type FormValues = Record<string, unknown>;

const httpFields = [
  'url', 'method', 'headers', 'body', 'basicAuthUser', 'basicAuthPass',
  'maxRedirects', 'acceptedStatusCodes', 'ignoreTls', 'keyword',
  'invertKeyword', 'jsonPath', 'expectedValue', 'proxyId',
] as const;

const portFields = ['hostname', 'port', 'ignoreTls'] as const;
const pingFields = ['hostname', 'packetSize'] as const;
const dnsFields = ['hostname', 'port', 'dnsResolveType', 'dnsResolveServer'] as const;
const grpcFields = ['grpcUrl', 'grpcServiceName', 'grpcMethod', 'grpcEnableTls', 'ignoreTls'] as const;
const mqttFields = ['hostname', 'port', 'mqttTopic', 'mqttSuccessMessage', 'mqttUsername', 'mqttPassword', 'ignoreTls'] as const;
const redisFields = ['hostname', 'port', 'databaseQuery'] as const;
const smtpFields = ['hostname', 'port', 'ignoreTls'] as const;
const tailscalePingFields = ['hostname'] as const;

function pick(values: FormValues, fields: readonly string[]): Record<string, unknown> {
  const result: Record<string, unknown> = {};
  for (const key of fields) {
    if (values[key] !== undefined) {
      result[key] = values[key];
    }
  }
  return result;
}

export function buildConfigFromForm(values: FormValues): MonitorConfig {
  const type = values.type as string;

  switch (type) {
    case 'http':
      return { kind: 'http', ...pick(values, httpFields) } as MonitorConfig;
    case 'port':
      return { kind: 'port', ...pick(values, portFields) } as MonitorConfig;
    case 'ping':
      return { kind: 'ping', ...pick(values, pingFields) } as MonitorConfig;
    case 'dns':
      return { kind: 'dns', ...pick(values, dnsFields) } as MonitorConfig;
    case 'grpc-keyword':
      return { kind: 'grpc-keyword', ...pick(values, grpcFields) } as MonitorConfig;
    case 'mqtt':
      return { kind: 'mqtt', ...pick(values, mqttFields) } as MonitorConfig;
    case 'redis':
      return { kind: 'redis', ...pick(values, redisFields) } as MonitorConfig;
    case 'push':
      return { kind: 'push' };
    case 'smtp':
      return { kind: 'smtp', ...pick(values, smtpFields) } as MonitorConfig;
    case 'tailscale-ping':
      return { kind: 'tailscale-ping', ...pick(values, tailscalePingFields) } as MonitorConfig;
    case 'group':
      return { kind: 'group' };
    default:
      return { kind: 'group' };
  }
}

export function flattenConfigToForm(monitor: FormValues): Record<string, unknown> {
  const config = monitor.config as (MonitorConfig | undefined);
  if (!config) return {};

  const { kind: _, ...rest } = config as Record<string, unknown>;
  return rest;
}
