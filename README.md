# eWHALES v1 Fast Export

This repository contains tools for exporting data from the eWHALES v1 database. There are two primary workflows supported by these tools:

## Backup Processing

The `my.cnf` and `convert_inserts.py` files are intended for use with database backups. They facilitate the conversion and processing of data dumps when a live database connection is not available or desired.

## Live Database Export (Golang Tools)

The Go-based tools in this repository are intended to work directly with a live database. They extract the required records and produce a CSV output in the exact format expected for the v1 eWHALES data processing pipeline.

The architecture consists of three separate tools:
1. **Local Tool (`ewhales-local`)**: The legacy standalone tool that queries the database directly and writes CSV files to the local disk.
2. **Server Tool (`ewhales-server`)**: A gRPC HTTP/2 server that receives streaming records from the client and writes them to CSV files.
3. **Client Tool (`ewhales-client`)**: A lightweight client that queries the database and streams the extracted data securely to the server.

### Building the Tools

Ensure you have Go installed on your system. To compile the binaries, run:
```bash
go build -o ewhales-local ./cmd/local
go build -o ewhales-server ./cmd/server
go build -o ewhales-client ./cmd/client
```

### Running the Tools

You can run the compiled binaries directly. Use the `--help` flag on any of them to see available options.

#### 1. Local Tool
The standalone exporter uses `config.json`.
```bash
./ewhales-local -progress -config config.json
```

#### 2. Server Tool
The server uses `server_config.json` to define the listening port, TLS certificates, and CSV output mapping.
```bash
./ewhales-server -config server_config.json
```

#### 3. Client Tool
The client uses `client_config.json` to define database credentials and the server's network address.
```bash
./ewhales-client -progress -config client_config.json
```

### Memory Statistics

You can monitor the tools' memory footprints while they run. When the `-memstats` flag is enabled (available on `local` and `client`), a background routine will periodically log the memory statistics to a CSV file. It records the internal Go Heap Allocations as well as the Resident Set Size (RSS) of the entire process tree using `gopsutil`.

Example memory profiling usage:
```bash
./ewhales-local -memstats -memstats-interval=5 -memstats-file="metrics.csv"
```

### Configuration

The tools use JSON files for configuration:
- `config.json`: Used by `ewhales-local` (contains database connection details and CSV field mappings).
- `client_config.json`: Used by `ewhales-client` (contains database connection details and `server_address`).
- `server_config.json`: Used by `ewhales-server` (contains `listen_port`, `tls_cert_file`, `tls_key_file`, and CSV field mappings).

Ensure your configurations are correctly set up before running the tools.

### TLS Certificates

The client and server communicate securely using gRPC over HTTP/2 with TLS encryption.

**Server Side (Development):**
When the server starts, it checks for the TLS files specified in `server_config.json` (e.g., `server.crt` and `server.key`). If it doesn't find them, it will automatically generate a self-signed development certificate and key for you.

**Client Side:**
By default in this repository, the client uses `InsecureSkipVerify: true` in its TLS configuration to seamlessly connect to the development server without needing to install the self-signed certificate into the system's root CA trust store.

If you are moving to production, you should use Let's Encrypt or your organization's CA. If you want to use strict verification with the generated development certificate, you can download the server's public certificate and configure the Go gRPC client to use it.

**Snippet: Fetching the Development Certificate**
You can easily extract the auto-generated public certificate from a running server using `openssl`:
```bash
openssl s_client -showcerts -connect localhost:8443 </dev/null 2>/dev/null | openssl x509 -outform PEM > server.crt
```
*(You can then mount or distribute `server.crt` to your client machines if you implement strict TLS checking in `grpc_client.go`)*

### Generating Test Data

`query_test.go` uses sql dump files (e.g. `test_multiple_logbook_logbook_entries.sql`) to generate test data.

The following query can be used to extract test data from the live database and save it to a sql file. It is recommended to limit the number of logbooks to a small number for testing purposes. Replace `limit 10` with a larger number to increase the amount of data extracted or specify specific logbooks by replacing `limit 10` with `and where meta_value in ("logbook-name-1", "logbook-name-2")`. The query can look for as many logbooks as you'd like, e.g. `and where meta_value in ("Westward-1978-1979", "A. Houghton (bark) 1853-1857", "T. A. Spofford (bark) 1851-1855")`. Just keep in mind that some logbooks have more entries than others.

```
with logbooks as (select
         post_id
     from logswp_postmeta where
                              meta_key = "logbook_id"
                            and meta_value is not null
                            and meta_value <> ''
                            and meta_value REGEXP '^[a-zA-z].*' limit 10),
    logbook_entries as (select
                            post_id
                        from logswp_postmeta where
                                                 meta_key = "logbook_id"
                                               and meta_value <> ''
                                               and meta_value in (select * from logbooks))
select * from logswp_postmeta where post_id in (select * from logbooks) or post_id in (select * from logbook_entries);
```

Once you've exported the data to a `.sql` dump file, it's highly recommended to anonymize it before committing it to the repository as test data, since logbook entries may contain PII, researcher names, or proprietary notes.

You can anonymize the file using the provided `anonymize.py` script. This script randomizes sensitive text fields while perfectly preserving numeric IDs and `post_id` referential integrity so the Golang tests can still parse the logic.

```bash
# Usage: python3 anonymize.py <input.sql> <output.sql>
python3 anonymize.py test_data.sql test_data_anon.sql
```

## Docker Usage

If you prefer not to install Go or Python on your local machine, you can run all the tools seamlessly via Docker and Docker Compose.

### 1. Using Docker Compose
The `docker-compose.yml` file sets up a local MySQL instance, builds the Server container, and builds the Client container.
```bash
docker-compose up -d --build
```
This spins up the server in the background. The client will run and exit once it completes the database export.

### 2. Running Python Tools via Docker Image
You can still run the Python conversion tools dynamically using the base builder image:
```bash
docker build --target builder -t ewhales-builder .

# Run the Python Converters:
docker run --rm -v $(pwd):/app -it ewhales-builder python3 /usr/local/bin/convert_inserts.py input.sql output.sql 5000
docker run --rm -v $(pwd):/app -it ewhales-builder python3 /usr/local/bin/anonymize.py test_data.sql test_data_anon.sql
```