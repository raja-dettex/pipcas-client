package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/raja-dettex/pipcas-client/server"
	"github.com/raja-dettex/pipcas-client/service"
	"github.com/raja-dettex/pipcas-client/storage"
)

var (
	astra_db_id       = os.Getenv("ASTRA_DB_ID")
	astra_db_region   = os.Getenv("ASTRA_DB_REGION")
	astra_db_token    = os.Getenv("ASTRA_DB_TOKEN")
	astra_db_keyspace = os.Getenv("KEYSPACE")
	LISTENADDR        = os.Getenv("LISTEN_ADDR")
	GATEWAY_HOST      = os.Getenv("GATEWAY_HOST")
	GATEWAY_PORT      = os.Getenv("GATEWAY_PORT")
	TTL               = os.Getenv("TTL")
)

func main() {
	tableOpts := service.TableOpts{TableName: "files_by_storage", Column1: "storage_uri", Column2: "file_name"}
	opts := service.CassandraDBOpts{Url: "https://api.astra.datastax.com", AstraDbId: astra_db_id, AstraDbToken: astra_db_token, AstraDbKeyspace: astra_db_keyspace, TableOpts: tableOpts}
	context := context.Background()
	db, err := service.NewCassandraDB(opts, context)
	if err != nil {
		log.Fatal("error connecting to cluster", err)
	}
	sOpts := server.ServerOpts{ListenAddr: ":2000", PipCasHost: "localhost", PipCasPort: "4000"}
	ttl, err := strconv.Atoi(TTL)
	if err != nil {
		log.Fatal("invalid ttl")
	}
	storage := storage.NewLocalFileStorage(ttl)
	s := server.NewAPIServer(sOpts, storage, db)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("server running on port : ", sOpts.ListenAddr)

}
