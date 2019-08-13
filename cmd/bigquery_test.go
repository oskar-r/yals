package cmd

import (
	"context"
	"reflect"
	"testing"

	bq "google.golang.org/api/bigquery/v2"
)

func Test_createRequest(t *testing.T) {
	type args struct {
		msg string
		s   *bq.Service
		pl  *bqConfig
	}
	s, err := bq.NewService(context.Background())
	if err != nil {
		t.Fatalf("%+v", err)
	}
	tests := []struct {
		name    string
		args    args
		want    *bq.TableDataInsertAllRequest
		wantErr bool
	}{
		{
			"TEST_1",
			args{
				msg: `{"my": "test"}`,
				s:   s,
				pl:  &bqConfig{},
			},
			&bq.TableDataInsertAllRequest{
				Rows: []*bq.TableDataInsertAllRequestRows{
					&bq.TableDataInsertAllRequestRows{
						Json: map[string]bq.JsonValue{
							"my": "test",
						},
					},
				},
			},
			false,
		},
		{
			"TEST_2",
			args{
				msg: `"my": "test"`,
				s:   s,
				pl:  &bqConfig{},
			},
			nil,
			true,
		},
	}
	/*

			"error": bq.JsonValue{"test"},
		}
	*/
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createRequest(tt.args.msg, tt.args.s, tt.args.pl)
			if (err != nil) != tt.wantErr {
				t.Errorf("createRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createRequest() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
