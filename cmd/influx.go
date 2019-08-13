package cmd

import (
	"encoding/json"
	"errors"
	"log"
	"regexp"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
)

var InfluxFlags = map[string]Flag{
	"database": Flag{
		Name:      "database",
		ShortName: "d",
		Usage:     "`Name of database to ship to`",
	},
	"serie": Flag{
		Name:      "serie",
		ShortName: "s",
		Usage:     "`Name of time serie to ship to`",
	},
	"connection": Flag{
		Name:      "connection",
		ShortName: "c",
		Usage:     "`Connection string to influx DB`",
	},
}

type influxConf struct {
	database string
	serie    string
}
type influxService struct {
	client *influx.Client
	config *influxConf
}

//NewInfluxService conects to influxDB and returns a new Stash
func NewInfluxService(database, serie, credentials string) (Stash, error) {
	c, err := influxClient(credentials)
	if err != nil {
		return nil, err
	}
	return &influxService{
		client: &c,
		config: &influxConf{
			database: database,
			serie:    serie,
		},
	}, nil
}

//Send implementes the send interface for sending to influx DB
func (s *influxService) Send(message string) error {
	pt, err := influxCreatePoint(message, s.config)
	cfg := influx.BatchPointsConfig{
		Database:  s.config.database,
		Precision: "s",
	}
	bp, err := influx.NewBatchPoints(cfg)
	if err != nil {
		return err
	}
	bp.AddPoints([]*influx.Point{pt})

	return (*s.client).Write(bp)
}
func influxClient(influxConStr string) (influx.Client, error) {
	cd, err := parseConnectionString(influxConStr)
	if err != nil {
		log.Printf("[ERROR] %+v", err)
	}
	return influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     cd["server"],
		Username: cd["username"],
		Password: cd["password"],
	})
}

func influxCreatePoint(msg string, pl *influxConf) (*influx.Point, error) {
	fields := make(map[string]interface{}, 0)
	err := json.Unmarshal([]byte(msg), &fields)
	if err != nil {
		return nil, err
	}
	tags := make(map[string]string, 0)
	now := time.Now()
	for k, v := range fields {
		switch v.(type) {
		case string:
			tags[k] = v.(string)
			delete(fields, k)
		case time.Time:
			now = v.(time.Time)
		case *time.Time:
			now = *(v.(*time.Time))
		}
	}

	return influx.NewPoint(pl.database, tags, fields, now)
}

func parseConnectionString(connection string) (map[string]string, error) {
	var re = regexp.MustCompile(`(?P<username>(.*)):(?P<password>(.*))@(?P<server>(http:\/\/.*:[0-9]{2,}))\/(?P<database>(.*))`)

	match := re.FindStringSubmatch(connection)
	result := make(map[string]string)

	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	if len(result) != 4 {
		return nil, errors.New("Could not parse influx DB connection string " + connection)
	}
	return result, nil
}
