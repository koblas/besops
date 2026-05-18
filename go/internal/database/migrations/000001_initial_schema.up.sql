-- Initial schema matching Uptime Kuma's existing database structure.
-- Uses UUID primary keys (TEXT) for all entities.

CREATE TABLE IF NOT EXISTS "user" (
    id TEXT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT 1,
    timezone VARCHAR(64) DEFAULT NULL,
    twofa_secret VARCHAR(64) DEFAULT NULL,
    twofa_status BOOLEAN NOT NULL DEFAULT 0,
    twofa_last_token VARCHAR(6) DEFAULT NULL,
    created_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS monitor (
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
    push_token VARCHAR(64) DEFAULT NULL,
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
    parent_id TEXT DEFAULT NULL,
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
    CONSTRAINT fk_monitor_user FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE,
    CONSTRAINT fk_monitor_parent FOREIGN KEY (parent_id) REFERENCES monitor(id) ON DELETE SET NULL
);

CREATE INDEX idx_monitor_user_id ON monitor(user_id);
CREATE INDEX idx_monitor_active ON monitor(active);

CREATE TABLE IF NOT EXISTS heartbeat (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL,
    status INTEGER NOT NULL,
    msg TEXT DEFAULT NULL,
    time TIMESTAMP NOT NULL,
    latency INTEGER DEFAULT NULL,
    important BOOLEAN NOT NULL DEFAULT 0,
    duration INTEGER NOT NULL DEFAULT 0,
    down_count INTEGER NOT NULL DEFAULT 0,
    end_time TIMESTAMP DEFAULT NULL,
    retries INTEGER NOT NULL DEFAULT 0,
    response BLOB DEFAULT NULL,
    CONSTRAINT fk_heartbeat_monitor FOREIGN KEY (monitor_id) REFERENCES monitor(id) ON DELETE CASCADE
);

CREATE INDEX idx_heartbeat_monitor_id ON heartbeat(monitor_id);
CREATE INDEX idx_heartbeat_monitor_time ON heartbeat(monitor_id, time);
CREATE INDEX idx_heartbeat_important ON heartbeat(important);
CREATE INDEX idx_heartbeat_monitor_important_time ON heartbeat(monitor_id, important, time);

CREATE TABLE IF NOT EXISTS notification (
    id TEXT PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT 1,
    user_id TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT 0,
    config TEXT NOT NULL,
    CONSTRAINT fk_notification_user FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);

CREATE INDEX idx_notification_user_id ON notification(user_id);

CREATE TABLE IF NOT EXISTS monitor_notification (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL,
    notification_id TEXT NOT NULL,
    CONSTRAINT fk_mn_monitor FOREIGN KEY (monitor_id) REFERENCES monitor(id) ON DELETE CASCADE,
    CONSTRAINT fk_mn_notification FOREIGN KEY (notification_id) REFERENCES notification(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_monitor_notification_unique ON monitor_notification(monitor_id, notification_id);

CREATE TABLE IF NOT EXISTS tag (
    id TEXT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(9) NOT NULL DEFAULT '#4B5563',
    created_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS monitor_tag (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    value VARCHAR(255) DEFAULT NULL,
    CONSTRAINT fk_mt_monitor FOREIGN KEY (monitor_id) REFERENCES monitor(id) ON DELETE CASCADE,
    CONSTRAINT fk_mt_tag FOREIGN KEY (tag_id) REFERENCES tag(id) ON DELETE CASCADE
);

CREATE INDEX idx_monitor_tag_monitor_id ON monitor_tag(monitor_id);

CREATE TABLE IF NOT EXISTS status_page (
    id TEXT PRIMARY KEY,
    slug VARCHAR(128) NOT NULL UNIQUE,
    title VARCHAR(150) NOT NULL,
    description TEXT DEFAULT NULL,
    icon TEXT DEFAULT NULL,
    theme VARCHAR(10) NOT NULL DEFAULT 'auto',
    published BOOLEAN NOT NULL DEFAULT 1,
    search_engine_index BOOLEAN NOT NULL DEFAULT 1,
    show_tags BOOLEAN NOT NULL DEFAULT 0,
    password VARCHAR(255) DEFAULT NULL,
    created_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_date TIMESTAMP DEFAULT NULL,
    footer_text TEXT DEFAULT NULL,
    custom_css TEXT DEFAULT NULL,
    show_powered_by BOOLEAN NOT NULL DEFAULT 1,
    show_certificate_expiry BOOLEAN NOT NULL DEFAULT 0,
    google_analytics_tag_id VARCHAR(30) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS "group" (
    id TEXT PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    created_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    public BOOLEAN NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT 1,
    weight INTEGER NOT NULL DEFAULT 1000,
    status_page_id TEXT DEFAULT NULL,
    CONSTRAINT fk_group_status_page FOREIGN KEY (status_page_id) REFERENCES status_page(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS monitor_group (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL,
    group_id TEXT NOT NULL,
    weight INTEGER NOT NULL DEFAULT 1000,
    send_url BOOLEAN NOT NULL DEFAULT 0,
    CONSTRAINT fk_mg_monitor FOREIGN KEY (monitor_id) REFERENCES monitor(id) ON DELETE CASCADE,
    CONSTRAINT fk_mg_group FOREIGN KEY (group_id) REFERENCES "group"(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS incident (
    id TEXT PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    style VARCHAR(10) NOT NULL DEFAULT 'info',
    created_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_updated_date TIMESTAMP DEFAULT NULL,
    pin BOOLEAN NOT NULL DEFAULT 1,
    active BOOLEAN NOT NULL DEFAULT 1,
    status_page_id TEXT DEFAULT NULL,
    CONSTRAINT fk_incident_status_page FOREIGN KEY (status_page_id) REFERENCES status_page(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS maintenance (
    id TEXT PRIMARY KEY,
    title VARCHAR(150) NOT NULL,
    description TEXT DEFAULT NULL,
    user_id TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT 1,
    strategy VARCHAR(30) NOT NULL DEFAULT 'manual',
    start_date TIMESTAMP DEFAULT NULL,
    end_date TIMESTAMP DEFAULT NULL,
    start_time VARCHAR(5) DEFAULT NULL,
    end_time VARCHAR(5) DEFAULT NULL,
    weekdays TEXT DEFAULT '[]',
    days_of_month TEXT DEFAULT '[]',
    interval_day INTEGER DEFAULT 1,
    cron VARCHAR(100) DEFAULT NULL,
    timezone VARCHAR(64) DEFAULT NULL,
    duration INTEGER DEFAULT NULL,
    CONSTRAINT fk_maintenance_user FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);

CREATE INDEX idx_maintenance_active ON maintenance(active);
CREATE INDEX idx_maintenance_user_id ON maintenance(user_id);

CREATE TABLE IF NOT EXISTS maintenance_status_page (
    id TEXT PRIMARY KEY,
    status_page_id TEXT NOT NULL,
    maintenance_id TEXT NOT NULL,
    CONSTRAINT fk_msp_status_page FOREIGN KEY (status_page_id) REFERENCES status_page(id) ON DELETE CASCADE,
    CONSTRAINT fk_msp_maintenance FOREIGN KEY (maintenance_id) REFERENCES maintenance(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS maintenance_timeslot (
    id TEXT PRIMARY KEY,
    maintenance_id TEXT NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP DEFAULT NULL,
    generated_next BOOLEAN NOT NULL DEFAULT 0,
    CONSTRAINT fk_mts_maintenance FOREIGN KEY (maintenance_id) REFERENCES maintenance(id) ON DELETE CASCADE
);

CREATE INDEX idx_maintenance_timeslot_maintenance ON maintenance_timeslot(maintenance_id);
CREATE INDEX idx_maintenance_timeslot_range ON maintenance_timeslot(maintenance_id, start_date, end_date);

CREATE TABLE IF NOT EXISTS monitor_maintenance (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL,
    maintenance_id TEXT NOT NULL,
    CONSTRAINT fk_mm_monitor FOREIGN KEY (monitor_id) REFERENCES monitor(id) ON DELETE CASCADE,
    CONSTRAINT fk_mm_maintenance FOREIGN KEY (maintenance_id) REFERENCES maintenance(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS monitor_tls_info (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL UNIQUE,
    info_json TEXT DEFAULT NULL,
    CONSTRAINT fk_tls_monitor FOREIGN KEY (monitor_id) REFERENCES monitor(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS proxy (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    protocol VARCHAR(10) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,
    auth BOOLEAN NOT NULL DEFAULT 0,
    username VARCHAR(255) DEFAULT NULL,
    password VARCHAR(255) DEFAULT NULL,
    active BOOLEAN NOT NULL DEFAULT 1,
    "default" BOOLEAN NOT NULL DEFAULT 0,
    created_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_proxy_user FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS setting (
    id TEXT PRIMARY KEY,
    key VARCHAR(200) NOT NULL UNIQUE,
    value TEXT DEFAULT NULL,
    type VARCHAR(20) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS api_key (
    id TEXT PRIMARY KEY,
    key VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    user_id TEXT NOT NULL,
    created_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN NOT NULL DEFAULT 1,
    expires TIMESTAMP DEFAULT NULL,
    CONSTRAINT fk_apikey_user FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notification_sent_history (
    id TEXT PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    monitor_id TEXT NOT NULL,
    days INTEGER NOT NULL,
    CONSTRAINT fk_nsh_monitor FOREIGN KEY (monitor_id) REFERENCES monitor(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_notification_sent_unique ON notification_sent_history(type, monitor_id, days);

CREATE TABLE IF NOT EXISTS status_page_cname (
    id TEXT PRIMARY KEY,
    status_page_id TEXT NOT NULL,
    domain VARCHAR(255) NOT NULL,
    CONSTRAINT fk_cname_status_page FOREIGN KEY (status_page_id) REFERENCES status_page(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS remote_browser (
    id TEXT PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    url TEXT NOT NULL,
    user_id TEXT NOT NULL,
    CONSTRAINT fk_rb_user FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS stat_minutely (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    latency REAL DEFAULT NULL,
    latency_min INTEGER DEFAULT NULL,
    latency_max INTEGER DEFAULT NULL,
    up INTEGER NOT NULL DEFAULT 0,
    down INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT fk_stat_min_monitor FOREIGN KEY (monitor_id) REFERENCES monitor(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_stat_minutely_unique ON stat_minutely(monitor_id, timestamp);

CREATE TABLE IF NOT EXISTS stat_hourly (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    latency REAL DEFAULT NULL,
    latency_min INTEGER DEFAULT NULL,
    latency_max INTEGER DEFAULT NULL,
    up INTEGER NOT NULL DEFAULT 0,
    down INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT fk_stat_hour_monitor FOREIGN KEY (monitor_id) REFERENCES monitor(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_stat_hourly_unique ON stat_hourly(monitor_id, timestamp);

CREATE TABLE IF NOT EXISTS stat_daily (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    latency REAL DEFAULT NULL,
    latency_min INTEGER DEFAULT NULL,
    latency_max INTEGER DEFAULT NULL,
    up INTEGER NOT NULL DEFAULT 0,
    down INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT fk_stat_daily_monitor FOREIGN KEY (monitor_id) REFERENCES monitor(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_stat_daily_unique ON stat_daily(monitor_id, timestamp);

CREATE TABLE IF NOT EXISTS domain_expiry (
    id TEXT PRIMARY KEY,
    last_check TIMESTAMP DEFAULT NULL,
    domain VARCHAR(255) NOT NULL UNIQUE,
    expiry TIMESTAMP DEFAULT NULL,
    last_expiry_notification_sent TIMESTAMP DEFAULT NULL
);
