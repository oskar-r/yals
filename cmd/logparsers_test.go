package cmd

import (
	"regexp"
	"testing"
)

func Test_parserConf_Parse(t *testing.T) {

	type args struct {
		text       string
		re         string
		dateLayout string
		parser     func(string) (string, error)
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"Test_1",
			args{
				text: `2019/08/13 21:27:46 POST  	|/user/login	|200	|13.6 ms	|size:168 B	[request_id:010d25f9-d949-4dee-990a-a46d300fa069	user:-1	role:not set]`,
				re:         `(?m)(?P<date>20\d{2}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2})\s(?P<method>[A-Z]{1,10})\s{0,}\|(?P<route>[A-Za-z0-9=?&/_]{0,})\s{0,}\|(?P<respose>[0-9]{3})\s{0,}\|(?P<resptime>[0-9\.]{0,}).*\s{0,}\|size:(?P<size>[0-9]{0,})\s{0,}B\s{0,}\[request_id:(?P<req_id>[0-9a-z-]{0,})\s{0,}user:(?P<subject>[0-9-]{0,})\s{0,}role:(?P<role>[a-z\s]{0,})]`,
				dateLayout: "2006/01/02 15:04:05",
				parser:     nil,
			},
			`{"date":"2019/08/13 21:27:46","method":"POST","req_id":"010d25f9-d949-4dee-990a-a46d300fa069","respose":"200","resptime":"13.6","role":"not set","route":"/user/login","size":"168","subject":"-1"}`,
			false,
		},
		{
			"Test_with_types",
			args{
				text: `2019/08/13 21:27:46 POST  	|/user/login	|200	|13.6 ms	|size:168 B	[request_id:010d25f9-d949-4dee-990a-a46d300fa069	user:-1	role:not set]`,
				re:         `(?m)(?P<date99t>20\d{2}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2})\s(?P<method>[A-Z]{1,10})\s{0,}\|(?P<route>[A-Za-z0-9=?&/_]{0,})\s{0,}\|(?P<respose99i>[0-9]{3})\s{0,}\|(?P<resptime99f>[0-9\.]{0,}).*\s{0,}\|size:(?P<size99i>[0-9]{0,})\s{0,}B\s{0,}\[request_id:(?P<req_id>[0-9a-z-]{0,})\s{0,}user:(?P<subject99i>[0-9-]{0,})\s{0,}role:(?P<role>[a-z\s]{0,})]`,
				dateLayout: "2006/01/02 15:04:05",
				parser:     nil,
			},
			`{"date":"2019-08-13T21:27:46Z","method":"POST","req_id":"010d25f9-d949-4dee-990a-a46d300fa069","respose":200,"resptime":13.6,"role":"not set","route":"/user/login","size":168,"subject":-1}`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &parserConf{
				re:         regexp.MustCompile(tt.args.re),
				dateLayout: tt.args.dateLayout,
				parser:     tt.args.parser,
			}
			got, err := p.Parse(tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("parserConf.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parserConf.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
