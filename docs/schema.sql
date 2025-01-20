CREATE TABLE ip_data (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    source_ip TEXT NOT NULL,
    destination_ip TEXT NOT NULL,
    protocol TEXT NOT NULL,
    source_port INTEGER,
    destination_port INTEGER,
    country TEXT,
    region TEXT,
    city TEXT,
    isp TEXT
);