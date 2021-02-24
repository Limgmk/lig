package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Limgmk/leedns/dns"
	D "github.com/miekg/dns"
	flag "github.com/spf13/pflag"
)

var (
	queryHost  string
	queryType  string
	queryNS    string
	queryClass string

	nsTypeUDP  bool
	nsTypeTCP  bool
	nsTypeTLS  bool
	nsTypeHTTP bool

	showWholeMsg bool
	showUsedTime bool
	showSecond   bool
	showJSONMsg  bool
	fmtJSONMsg   bool
	showShort    bool
)

func parseFlagByString(str []string) {
	var defArgs []string
	str = append([]string{""}, str...)
	defArgs, os.Args = os.Args, str
	flag.Parse()
	os.Args = defArgs
}

func bindFlag() {

	flag.StringVarP(&queryHost, "query", "q", "", "Host name or IP address to query")
	flag.StringVarP(&queryType, "type", "t", "A", "Type of the DNS record being queried (A, MX, NS...)")
	flag.StringVarP(&queryNS, "nameserver", "n", "", "Address of the nameserver to send packets to")
	flag.StringVar(&queryClass, "class", "IN", "Network class of the DNS record being queried (IN, CH, HS)")

	flag.BoolVarP(&nsTypeUDP, "udp", "U", false, "Use the DNS protocol over UDP")
	flag.BoolVarP(&nsTypeTCP, "tcp", "T", false, "Use the DNS protocol over TCP")
	flag.BoolVarP(&nsTypeTLS, "tls", "S", false, "Use the DNS-over-TLS protocol")
	flag.BoolVarP(&nsTypeHTTP, "http", "H", false, "Use the DNS-over-HTTPS protocol")

	flag.BoolVar(&showWholeMsg, "message", false, "Show whole DNS message instead of only answers")
	flag.BoolVarP(&showJSONMsg, "json", "J", false, "Display the output as JSON")
	flag.BoolVar(&fmtJSONMsg, "fmtjson", false, "Format JSON data when display the output as JSON")
	flag.BoolVar(&showSecond, "seconds", false, "Do not format durations, display them as seconds")
	flag.BoolVar(&showUsedTime, "time", false, "Print how long the response took to arrive")
	flag.BoolVarP(&showShort, "short", "1", false, "Short mode: display nothing but the first result")
}

func parseFlags() error {

	bindFlag()

	args := os.Args[1:]
	if len(args) == 0 {
		flag.Usage()
		return errors.New("")
	}

	if !strings.HasPrefix(args[0], "-") {
		queryHost = strings.TrimLeft(args[0], "-")
		args = args[1:]
	}

	for _, arg := range args {
		if strings.HasPrefix(arg, "@") {
			queryNS = arg
		}
		if num := D.StringToType[arg]; num != 0 {
			queryType = arg
		}
	}

	parseFlagByString(args)

	if queryHost == "" {
		flag.Usage()
		return errors.New("")
	}

	if queryNS == "" {
		conf, err := D.ClientConfigFromFile("/etc/resolv.conf")
		if err != nil {
			queryNS = "8.8.8.8"
			printNotes(fmt.Sprintf("coundn't get nameserver form /etc/resolv.cnf, will query from %s", queryNS))
		} else {
			queryNS = conf.Servers[0]
			printNotes(fmt.Sprintf("will use system default nameserver: %s", queryNS))
		}
	}

	if strings.HasPrefix(queryNS, "@") {
		queryNS = strings.TrimPrefix(queryNS, "@")
	}

	return nil
}

func main() {

	if err := parseFlags(); err != nil {
		return
	}

	var query = new(D.Msg)

	typeNum := D.StringToType[queryType]
	if typeNum == 0 {
		printError(fmt.Sprintf("Invalid query type: %s", queryType))
		return
	}

	query.SetQuestion(D.Fqdn(queryHost), typeNum)

	classNum := D.StringToClass[queryClass]
	if classNum == 0 {
		printError(fmt.Sprintf("Invalid query class: %s", queryClass))
		return
	}

	query.Question[0].Qclass = classNum

	log.Println(query.Question[0].Qclass)

	query.SetEdns0(4096, false)

	var client dns.Client
	var result *D.Msg
	var clientURL string

	switch {
	default:
		clientURL = fmt.Sprintf("udp://%s", queryNS)
	case nsTypeTCP:
		clientURL = fmt.Sprintf("tcp://%s", queryNS)
	case nsTypeTLS:
		clientURL = fmt.Sprintf("tls://%s", queryNS)
	case nsTypeHTTP:
		clientURL = queryNS
	}

	client, err := dns.NewClient(clientURL)
	if err != nil {
		flag.Usage()
		printError(fmt.Sprintf("create nameserver client failed: %s", err.Error()))
		return
	}

	if showUsedTime {
		startTime := time.Now().UnixNano()
		result, err = client.Exchange(query)
		endTime := time.Now().UnixNano()
		if err != nil {
			printError(fmt.Sprintf("query host from %s failed: %s", clientURL, err.Error()))
			return
		}
		printResult(result)
		fmt.Printf("Ran in %dms\n", (endTime-startTime)/1e6)
	} else {
		result, err = client.Exchange(query)
		if err != nil {
			printError(fmt.Sprintf("query host from %s failed: %s", clientURL, err.Error()))
			return
		}
		printResult(result)
	}
}
