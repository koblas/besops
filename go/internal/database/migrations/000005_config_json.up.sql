-- Consolidate type-specific monitor columns into a single config_json column.
-- The JSON format matches the OAS MonitorConfig discriminated union (keyed on "kind").

PRAGMA foreign_keys=OFF;

CREATE TABLE monitor_new (
    id TEXT PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT 1,
    user_id TEXT NOT NULL,
    interval INTEGER NOT NULL DEFAULT 60,
    type VARCHAR(30) NOT NULL,
    weight INTEGER DEFAULT 2000,
    created_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    maxretries INTEGER NOT NULL DEFAULT 0,
    upside_down BOOLEAN NOT NULL DEFAULT 0,
    retry_interval INTEGER NOT NULL DEFAULT 60,
    timeout INTEGER NOT NULL DEFAULT 48,
    description TEXT DEFAULT NULL,
    resend_interval INTEGER NOT NULL DEFAULT 0,
    expiry_notification BOOLEAN DEFAULT 0,
    config_json TEXT NOT NULL DEFAULT '{}',
    CONSTRAINT fk_monitor_user FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);

INSERT INTO monitor_new (id, name, active, user_id, interval, type, weight, created_date, maxretries, upside_down, retry_interval, timeout, description, resend_interval, expiry_notification, config_json)
SELECT
    id, name, active, user_id, interval, type, weight, created_date, maxretries, upside_down, retry_interval, timeout, description, resend_interval, expiry_notification,
    CASE type
        WHEN 'http' THEN json_object(
            'kind', 'http',
            'url', url,
            'method', COALESCE(method, 'GET'),
            'body', body,
            'headers', json(COALESCE(headers, '[]')),
            'basicAuthUser', basic_auth_user,
            'basicAuthPass', basic_auth_pass,
            'maxRedirects', maxredirects,
            'acceptedStatusCodes', json(COALESCE(accepted_statuscodes_json, '["200-299"]')),
            'ignoreTls', CASE WHEN ignore_tls THEN json('true') ELSE json('false') END,
            'keyword', keyword,
            'invertKeyword', CASE WHEN invert_keyword THEN json('true') ELSE json('false') END,
            'jsonPath', json_path,
            'expectedValue', expected_value,
            'proxyId', proxy_id
        )
        WHEN 'port' THEN json_object(
            'kind', 'port',
            'hostname', hostname,
            'port', port,
            'ignoreTls', CASE WHEN ignore_tls THEN json('true') ELSE json('false') END
        )
        WHEN 'ping' THEN json_object(
            'kind', 'ping',
            'hostname', hostname,
            'packetSize', packet_size
        )
        WHEN 'dns' THEN json_object(
            'kind', 'dns',
            'hostname', hostname,
            'port', port,
            'dnsResolveType', dns_resolve_type,
            'dnsResolveServer', dns_resolve_server
        )
        WHEN 'grpc-keyword' THEN json_object(
            'kind', 'grpc-keyword',
            'grpcUrl', grpc_url,
            'grpcServiceName', grpc_service_name,
            'grpcMethod', grpc_method,
            'grpcEnableTls', CASE WHEN grpc_enable_tls THEN json('true') ELSE json('false') END,
            'ignoreTls', CASE WHEN ignore_tls THEN json('true') ELSE json('false') END
        )
        WHEN 'mqtt' THEN json_object(
            'kind', 'mqtt',
            'hostname', hostname,
            'port', port,
            'mqttTopic', mqtt_topic,
            'mqttSuccessMessage', mqtt_success_message,
            'mqttUsername', mqtt_username,
            'mqttPassword', mqtt_password,
            'ignoreTls', CASE WHEN ignore_tls THEN json('true') ELSE json('false') END
        )
        WHEN 'redis' THEN json_object(
            'kind', 'redis',
            'hostname', hostname,
            'port', port,
            'databaseQuery', database_query
        )
        WHEN 'smtp' THEN json_object(
            'kind', 'smtp',
            'hostname', hostname,
            'port', port,
            'ignoreTls', CASE WHEN ignore_tls THEN json('true') ELSE json('false') END
        )
        WHEN 'tailscale-ping' THEN json_object(
            'kind', 'tailscale-ping',
            'hostname', hostname
        )
        WHEN 'group' THEN json_object(
            'kind', 'group',
            'tagIds', json(COALESCE(group_tag_ids_json, '[]'))
        )
        ELSE json_object('kind', type)
    END
FROM monitor;

DROP TABLE monitor;
ALTER TABLE monitor_new RENAME TO monitor;

CREATE INDEX idx_monitor_user_id ON monitor(user_id);
CREATE INDEX idx_monitor_active ON monitor(active);

PRAGMA foreign_keys=ON;
