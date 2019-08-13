package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	bq "google.golang.org/api/bigquery/v2"
)

const (
	bqURL = "https://www.googleapis.com/auth/bigquery"
)

//BQFlags defines the subcomands for bigquery
var BQFlags = map[string]Flag{
	"project": Flag{
		Name:      "project",
		ShortName: "p",
		Usage:     "`project name`",
	},
	"dataset": Flag{
		Name:      "dataset",
		ShortName: "d",
		Usage:     "`Name of dataset to ship to`",
	},
	"table": Flag{
		Name:      "table",
		ShortName: "t",
		Usage:     "`Name of table to ship to`",
	},
	"serviceaccount": Flag{
		Name:      "serviceaccount",
		ShortName: "s",
		Usage:     "`path to json file with service account credentials`",
	},
}

type bqConfig struct {
	project string
	dataset string
	table   string
}
type biqQueryService struct {
	client *bq.Service
	config *bqConfig
}

//NewBqService conects to BigQuery and returns a new Stash with configuration
func NewBqService(project, dataset, table, credentials string) (Stash, error) {
	client, err := gAuth(credentials, bqURL)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return nil, err
	}
	s, err := bq.New(client)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	//Verify that the dataset and table exists
	ts := bq.NewTablesService(s).Get(project, dataset, table)
	if _, err := ts.Do(); err != nil {
		log.Fatalf("%+v", err)
	}
	return &biqQueryService{
		client: s,
		config: &bqConfig{
			project: project,
			table:   table,
			dataset: dataset,
		},
	}, nil
}

func (s *biqQueryService) Send(message string) error {
	return bqImpl(message, s.client, s.config)
}
func gAuth(serviceAccount string, api string) (*http.Client, error) {
	// Your credentials should be obtained from the Google
	// Developer Console (https://console.developers.google.com).
	// Navigate to your project, then see the "Credentials" page
	// under "APIs & Auth".
	// To create a service account client, click "Create new Client ID",
	// select "Service Account", and click "Create Client ID". A JSON
	// key file will then be downloaded to your computer.
	//"/home/oskar/go/src/zeromqClient/Formida-data-dump-b53ab3a70f14.json"
	data, err := ioutil.ReadFile(serviceAccount)
	if err != nil {
		return nil, err
	}
	// "https://www.googleapis.com/auth/bigquery"
	conf, err := google.JWTConfigFromJSON(data, api)
	if err != nil {
		return nil, err
	}
	// Initiate an http.Client. The following GET request will be
	// authorized and authenticated on the behalf of
	// your service account.
	client := conf.Client(oauth2.NoContext)
	client.Get("...")
	return client, nil
}

func bqImpl(msg string, s *bq.Service, pl *bqConfig) error {
	req, err := createRequest(msg, s, pl)
	if err != nil {
		return err
	}
	return doBigQueryCall(s.Tabledata.InsertAll(pl.project, pl.dataset, pl.table, req))
}

func createRequest(msg string, s *bq.Service, pl *bqConfig) (*bq.TableDataInsertAllRequest, error) {
	logMsg := make(map[string]bq.JsonValue, 0)
	err := json.Unmarshal([]byte(msg), &logMsg)
	if err != nil {
		return nil, err
	}
	rows := make([]*bq.TableDataInsertAllRequestRows, 0)
	row := &bq.TableDataInsertAllRequestRows{
		Json: logMsg,
	}
	rows = append(rows, row)
	return &bq.TableDataInsertAllRequest{
		Rows: rows,
	}, nil
}
func doBigQueryCall(call *bq.TabledataInsertAllCall) error {
	resp, err := call.Do()
	if err != nil {
		if strings.Contains(err.Error(), "503") {
			log.Printf("[ERROR] - retry in 3 seconds %+v", err)
			time.Sleep(3000 * time.Millisecond) //Sleep 3 seconds and retry
			resp, err = call.Do()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Printf("[ERROR] cant ship %+v", err)
		}
	} else {
		fmt.Printf("[INFO] Shiped log row with response: %d\n", resp.ServerResponse.HTTPStatusCode)
	}
	return nil
}
