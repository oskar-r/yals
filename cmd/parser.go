package cmd

type Parser interface {
	Parse(logRow string) (string, error)
}

var parserImpl Parser

func SetupParser(p Parser) {
	parserImpl = p
}
func Parse(logRow string) (string, error) {
	return parserImpl.Parse(logRow)
}
