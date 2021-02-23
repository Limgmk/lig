package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
	jsonDNS "github.com/m13253/dns-over-https/json-dns"
	D "github.com/miekg/dns"
)

func printError(err string) {
	header := color.HiRedString("Error")
	fmt.Printf("%s: %s\n", header, err)
}

func printNotes(notes string) {
	header := color.YellowString("Notes")
	fmt.Printf("%s: %s\n", header, notes)
}

func formatTTL(ttl uint32) string {
	var str string

	if showSecond {
		return strconv.FormatUint(uint64(ttl), 10)
	}

	min := ttl / 60
	second := ttl % 60

	if min == 0 {
		str = fmt.Sprintf("%ds", second)
	} else {
		str = fmt.Sprintf("%dm%ds", min, second)
	}
	return str
}

func formatCNAME(cname string) string {
	return fmt.Sprintf("\"%s\"", cname)
}

func formatTXT(txt []string) string {
	for _, t := range txt {
		txt = append(txt, fmt.Sprintf("\"%s\"", t))
	}
	return strings.Join(txt, " ")
}

func printTypeA(answare D.RR) {
	a := answare.(*D.A)

	aType := color.HiGreenString("A")

	aName := color.HiBlueString(a.Header().Name)

	aTTL := formatTTL(a.Header().Ttl)

	result := a.A.String()

	fmt.Printf("%s\t%s\t%s\t%s\n", aType, aName, aTTL, result)
}

func printTypeAAAA(answare D.RR) {
	a := answare.(*D.AAAA)

	aType := color.HiGreenString("AAAA")

	aName := color.HiBlueString(a.Header().Name)

	aTTL := formatTTL(a.Header().Ttl)

	result := a.AAAA.String()

	fmt.Printf("%s\t%s\t%s\t%s\n", aType, aName, aTTL, result)
}

func printTypeCNAME(answare D.RR) {
	a := answare.(*D.CNAME)

	aType := color.YellowString("CNAME")

	aName := color.HiBlueString(a.Header().Name)

	aTTL := formatTTL(a.Header().Ttl)

	result := formatCNAME(a.Target)

	fmt.Printf("%s\t%s\t%s\t%s\n", aType, aName, aTTL, result)
}

func printTypeTXT(answare D.RR) {
	a := answare.(*D.TXT)

	aType := color.YellowString("TXT")

	aName := color.HiBlueString(a.Header().Name)

	aTTL := formatTTL(a.Header().Ttl)

	result := formatTXT(a.Txt)

	fmt.Printf("%s\t%s\t%s\t%s\n", aType, aName, aTTL, result)
}

func printColourful(m *D.Msg) {
	for _, a := range m.Answer {
		switch a.(type) {
		case *D.A:
			printTypeA(a)
		case *D.AAAA:
			printTypeAAAA(a)
		case *D.CNAME:
			printTypeCNAME(a)
		case *D.TXT:
			printTypeTXT(a)
		}
	}
}

func printJSONMsg(m *D.Msg) {
	jsonMsg := jsonDNS.Marshal(m)
	data, err := json.Marshal(jsonMsg)
	if err != nil {
		return
	}
	fmt.Println(string(data))
}

func printAnswerSimple(m *D.Msg) {
	for _, a := range m.Answer {
		switch a2 := a.(type) {
		case *D.A:
			fmt.Println(a2.A.String())
		case *D.AAAA:
			fmt.Println(a2.AAAA.String())
		case *D.CNAME:
			fmt.Println(formatCNAME(a2.Target))
		case *D.TXT:
			fmt.Println(formatTXT(a2.Txt))
		}
	}
}

func printResult(m *D.Msg) {

	if showShort {
		printAnswerSimple(m)
		return
	}

	if showJSONMsg {
		printJSONMsg(m)
		return
	}
	if showWholeMsg {
		fmt.Println(m)
		return
	}

	printColourful(m)

	//for _, ex := range m.Extra {
	//	switch ee := ex.(type) {
	//	case *D.OPT:
	//		//log.Println("is opt")
	//		//ee.
	//		fmt.Printf("OPT\t \t \t \t \t%d %d %d %d %s\n", ee.UDPSize(), ee.Version(), ee.ExtendedRcode(), ee.Version(), ee.)
	//	}
	//}
}
