<!-- markdownlint-disable MD041 -->

![Minerva Logo](minerva_300px.png)

# Minerva

Minerva is a tool for processing, analyzing, and visualizing log data from network devices. It extracts meaningful insights from logs by detecting suspicious activity, performing geolocation lookups, and storing results in a PostgreSQL database for further analysis and reporting.

## Features

- **Log Processing**: Real-time parsing of network logs for potential security threats.
- **Geolocation Lookups**: Automatic retrieval of location data for suspicious IP addresses.
- **Database Integration**: Secure storage of processed log data in PostgreSQL.
- **Automation**: Supports automated log ingestion via launchd on macOS (or systemd on Linux).
- **Modular Design**: Easily extendable for additional functionality.

## Getting Started

### Prerequisites

- [Go](https://golang.org) (latest stable version recommended)
- [PostgreSQL](https://www.postgresql.org) database
- (Optional) [ip-api.com](https://ip-api.com) for geolocation lookups (free usage tier available)

### Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/rduncangt/minerva
   cd minerva
   ```

2. Download dependencies:

   ```bash
   go mod download
   ```

3. Set up the database:

   Create a PostgreSQL database and run the provided schema file (e.g., `docs/data_schema.sql`).

4. Configure the application:

   Copy the sample configuration file and update it with your settings:

   ```bash
   cp minerva_config.example.toml minerva_config.toml
   ```

   Edit `minerva_config.toml` to set your database connection details and any other required configuration.

## Usage

### Running the Tool

You can run Minerva directly using Go:

```bash
cat /path/to/log/file | go run cmd/minerva/main.go
```

For production, compile the binary:

```bash
cd cmd/minerva
go build -o minerva
sudo mv minerva /usr/local/bin/minerva
```

Then, run it with your log source. For example:

```bash
ssh logserver "cat /var/log/syslog" | /usr/local/bin/minerva
```

### Automation

Minerva’s log ingestion can be automated using launchd on macOS (or systemd on Linux). Detailed instructions for automation are available in [docs/automation.md](docs/automation.md).

## Testing

Run the full test suite with:

```bash
go test ./... -v
```

## Development

Refer to [`todo.md`](./todo.md) for current development goals and tasks.

## Contributing

Contributions are welcome! Please fork the repository and open a pull request with your changes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
