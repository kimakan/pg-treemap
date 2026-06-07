// Package collector collects metadata and stores a snapshot of it in a file
package collector

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/kimakan/pg-treemap/internal/config"
	"github.com/kimakan/pg-treemap/internal/model"
	"github.com/kimakan/pg-treemap/internal/pgadapter"
)

var ignoredDatabases = []string{
	"postgres",
	"template0",
	"template1",
	"template_postgis",
}

func Collect(conf config.Config) error {
	hosts := make([]model.HostMetadata, 0)
	fmt.Println("Collecting metadata...")
	for _, host := range conf.Hosts {
		hostDatabases, err := collectHost(host)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to collect from %s : %v\n", host.Label, err)
			continue
		}
		fmt.Fprintf(os.Stdout, " - %v: %d database(s)\n", host.Label, len(hostDatabases.Databases))
		hosts = append(hosts, *hostDatabases)
	}

	data, err := json.Marshal(hosts)
	if err != nil {
		return fmt.Errorf("failed to convert fetched metadata into json format: %v", err)
	}

	err = os.WriteFile("snapshot.json", data, 0o644)
	if err != nil {
		return fmt.Errorf("cannot write collected data into the snapshot: %v\n", err)
	}
	hostLabels := make([]string, 0)
	for _, h := range hosts {
		hostLabels = append(hostLabels, h.Label)
	}
	return nil
}

// collectHost collects the metadata (primarily sizes) of all databases
func collectHost(hostConf config.HostConfig) (*model.HostMetadata, error) {
	databaseURL := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s",
		hostConf.User,
		hostConf.Password,
		hostConf.Host,
		hostConf.Port,
	)
	defaultDatabaseName := "postgres"
	databaseFullURL := databaseURL + "/" + defaultDatabaseName
	adapter, err := pgadapter.NewAdapter(databaseFullURL)
	if err != nil {
		return nil, err
	}

	databaseNames, _ := adapter.GetAllDatabaseNames()

	var databases []model.DatabaseMetadata
	for _, datname := range databaseNames {
		newDatabase, err := adapter.GetDatabaseMetadata(datname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: failed fetching metadata from the database: %s: %v\n", datname, err)
		} else {
			// ignore some basic/not interesting databases
			if !slices.Contains(ignoredDatabases, newDatabase.DBname) {
				databases = append(databases, *newDatabase)
			}
		}
	}

	for db_index, database := range databases {
		adapter, err := pgadapter.NewAdapter(databaseURL + "/" + database.DBname)
		if err != nil {
			fmt.Printf("%s: %v", database.DBname, err)
		} else {
			schemaNames, err := adapter.GetAllSchemaNames()
			if err != nil {
				fmt.Printf("Cannot get schemas from %s\n", database.DBname)
			} else {
				databases[db_index].Schemas = make([]model.SchemaMetadata, len(schemaNames))
				for i, schemaName := range schemaNames {
					schemaMetadata, err := adapter.GetSchemaMetadata(schemaName)
					if err != nil {
						fmt.Fprintf(os.Stderr, "schema '%s': %v", schemaName, err)
					} else {
						databases[db_index].Schemas[i] = *schemaMetadata
					}
				}
			}
		}
	}

	// Calc total size
	var hostTotalSize int64 = 0
	for _, database := range databases {
		hostTotalSize += database.Size
	}

	return &model.HostMetadata{
		Label:     hostConf.Label,
		Size:      hostTotalSize,
		Databases: databases,
	}, nil
}
