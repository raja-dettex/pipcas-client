package service

import (
	"context"
	"fmt"
	"time"

	gocqlastra "github.com/datastax/gocql-astra"
	"github.com/gocql/gocql"
	"github.com/raja-dettex/pipcas-client/types"
)

type TableOpts struct {
	TableName string
	Column1   string
	Column2   string
}

type CassandraDBOpts struct {
	Url             string
	AstraDbId       string
	AstraDbRegion   string
	AstraDbToken    string
	AstraDbKeyspace string
	TableOpts       TableOpts
}

type CassandraDB struct {
	opts    CassandraDBOpts
	Cluster *gocql.ClusterConfig
	ctx     context.Context
}

func NewCassandraDB(opts CassandraDBOpts, context context.Context) (*CassandraDB, error) {
	cDb := &CassandraDB{opts: opts, ctx: context}
	cluster, err := gocqlastra.NewClusterFromURL(opts.Url, opts.AstraDbId, opts.AstraDbToken, time.Second*20)
	if err != nil {
		return nil, err
	}
	cluster.Keyspace = opts.AstraDbKeyspace

	cDb.Cluster = cluster
	return cDb, nil
}

func (db *CassandraDB) CreateSchema() error {
	session, err := gocql.NewSession(*db.Cluster)
	if err != nil {
		return err
	}
	defer session.Close()
	db.Cluster.Timeout = time.Second * 12
	timeOutContext, cancel := context.WithTimeout(db.ctx, time.Second*8)
	defer cancel()
	errCh := make(chan error)
	query := fmt.Sprintf("create table if not exists %s(%s text, %s text, PRIMARY KEY((%s), %s))", db.opts.TableOpts.TableName, db.opts.TableOpts.Column1, db.opts.TableOpts.Column2,
		db.opts.TableOpts.Column1, db.opts.TableOpts.Column2)
	go func(ch chan error) {
		ch <- session.Query(query).Exec()

	}(errCh)
	for {
		select {
		case err := <-errCh:
			return err
		case <-timeOutContext.Done():
			return types.ErrTimedOut
		}
	}

}

func (db *CassandraDB) Save(storageUri, fileName string) error {
	session, err := gocql.NewSession(*db.Cluster)

	if err != nil {
		return err
	}
	defer session.Close()
	db.Cluster.Timeout = time.Second * 12
	timeOutContext, cancel := context.WithTimeout(db.ctx, time.Second*8)
	defer cancel()
	errCh := make(chan error)
	go func(ch chan error) {
		err := session.Query(`INSERT INTO files_by_storage (storage_uri, file_name) VALUES (?, ?)`,
			storageUri, fileName).Exec()
		fmt.Println("insertion error ", err)
		ch <- err

	}(errCh)
	for {
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
			return nil
		case <-timeOutContext.Done():
			return types.ErrTimedOut
		}
	}
}

func (db *CassandraDB) GetAllBy(storageUri string) (*[]string, error) {
	var fileNames []string
	session, err := gocql.NewSession(*db.Cluster)

	if err != nil {
		return nil, err
	}
	defer session.Close()
	db.Cluster.Timeout = time.Second * 7
	timeOutContext, cancel := context.WithTimeout(db.ctx, time.Second*5)
	defer cancel()
	iterCh := make(chan *gocql.Iter)
	go func(ch chan *gocql.Iter) {
		ch <- session.Query(fmt.Sprintf("select %s from %s where %s=?", db.opts.TableOpts.Column2, db.opts.TableOpts.TableName, db.opts.TableOpts.Column1), storageUri).Iter()

	}(iterCh)
	for {
		select {
		case iter := <-iterCh:
			var value string
			for iter.Scan(&value) {
				fileNames = append(fileNames, value)
			}
			if err := iter.Close(); err != nil {
				return nil, err
			}
			return &fileNames, nil
		case <-timeOutContext.Done():
			return nil, types.ErrTimedOut
		}
	}
}
func (db *CassandraDB) GetBy(fileName string) (string, error) {
	session, err := gocql.NewSession(*db.Cluster)

	if err != nil {
		return "", err
	}
	defer session.Close()
	db.Cluster.Timeout = time.Second * 12
	timeOutContext, cancel := context.WithTimeout(db.ctx, time.Second*8)
	defer cancel()
	iterCh := make(chan *gocql.Iter)
	go func(ch chan *gocql.Iter) {
		ch <- session.Query(fmt.Sprintf("select %s from %s where %s=? ALLOW FILTERING", db.opts.TableOpts.Column1, db.opts.TableOpts.TableName, db.opts.TableOpts.Column2), fileName).Iter()

	}(iterCh)
	for {
		select {
		case iter := <-iterCh:
			var value string
			iter.Scan(&value)

			if err := iter.Close(); err != nil {
				return "", err
			}
			return value, nil

		case <-timeOutContext.Done():
			return "", types.ErrTimedOut
		}
	}
}
