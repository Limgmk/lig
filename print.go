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

type Response struct {
	*jsonDNS.Response
	Question []PatchQuestion `json:"Question"`
}

type PatchQuestion struct{
	*jsonDNS.Question
	Class uint16 `json:"class"`
}

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

func printTypeA(a *D.A) {

	aType := color.HiGreenString("A")

	aName := color.HiBlueString(a.Header().Name)

	aTTL := formatTTL(a.Header().Ttl)

	result := a.A.String()

	fmt.Printf("%s\t%s\t%s\t%s\n", aType, aName, aTTL, result)
}

func printTypeAAAA(a *D.AAAA) {

	aType := color.HiGreenString("AAAA")

	aName := color.HiBlueString(a.Header().Name)

	aTTL := formatTTL(a.Header().Ttl)

	result := a.AAAA.String()

	fmt.Printf("%s\t%s\t%s\t%s\n", aType, aName, aTTL, result)
}

func printTypeCNAME(a *D.CNAME) {

	aType := color.YellowString("CNAME")

	aName := color.HiBlueString(a.Header().Name)

	aTTL := formatTTL(a.Header().Ttl)

	result := formatCNAME(a.Target)

	fmt.Printf("%s\t%s\t%s\t%s\n", aType, aName, aTTL, result)
}

func printTypeTXT(a *D.TXT) {

	aType := color.YellowString("TXT")

	aName := color.HiBlueString(a.Header().Name)

	aTTL := formatTTL(a.Header().Ttl)

	result := formatTXT(a.Txt)

	fmt.Printf("%s\t%s\t%s\t%s\n", aType, aName, aTTL, result)
}

func printTypeMX(a *D.MX) {
	aType := color.HiCyanString("MX")

	aName := color.HiBlueString(a.Header().Name)

	aTTL := formatTTL(a.Header().Ttl)

	p := a.Preference

	result := fmt.Sprintf("\"%s\"", a.Mx)

	fmt.Printf("%s\t%s\t%s\t%d\t%s\n", aType, aName, aTTL, p, result)

}

func printColourful(m *D.Msg) {
	for _, a := range m.Answer {
		switch a2 := a.(type) {
		case *D.A:
			printTypeA(a2)
		case *D.AAAA:
			printTypeAAAA(a2)
		case *D.CNAME:
			printTypeCNAME(a2)
		case *D.TXT:
			printTypeTXT(a2)
		case *D.MX:
			printTypeMX(a2)
		}
	}
}

func printJSONMsg(m *D.Msg) {
	jsonMsg := jsonDNS.Marshal(m)

	jsonMsg2 := new(Response)
	jsonMsg2.Response = jsonMsg

	patchQ := new(PatchQuestion)
	patchQ.Question = &jsonMsg.Question[0]
	patchQ.Class = m.Question[0].Qclass

	jsonMsg2.Question = append(jsonMsg2.Question, *patchQ)

	var data []byte
	var err error
	if fmtJSONMsg {
		data, err = json.MarshalIndent(jsonMsg2, "", "\t")
	} else {
		data, err = json.Marshal(jsonMsg2)
	}
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
		case *D.MX:
			fmt.Println(fmt.Sprintf("%d \"%s\"", a2.Preference, a2.Mx))
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
}
