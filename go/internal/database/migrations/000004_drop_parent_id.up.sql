-- Backfill: convert parentId-based group membership to tag-based membership.
-- For each group monitor that has children, create a tag and assign it.

INSERT INTO tag (id, name, color, created_date)
SELECT
    lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' || substr(hex(randomblob(2)),2) || '-' || substr('89ab', abs(random()) % 4 + 1, 1) || substr(hex(randomblob(2)),2) || '-' || hex(randomblob(6))),
    'group:' || g.name,
    '#6366f1',
    CURRENT_TIMESTAMP
FROM monitor g
WHERE g.type = 'group'
  AND EXISTS (SELECT 1 FROM monitor c WHERE c.parent_id = g.id);

INSERT INTO monitor_tag (id, tag_id, monitor_id, value)
SELECT
    lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' || substr(hex(randomblob(2)),2) || '-' || substr('89ab', abs(random()) % 4 + 1, 1) || substr(hex(randomblob(2)),2) || '-' || hex(randomblob(6))),
    t.id,
    c.id,
    ''
FROM monitor c
JOIN monitor g ON c.parent_id = g.id AND g.type = 'group'
JOIN tag t ON t.name = 'group:' || g.name;

UPDATE monitor
SET group_tag_ids_json = (
    SELECT '["' || t.id || '"]'
    FROM tag t
    WHERE t.name = 'group:' || monitor.name
)
WHERE type = 'group'
  AND EXISTS (SELECT 1 FROM monitor c WHERE c.parent_id = monitor.id);

-- SQLite requires table recreation to drop a column with a FK constraint.
PRAGMA foreign_keys=OFF;

CREATE TABLE monitor_new (
    id TEXT PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT 1,
    user_id TEXT NOT NULL,
    interval INTEGER NOT NULL DEFAULT 60,
    url TEXT DEFAULT NULL,
    type VARCHAR(30) NOT NULL,
    weight INTEGER DEFAULT 2000,
    hostname VARCHAR(255) DEFAULT NULL,
    port INTEGER DEFAULT NULL,
    created_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    keyword VARCHAR(500) DEFAULT NULL,
    maxretries INTEGER NOT NULL DEFAULT 0,
    ignore_tls BOOLEAN NOT NULL DEFAULT 0,
    upside_down BOOLEAN NOT NULL DEFAULT 0,
    maxredirects INTEGER NOT NULL DEFAULT 10,
    accepted_statuscodes_json TEXT DEFAULT '["200-299"]',
    dns_resolve_type VARCHAR(10) DEFAULT 'A',
    dns_resolve_server VARCHAR(255) DEFAULT NULL,
    dns_last_result TEXT DEFAULT NULL,
    retry_interval INTEGER NOT NULL DEFAULT 60,
    method VARCHAR(10) NOT NULL DEFAULT 'GET',
    body TEXT DEFAULT NULL,
    headers TEXT DEFAULT NULL,
    basic_auth_user VARCHAR(255) DEFAULT NULL,
    basic_auth_pass VARCHAR(255) DEFAULT NULL,
    proxy_id TEXT DEFAULT NULL,
    expiry_notification BOOLEAN DEFAULT 0,
    mqtt_topic VARCHAR(255) DEFAULT NULL,
    mqtt_success_message VARCHAR(255) DEFAULT NULL,
    mqtt_username VARCHAR(255) DEFAULT NULL,
    mqtt_password VARCHAR(255) DEFAULT NULL,
    database_connection_string TEXT DEFAULT NULL,
    database_query TEXT DEFAULT NULL,
    auth_method VARCHAR(50) DEFAULT NULL,
    auth_domain VARCHAR(255) DEFAULT NULL,
    auth_workstation VARCHAR(255) DEFAULT NULL,
    grpc_url TEXT DEFAULT NULL,
    grpc_protobuf TEXT DEFAULT NULL,
    grpc_body TEXT DEFAULT NULL,
    grpc_metadata TEXT DEFAULT NULL,
    grpc_method VARCHAR(255) DEFAULT NULL,
    grpc_service_name VARCHAR(255) DEFAULT NULL,
    grpc_enable_tls BOOLEAN NOT NULL DEFAULT 0,
    radius_username VARCHAR(255) DEFAULT NULL,
    radius_password VARCHAR(255) DEFAULT NULL,
    radius_calling_station_id VARCHAR(255) DEFAULT NULL,
    radius_called_station_id VARCHAR(255) DEFAULT NULL,
    radius_secret VARCHAR(255) DEFAULT NULL,
    resend_interval INTEGER NOT NULL DEFAULT 0,
    packet_size INTEGER NOT NULL DEFAULT 56,
    game VARCHAR(255) DEFAULT NULL,
    http_body_encoding VARCHAR(30) DEFAULT NULL,
    description TEXT DEFAULT NULL,
    tls_ca TEXT DEFAULT NULL,
    tls_cert TEXT DEFAULT NULL,
    tls_key TEXT DEFAULT NULL,
    invert_keyword BOOLEAN NOT NULL DEFAULT 0,
    json_path VARCHAR(255) DEFAULT NULL,
    expected_value VARCHAR(255) DEFAULT NULL,
    kafka_producer_topic VARCHAR(255) DEFAULT NULL,
    kafka_producer_brokers TEXT DEFAULT NULL,
    kafka_producer_ssl BOOLEAN NOT NULL DEFAULT 0,
    kafka_producer_allow_auto_topic_creation BOOLEAN NOT NULL DEFAULT 0,
    kafka_producer_sasl_options TEXT DEFAULT NULL,
    kafka_producer_message TEXT DEFAULT NULL,
    oauth_client_id VARCHAR(255) DEFAULT NULL,
    oauth_client_secret VARCHAR(255) DEFAULT NULL,
    oauth_token_url TEXT DEFAULT NULL,
    oauth_scopes VARCHAR(500) DEFAULT NULL,
    oauth_auth_method VARCHAR(50) DEFAULT NULL,
    timeout INTEGER NOT NULL DEFAULT 48,
    gamedig_given_port_only BOOLEAN NOT NULL DEFAULT 1,
    save_response BOOLEAN NOT NULL DEFAULT 0,
    save_error_response BOOLEAN NOT NULL DEFAULT 0,
    response_max_length INTEGER NOT NULL DEFAULT 0,
    system_service_name VARCHAR(255) DEFAULT NULL,
    rabbitmq_nodes TEXT DEFAULT NULL,
    rabbitmq_username VARCHAR(255) DEFAULT NULL,
    rabbitmq_password VARCHAR(255) DEFAULT NULL,
    remote_browser TEXT DEFAULT NULL,
    domain_expiry_notification BOOLEAN DEFAULT 0,
    group_tag_ids_json TEXT DEFAULT NULL,
    CONSTRAINT fk_monitor_user FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);

INSERT INTO monitor_new SELECT
    id, name, active, user_id, interval, url, type, weight, hostname, port,
    created_date, keyword, maxretries, ignore_tls, upside_down, maxredirects,
    accepted_statuscodes_json, dns_resolve_type, dns_resolve_server, dns_last_result,
    retry_interval, method, body, headers, basic_auth_user, basic_auth_pass,
    proxy_id, expiry_notification, mqtt_topic, mqtt_success_message, mqtt_username,
    mqtt_password, database_connection_string, database_query, auth_method, auth_domain,
    auth_workstation, grpc_url, grpc_protobuf, grpc_body, grpc_metadata, grpc_method,
    grpc_service_name, grpc_enable_tls, radius_username, radius_password,
    radius_calling_station_id, radius_called_station_id, radius_secret, resend_interval,
    packet_size, game, http_body_encoding, description, tls_ca, tls_cert, tls_key,
    invert_keyword, json_path, expected_value, kafka_producer_topic, kafka_producer_brokers,
    kafka_producer_ssl, kafka_producer_allow_auto_topic_creation, kafka_producer_sasl_options,
    kafka_producer_message, oauth_client_id, oauth_client_secret, oauth_token_url,
    oauth_scopes, oauth_auth_method, timeout, gamedig_given_port_only, save_response,
    save_error_response, response_max_length, system_service_name, rabbitmq_nodes,
    rabbitmq_username, rabbitmq_password, remote_browser, domain_expiry_notification,
    group_tag_ids_json
FROM monitor;

DROP TABLE monitor;
ALTER TABLE monitor_new RENAME TO monitor;

CREATE INDEX idx_monitor_user_id ON monitor(user_id);
CREATE INDEX idx_monitor_active ON monitor(active);

PRAGMA foreign_keys=ON;
