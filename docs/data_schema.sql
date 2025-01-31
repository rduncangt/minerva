--
-- data_schema.sql
--
-- This file defines the schema for the log_data and ip_geo tables. 
-- It includes table creation statements, indexes, and foreign key constraints.
--

--
-- log_data - Table to store log entries
--

CREATE TABLE log_data (
    id SERIAL PRIMARY KEY,           -- Unique identifier for each log entry
    timestamp TIMESTAMP NOT NULL,     -- The exact time the log entry was recorded
    source_ip TEXT NOT NULL,          -- The IP address from which the packet originated
    destination_ip TEXT NOT NULL,     -- The IP address to which the packet was directed
    protocol TEXT NOT NULL,           -- The protocol used (e.g., TCP, UDP, ICMP)
    source_port INTEGER,              -- The source port of the packet
    destination_port INTEGER,         -- The destination port of the packet
    action TEXT,                      -- Tracks what was done to the packet
    reason TEXT,                      -- Categorizes the attack or packet handling.
    packet_length INTEGER,            -- The size of the packet. Useful for traffic pattern analysis
    ttl INTEGER                       -- Time-to-Live (TTL) value. Indicates distance or latency to the source
);

--
-- Create indexes on log_data table for faster querying
--

CREATE INDEX idx_log_timestamp ON log_data(timestamp);
CREATE INDEX idx_log_source_ip ON log_data(source_ip);
CREATE INDEX idx_log_destination_ip ON log_data(destination_ip);
CREATE INDEX idx_log_action ON log_data(action);
CREATE INDEX idx_log_reason ON log_data(reason);


GRANT INSERT, SELECT ON log_data TO minerva_user;
GRANT UPDATE, DELETE ON log_data TO minerva_user;

GRANT USAGE, SELECT ON SEQUENCE log_data_id_seq TO minerva_user;
GRANT UPDATE ON SEQUENCE log_data_id_seq TO minerva_user;

--
-- ip_geo - Table to store IP address geolocation data
--

CREATE TABLE ip_geo (
    id SERIAL PRIMARY KEY,
    ip_address TEXT UNIQUE NOT NULL,
    country TEXT,
    region TEXT,
    city TEXT,
    isp TEXT,
    last_updated TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_ip_address ON ip_geo(ip_address);

-- Define a unique constraint to prevent duplicate log entries
ALTER TABLE log_data
    ADD CONSTRAINT unique_log_entry UNIQUE (timestamp, source_ip, destination_ip, protocol, source_port, destination_port);

--
-- Add foreign key constraints to ensure data integrity
--

ALTER TABLE log_data
    ADD CONSTRAINT fk_log_data_source_ip_ip_geo
        FOREIGN KEY (source_ip) REFERENCES ip_geo(ip_address);

ALTER TABLE log_data
    ADD CONSTRAINT fk_log_data_destination_ip_ip_geo
        FOREIGN KEY (destination_ip) REFERENCES ip_geo(ip_address);