# PostgreSQL Setup for Minerva

This document provides step-by-step instructions for setting up PostgreSQL on macOS for use with the Minerva project. Follow these steps to ensure PostgreSQL is correctly installed, configured, and ready for use.

---

## 1. Install PostgreSQL

### Using Homebrew

1. Install PostgreSQL version 14:

   ```bash
   brew install postgresql@14
   ```

2. Link PostgreSQL to your environment:

   ```bash
   brew link postgresql@14 --force
   ```

3. Verify the installation:

   ```bash
   postgres --version
   ```

   Expected output should indicate PostgreSQL 14 is installed.

4. Start the PostgreSQL service:

   ```bash
   brew services start postgresql
   ```

### Verify Service Status

Check if the service is running:

```bash
brew services list
```

Ensure `postgresql@14` is listed as `started`.

---

## 2. Initialize the Database Cluster

If the database cluster has not been initialized, perform the following:

1. Initialize the database:

   ```bash
   initdb /usr/local/var/postgres
   ```

2. Start the server manually (optional):

   ```bash
   pg_ctl -D /usr/local/var/postgres start
   ```

---

## 3. Create the Minerva Database

1. Access the PostgreSQL CLI:

   ```bash
   psql postgres
   ```

2. Create the `minerva` database:

   ```sql
   CREATE DATABASE minerva;
   ```

3. Verify the database creation:

   ```sql
   \l
   ```

   The output should list the `minerva` database.

---

## 4. Configure the Database User

1. Create a user for the Minerva application:

   ```sql
   CREATE USER minerva_user WITH PASSWORD 'secure_password';
   ```

2. Grant the user privileges on the `minerva` database:

   ```sql
   GRANT ALL PRIVILEGES ON DATABASE minerva TO minerva_user;
   ```

3. Optional: Configure the user environment:

   ```sql
   ALTER ROLE minerva_user SET client_encoding TO 'utf8';
   ALTER ROLE minerva_user SET default_transaction_isolation TO 'read committed';
   ALTER ROLE minerva_user SET timezone TO 'UTC';
   ```

---

## 5. Define the Database Schema

### Schema File

Save the following schema as `schema.sql`:

```sql
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
```

### Apply the Schema

1. Run the following command to apply the schema:

   ```bash
   psql -U minerva_user -d minerva -f schema.sql
   ```

2. Verify the table creation:

   ```sql
   \dt
   ```

   Ensure the `ip_data` table is listed.

---

## 6. Test Database Connectivity

1. Install the PostgreSQL driver for Go:

   ```bash
   go get github.com/lib/pq
   ```

2. Test connectivity in a Go program:

   ```go
   package main

   import (
       "database/sql"
       "fmt"
       "log"

       _ "github.com/lib/pq"
   )

   func main() {
       connStr := "host=localhost port=5432 user=minerva_user password=secure_password dbname=minerva sslmode=disable"
       db, err := sql.Open("postgres", connStr)
       if err != nil {
           log.Fatalf("Failed to connect to the database: %v", err)
       }

       err = db.Ping()
       if err != nil {
           log.Fatalf("Failed to ping the database: %v", err)
       }

       fmt.Println("Database connection successful!")
   }
   ```

3. Run the program:

   ```bash
   go run main.go
   ```

   Expected output:

   ```
   Database connection successful!
   ```

---

## 7. Maintenance Commands

### Stop the PostgreSQL Service

```bash
brew services stop postgresql
```

### Restart the PostgreSQL Service

```bash
brew services restart postgresql
```

### Access the Minerva Database

```bash
psql -U minerva_user -d minerva
```

---

By following this guide, PostgreSQL should be fully set up and ready for use with Minerva.
