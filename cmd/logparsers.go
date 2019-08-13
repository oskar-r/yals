package cmd

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"regexp"
	"strconv"
	"time"
)

const (
	formidaRe = `(?m)(?P<date>20\d{2}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2})\s(?P<method>[A-Z]{1,10})\s{0,}\|(?P<route>[A-Za-z0-9=?&/_]{0,})\s{0,}\|(?P<respose>[0-9]{3})\s{0,}\|(?P<resptime>[0-9\.]{0,}).*\s{0,}\|size:(?P<size>[0-9]{0,})\s{0,}B\s{0,}\[request_id:(?P<req_id>[0-9a-z-]{0,})\s{0,}user:(?P<subject>[0-9-]{0,})\s{0,}role:(?P<role>[a-z\s]{0,})]`
	//Regex for ngninx logs
	nginxRe = `(?P<ipaddress>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) - - \[(?P<date>\d{1,2}\/\w{3}\/\d{4}:\d{2}:\d{2}:\d{2} [\+|\-]\d{4})\] \[(?P<timestamp>[0-9\.]{14})\] \"(?P<method>GET|POST|PUT|DELETE) (?P<url>\/.*) HTTP\/[0-9\.]{1,3}" (?P<statuscode>\d{1,3})  (?P<bodysize>[0-9]{1,20}) (?P<resptime>[0-9\.]{5,7})`
)

var ref = regexp.MustCompile(formidaRe)
var ren = regexp.MustCompile(nginxRe)

type formidaPayload struct {
	Date         time.Time `json:"date,omitempty"`
	Method       string    `json:"method,omitempty"`
	Route        string    `json:"route,omitempty"`
	ResponseCode int       `json:"response_code,omitempty"`
	ExecTime     float64   `json:"exec_time,omitempty"`
	Size         int       `json:"size,omitempty"`
	RequestID    string    `json:"request_id,omitempty"`
	SubjectID    int       `json:"subject_id,omitempty"`
	Role         string    `json:"role,omitempty"`
}

func formidaParser(text string) (string, error) {
	matches := ref.FindStringSubmatch(text)
	if len(matches) != 10 {
		return "", errors.New("Matches is less than 10 act.size = " + strconv.Itoa(len(matches)) + " input: " + text)
	}

	nt, err := time.Parse("2006/01/02 15:04:05", matches[1])
	if err != nil {
		return "", err
	}

	var rc, size, sub int
	var respTime float64
	if rc, err = strconv.Atoi(matches[4]); err != nil {
		return "", errors.New("Can't convert response code err = " + err.Error())
	}
	if size, err = strconv.Atoi(matches[6]); err != nil {
		return "", errors.New("Can't convert size err = " + err.Error())
	}
	if sub, err = strconv.Atoi(matches[8]); err != nil {
		return "", errors.New("Can't convert subject err = " + err.Error())
	}
	if respTime, err = strconv.ParseFloat(matches[5], 32); err != nil {
		return "", errors.New("Can't convert resp time err = " + err.Error())
	}
	fp := formidaPayload{
		Date:         nt,
		Method:       matches[2],
		Route:        matches[3],
		ResponseCode: rc,
		Size:         size,
		ExecTime:     respTime,
		RequestID:    matches[7],
		Role:         matches[9],
	}
	if sub != -1 {
		fp.SubjectID = sub
	}
	b, err := json.Marshal(&fp)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type nginxPayload struct {
	IPaddress  string    `json:"ipaddress,omitempty"`
	Date       time.Time `json:"date,omitempty"`
	Method     string    `json:"method,omitempty"`
	URL        string    `json:"url,omitempty"`
	StatusCode int       `json:"statuscode,omitempty"`
	BodySize   int       `json:"bodysize,omitempty"`
	RespTime   float64   `json:"resptime,omitempty"`
	Timestamp  int       `json:"timestamp,omitempty"`
}

func nginxParser(text string) (string, error) {

	matches := ren.FindStringSubmatch(text)
	if len(matches) != 9 {
		return "", errors.New("Matches is less than 10 act.size = " + strconv.Itoa(len(matches)) + " input: " + text)
	}
	np := nginxPayload{
		IPaddress: matches[1],
		Method:    matches[3],
		URL:       matches[4],
	}

	newTime, err := time.Parse("02/Jan/2006:15:04:05 -0700", matches[2])
	if err != nil {
		log.Printf("[ERROR] parse time %+v", err)
	}

	np.Date = newTime
	np.StatusCode, _ = strconv.Atoi(matches[6])
	np.BodySize, _ = strconv.Atoi(matches[7])
	tmp, err := strconv.ParseFloat(matches[8], 64)
	if err == nil {
		np.RespTime = tmp * 1000
	}
	b, err := json.Marshal(&np)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type parserConf struct {
	re         *regexp.Regexp
	dateLayout string
	parser     func(string) (string, error)
}

func (p *parserConf) genericParser(text string) (string, error) {
	matches := p.re.FindStringSubmatch(text)
	if len(matches) == 0 {
		return "", errors.New("No matches")
	}

	temp := make(map[string]interface{}, 0)

	for k, v := range matches {
		if p.re.SubexpNames()[k] != "" {
			t := strings.Split(p.re.SubexpNames()[k], "99")
			if len(t) == 2 {
				switch t[1] {
				case "i":
					temp[t[0]], _ = strconv.Atoi(v)
				case "f":
					temp[t[0]], _ = strconv.ParseFloat(v, 64)
				case "t":
					temp[t[0]], _ = time.Parse(p.dateLayout, v)
				default:
					temp[t[0]] = v
				}
			} else {
				temp[t[0]] = v
			}
		}
	}
	b, err := json.Marshal(&temp)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

//NewParser returns a Parser that can be used to parse logfiles
func NewParser(regex, dateLayout string, parser func(string) (string, error)) Parser {
	return &parserConf{
		re:         regexp.MustCompile(regex),
		dateLayout: dateLayout,
		parser:     parser,
	}
}

func (p *parserConf) Parse(text string) (string, error) {
	if p.parser != nil {
		return p.parser(text)
	}
	return p.genericParser(text)
}
