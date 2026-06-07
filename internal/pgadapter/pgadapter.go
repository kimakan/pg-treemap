// Package pgadapter queries the Postgres database(s) and returns information
// about its/their structure.
package pgadapter

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/kimakan/pg-treemap/internal/model"
)

type Adapter struct {
	databaseURL string
	conn        *pgx.Conn
}

func NewAdapter(databaseURL string) (*Adapter, error) {
	adapter := Adapter{
		databaseURL: databaseURL,
	}
	err := adapter.connect()
	if err != nil {
		return nil, err
	}
	defer adapter.close()

	return &adapter, nil
}

func (adapter *Adapter) connect() error {
	conn, err := pgx.Connect(context.Background(), adapter.databaseURL)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v\n", err)
	}
	adapter.conn = conn
	return nil
}

func (adapter *Adapter) close() error {
	err := adapter.conn.Close(context.Background())
	adapter.conn = nil
	if err != nil {
		return fmt.Errorf("unable to close connection to the database: %v", err)
	}
	return nil
}

func (adapter *Adapter) GetAllDatabaseNames() ([]string, error) {
	err := adapter.connect()
	if err != nil {
		return nil, err
	}
	defer adapter.close()

	rows, _ := adapter.conn.Query(context.Background(), "select datname from pg_database order by datname desc;")
	defer rows.Close()

	databases, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("unable to get names of the databases : %v", err)
	}

	return databases, nil
}

func (adapter *Adapter) GetAllSchemaNames() ([]string, error) {
	err := adapter.connect()
	if err != nil {
		return nil, err
	}
	defer adapter.close()

	rows, err := adapter.conn.Query(context.Background(), "select nspname as schema_name from pg_namespace;")
	if err != nil {
		return nil, fmt.Errorf("unable to query the database: %v", err)
	}
	defer rows.Close()

	schemaNames, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("unable to get names of the schemas : %v", err)
	}

	return schemaNames, nil
}

func (adapter *Adapter) GetDatabaseMetadata(datname string) (*model.DatabaseMetadata, error) {
	err := adapter.connect()
	if err != nil {
		return nil, err
	}
	defer adapter.close()

	// Get the total size of the database
	rows, err := adapter.conn.Query(
		context.Background(),
		"select datname, pg_database_size(datname) as db_size from pg_database where datname=$1;",
		datname,
	)
	defer rows.Close()

	if err != nil {
		return nil, fmt.Errorf("unable to query the database: %v\n", err)
	}

	database, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.DatabaseMetadata])
	if err != nil {
		return nil, fmt.Errorf("unable to get metadata for the database '%s': %v\n", datname, err)
	}

	return &database, nil
}

func (adapter *Adapter) GetSchemaMetadata(schemaName string) (*model.SchemaMetadata, error) {
	err := adapter.connect()
	defer adapter.close()
	if err != nil {
		return nil, err
	}

	// get all tables
	rows, err := adapter.conn.Query(
		context.Background(), `
		SELECT
			c.relname as table_name,
			pg_total_relation_size(c.oid) as table_size,
			pg_relation_size(c.oid) as heap_size,
			pg_indexes_size(c.oid) as indexes_size
			FROM pg_class c
			JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE n.nspname = $1
  AND c.relkind = 'r';
`,
		schemaName,
	)
	if err != nil {
		return nil, err
	}

	tables, _ := pgx.CollectRows(rows, pgx.RowToStructByName[model.TableMetadata])

	var schemaSize int64 = 0
	for _, t := range tables {
		schemaSize += t.Size
	}

	for i := range tables {
		indexes := model.StoragePart{
			Name: "indexes",
			Size: tables[i].IndexesSize,
		}
		if indexes.Size > 0 {
			// Get indexes
			rows, err = adapter.conn.Query(
				context.Background(), `
				SELECT
					i.relname AS name,
					pg_relation_size(i.oid) AS size
				FROM pg_index x
				JOIN pg_class t ON t.oid = x.indrelid
				JOIN pg_class i ON i.oid = x.indexrelid
				JOIN pg_namespace n ON n.oid = t.relnamespace
				WHERE n.nspname = $1
				AND t.relname = $2;
		`,
				schemaName, tables[i].TableName,
			)
			if err != nil {
				return nil, fmt.Errorf("cannot retrieve the index sizes for table '%s'.'%s' :%v", schemaName, tables[i].TableName, err)
			}
			indexMetadata, _ := pgx.CollectRows(rows, pgx.RowToStructByName[model.StoragePart])
			indexes.Children = indexMetadata
		}

		heap := model.StoragePart{
			Name: "heap",
			Size: tables[i].HeapSize,
		}
		toast := model.StoragePart{
			Name: "toast",
			Size: tables[i].Size - heap.Size - indexes.Size,
		}
		tables[i].StorageParts = []model.StoragePart{heap, toast, indexes}
	}

	schemaMetadata := model.SchemaMetadata{
		SchemaName: schemaName,
		Size:       schemaSize,
		Tables:     tables,
	}

	rows.Close()
	return &schemaMetadata, nil
}
