# Minerva

Minerva is a tool for processing, analyzing, and visualizing log data. It focuses on extracting meaningful insights from network logs, particularly suspicious activity, and includes functionality for geolocation lookups, database storage, and reporting.

## Features

- **Log Processing**: Parse network logs for suspicious activity.
- **Geolocation**: Fetch location data for source IPs.
- **Database Integration**: Store processed data in PostgreSQL.
- **Reporting**: Summarize activity in structured outputs.
- **Extensible**: Modular design to allow for future enhancements.

## Getting Started

### Prerequisites

- Go (latest stable version recommended)
- PostgreSQL database
- [ip-api.com](https://ip-api.com) API key (optional for geolocation)

### Installation

1. Clone the repository:

   ```bash
   git clone <repository-url>
   cd minerva
   ```

2. Install dependencies:

   ```bash
   go mod download
   ```

3. Set up the PostgreSQL database using the schema in `docs/schema.sql`.

### Configuration

Update the database connection details in `main.go`:

```go
db.Connect("localhost", "5432", "minerva_user", "secure_password", "minerva")
```

## Usage

### Running the Tool

To process logs from a file:

```bash
cat /path/to/log/file | go run cmd/minerva/main.go
```

Optional flags:

- `-limit`: Limit the number of entries to process (default: no limit).
- `-r`: Process logs in reverse (oldest-first) order.

### SSH Configuration for Log Server

To simplify SSH access to the log server, configure a name in your `~/.ssh/config` file. This allows you to reference the server using an alias instead of its IP address. Add the following lines to your `~/.ssh/config`:

```plaintext
Host logserver
    HostName <IP_ADDRESS>
    User <USERNAME>
    IdentityFile ~/.ssh/<PRIVATE_KEY>
```

Replace `<IP_ADDRESS>`, `<USERNAME>`, and `<PRIVATE_KEY>` with the appropriate values for your setup. You can now SSH into the log server with:

```bash
ssh logserver
```

### Example Command

```bash
ssh logserver "cat /var/log/syslog" | go run cmd/minerva/main.go -limit 50
```

### Generating Reports

Summarized IP activity reports can be generated by invoking the reporting functionality in `output.go`. The reports are stored in structured formats for analysis.

## Development

### To-Do List

Refer to [`todo.md`](./todo.md) for current development goals and tasks.

### Testing

Run unit tests:

```bash
go test ./... -v
```

## Contributing

Contributions are welcome! Please fork the repository and open a pull request with your proposed changes.

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
