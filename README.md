# pg-treemap

`pg-treemap` collects PostgreSQL database, schema, and table sizes, writes them
to a local `snapshot.json` file, and serves a browser-based treemap for exploring
where storage is being used.

## Requirements

- Go 1.26 or newer
- Network access to the PostgreSQL host or hosts you want to inspect
- A PostgreSQL role that can connect to the maintenance database and the
  databases being inspected

## Quick Start

Create the default configuration file:

```sh
go run ./cmd/pgtm -create-conf
```

Edit `conf.json` for your PostgreSQL connection details:

```json
{
  "hosts": [
    {
      "host": "localhost",
      "name": "postgres",
      "port": "5432",
      "user": "treemap_collector",
      "password": "verysecretpw"
    }
  ],
  "serve_addr": "localhost:8080"
}
```

Collect metadata into `snapshot.json`:

```sh
go run ./cmd/pgtm -collect
```

Start the web UI:

```sh
go run ./cmd/pgtm -serve
```

Open the address from `serve_addr`, for example:

```text
http://localhost:8080
```

## Configuration

The CLI reads `conf.json` by default. Use `-conf` to load a different file:

```sh
go run ./cmd/pgtm -conf ./prod-conf.json -collect
go run ./cmd/pgtm -conf ./prod-conf.json -serve
```

## PostgreSQL Access

The collector connects first to the `postgres` database to list databases, then
connects to each non-template database to collect schema and table sizes.

A practical collector role usually needs:

```sql
CREATE ROLE treemap_collector LOGIN PASSWORD 'verysecretpw';
GRANT CONNECT ON DATABASE your_database TO treemap_collector;
```

Repeat the `GRANT CONNECT` command for each database that should appear in the
treemap. The role must also be able to read PostgreSQL catalog metadata in those
databases.

## Output

`-collect` writes `snapshot.json` in the current working directory. `-serve`
serves the embedded web UI and exposes that same file through `/api`.

Because `snapshot.json` is a point-in-time snapshot, rerun collection whenever
you want to refresh the treemap:

```sh
go run ./cmd/pgtm -collect
go run ./cmd/pgtm -serve
```

## Build

Build a standalone binary:

```sh
go build -o pgtm ./cmd/pgtm
```

Then use it the same way:

```sh
./pgtm -create-conf
./pgtm -collect
./pgtm -serve
```

