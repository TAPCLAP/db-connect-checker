package mysqlcheck

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/tapclap/db-connect-checker/pkg/types"
)

func CheckConnections(config []types.MysqlConfig, tries int) error {
	var errChan = make(chan error, len(config))

	for _, cfg := range config {
		go func(cfg types.MysqlConfig) {
			i := 1
			for i = 1; i <= tries; i += 1 {
				sleepS := 3*i + 1
				sleep := time.Duration(sleepS) * time.Second
				err := CheckConnection(cfg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "[%s:%s/%s] Try (%d/%d) sleep %d seconds error: %v\n", cfg.Host, cfg.Port, cfg.Name, i, tries, sleepS, err)
					time.Sleep(sleep)
					continue
				}
				fmt.Println("Connect success")
				break
			}

			if i == tries+1 {
				errChan <- fmt.Errorf("[%s:%s/%s] connection attempts have failed", cfg.Host, cfg.Name, cfg.Port)
			} else {
				errChan <- nil
			}
		}(cfg)
	}

	for range config {
		if err := <-errChan; err != nil {
			return err
		}
	}
	return nil
}

func CheckConnection(config types.MysqlConfig) error {
	connectString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", config.User, config.Pass, config.Host, config.Port, config.Name)
	if config.TLS {
		tlsConfigName := fmt.Sprintf("custom-tls-%s-%s", config.Host, config.Name)
		err := mysql.RegisterTLSConfig(tlsConfigName, config.TLSConfig)
		if err != nil {
			return fmt.Errorf("cannot register TLS config for MySQL connection: %v", err)
		}
		connectString = fmt.Sprintf("%s?tls=%s", connectString, tlsConfigName)
	}
	db, err := sql.Open("mysql", connectString)
	if err != nil {
		return fmt.Errorf("error connect: %v", err)
	}
	defer db.Close()

	_, err = getSQLTables(db)
	if err != nil {
		return fmt.Errorf("error getting tables: %v", err)
	}

	return nil
}

func getSQLTables(db *sql.DB) ([]string, error) {
	errorFuncName := "Func GetSQLTables() error"
	query := "SHOW TABLES"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tableRows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: query: '%s': %v", errorFuncName, query, err)
	}
	defer tableRows.Close()

	var tables []string
	for tableRows.Next() {
		var table string
		err = tableRows.Scan(&table)
		if err != nil {
			return nil, fmt.Errorf("%s: for query '%s', cannot read table. Error: %v", errorFuncName, query, err)
		}
		tables = append(tables, table)
	}
	return tables, nil
}
