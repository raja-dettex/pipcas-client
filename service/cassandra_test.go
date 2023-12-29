package service_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/raja-dettex/pipcas-client/service"
)

const (
	astra_db_id       = "9e4be5ec-c374-4219-93cb-2346169c8a8d"
	astra_db_Region   = "us-east1"
	astra_db_token    = "AstraCS:WPmMwfdEhfIvqqwfUkTBvYcR:0ab3f34307cacedee72d3262b52ecbe7a988d9a074c8af439f4ec35c1225bcbb"
	astra_db_keyspace = "filestore_keyspace"
)

func TestCassandraService(t *testing.T) {
	tableOpts := service.TableOpts{TableName: "files_by_storage", Column1: "storage_uri", Column2: "file_name"}
	opts := service.CassandraDBOpts{Url: "https://api.astra.datastax.com", AstraDbId: astra_db_id, AstraDbToken: astra_db_token, AstraDbKeyspace: astra_db_keyspace, TableOpts: tableOpts}
	context := context.Background()
	db, err := service.NewCassandraDB(opts, context)
	if err != nil {
		log.Fatal("error connecting to cluster", err)
	}
	err = db.Save("localhost:5000", "test.txt")
	fmt.Println(err)
	loc, err := db.GetBy("test.txt")
	fmt.Println(err)
	fmt.Println("location is ", loc)
}
