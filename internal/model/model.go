// Package model holds the structs for the metadata of databases,
// schemata and tables
package model

type HostMetadata struct {
	Label     string             `json:"name"`
	Size      int64             `json:"value"`
	Databases []DatabaseMetadata `json:"children,omitempty"`
}

type DatabaseMetadata struct {
	DBname  string           `db:"datname" json:"name"`
	Size    int64            `db:"db_size" json:"value"`
	Schemas []SchemaMetadata `db:"-"       json:"children,omitempty"`
}

type SchemaMetadata struct {
	SchemaName string          `db:"schema_name" json:"name"`
	Size       int64           `db:"schema_size" json:"value"`
	Tables     []TableMetadata `db:"-"           json:"children,omitempty"`
}

type TableMetadata struct {
	TableName string `db:"table_name" json:"name"`
	Size      int64  `db:"table_size" json:"value"`
}
