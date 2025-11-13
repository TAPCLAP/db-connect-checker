package main

import (
	"fmt"
	"os"
	"time"

	"context"
	"net/url"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tapclap/db-connect-checker/pkg/metrics"
	"github.com/tapclap/db-connect-checker/pkg/mysqlcheck"
	"github.com/tapclap/db-connect-checker/pkg/util"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	dbType := util.GetEnvString("DB_TYPE", "mysql")
	exporterEnabled := util.GetEnvBool("EXPORTER", false)

	mysqlConfigs := util.GetAllMysqlConfigsFromEnvs()

	// mongodb
	mongoUri := os.Getenv("MONGODB_URI")
	if mongoUri == "" && dbType == "mongodb" {
		fmt.Fprintf(os.Stderr, "\"MONGODB_URI\" not set, but \"DB_TYPE\" is set \"mongodb\"")
		os.Exit(1)
	}

	tries := util.GetEnvNumber("TRIES", 10)

	if exporterEnabled {
		checkIntervalSeconds := util.GetEnvNumber("CHECK_INTERVAL", 30)
		checkInterval := time.Duration(checkIntervalSeconds) * time.Second

		mysqlExporter := metrics.NewMultiMySQLExporter(mysqlConfigs, checkInterval)

		mysqlExporter.Start()
		defer mysqlExporter.Stop()

		prometheus.MustRegister(mysqlExporter)

		http.Handle("/metrics", promhttp.Handler())

		port := util.GetEnvString("EXPORTER_PORT", "38080")
		addr := fmt.Sprintf(":%s", port)

		fmt.Printf("Starting metrics exporter on %s/metrics\n", addr)
		fmt.Printf("Check interval: %v\n", checkInterval)
		if err := http.ListenAndServe(addr, nil); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting HTTP server: %v\n", err)
			os.Exit(1)
		}
	} else {
		err := mysqlcheck.CheckConnections(mysqlConfigs, tries)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if dbType == "mongodb" {
			i := 1
			for i = 1; i <= tries; i += 1 {
				sleepS := 3*i + 1
				sleep := time.Duration(sleepS) * time.Second

				url, err := url.Parse(mongoUri)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: cannot get db from uri: %v\n", err)
					os.Exit(1)
				}

				dbName := url.Path
				if dbName[0] == '/' {
					dbName = url.Path[1:]
				}

				client, err := mongo.NewClient(options.Client().ApplyURI(mongoUri))
				if err != nil {
					fmt.Fprintf(os.Stderr, "Try (%d/%d) sleep %d seconds error mongodb connect to '%s': %v\n", i, tries, sleepS, url.Host, err)
					time.Sleep(sleep)
					continue
				}

				ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
				err = client.Connect(ctx)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error cannot create context %v\n", err)
					os.Exit(1)
				}

				_, err = client.Database(dbName).ListCollectionNames(ctx, bson.D{})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Try (%d/%d) sleep %d seconds error list collections: %v\n", i, tries, sleepS, err)
					time.Sleep(sleep)
					continue
				}

				fmt.Println("Connect success")
				break

			}

			if i == tries+1 {
				fmt.Fprintf(os.Stderr, "Connection attempts have failed")
				os.Exit(2)
			}
		}
	}

}
