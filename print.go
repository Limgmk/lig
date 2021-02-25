package main

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	jsonDNS "github.com/m13253/dns-over-https/json-dns"
	D "github.com/miekg/dns"
	"strconv"
	"strings"
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

func formatSecond(sec uint32) string {

	if sec == 0 {
		return "0s"
	}

	s := sec % 60
	m := (sec / 60) % 60
	h := (sec / (60 * 60)) % 24
	d := sec / (60 * 60 * 24)

	var result string
	tmp := fmt.Sprintf("%sd", strconv.FormatUint(uint64(d), 10))
	if d == 0 {
		tmp = ""
	}
	result += tmp

	tmp = fmt.Sprintf("%sh", strconv.FormatUint(uint64(h), 10))
	if h == 0 {
		if result == "" {
			tmp = ""
		}
	} else if h < 10 {
		if result != "" {
			tmp = "0" + tmp
		}
	}
	result += tmp

	tmp = fmt.Sprintf("%sm", strconv.FormatUint(uint64(m), 10))
	if m == 0 {
		if result != "" {
			tmp = "0" + tmp
		} else {
			tmp = ""
		}
	} else if m < 10 {
		if result != "" {
			tmp = "0" + tmp
		}
	}
	result += tmp

	tmp = fmt.Sprintf("%ss", strconv.FormatUint(uint64(s), 10))
	if s == 0 {
		if result != "" {
			tmp = "0" + tmp
		} else {
			tmp = ""
		}
	} else if s < 10 {
		if result != "" {
			tmp = "0" + tmp
		}
	}
	result += tmp

	return result
}

func formatTTL(ttl uint32) string {

	if showSecond {
		return fmt.Sprintf("%ss", strconv.FormatUint(uint64(ttl), 10))
	}

	return formatSecond(ttl)
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

	typeStr := color.HiGreenString("A")

	host := color.HiBlueString(a.Header().Name)

	ttl := formatTTL(a.Header().Ttl)

	result := a.A.String()

	fmt.Printf("%s\t%s\t%s\t%s\n", typeStr, host, ttl, result)
}

func printTypeAAAA(a *D.AAAA) {

	typeStr := color.HiGreenString("AAAA")

	host := color.HiBlueString(a.Header().Name)

	ttl := formatTTL(a.Header().Ttl)

	result := a.AAAA.String()

	fmt.Printf("%s\t%s\t%s\t%s\n", typeStr, host, ttl, result)
}

func printTypeCNAME(a *D.CNAME) {

	typeStr := color.YellowString("CNAME")

	host := color.HiBlueString(a.Header().Name)

	ttl := formatTTL(a.Header().Ttl)

	result := formatCNAME(a.Target)

	fmt.Printf("%s\t%s\t%s\t%s\n", typeStr, host, ttl, result)
}

func printTypeTXT(a *D.TXT) {

	typeStr := color.YellowString("TXT")

	host := color.HiBlueString(a.Header().Name)

	ttl := formatTTL(a.Header().Ttl)

	result := formatTXT(a.Txt)

	fmt.Printf("%s\t%s\t%s\t%s\n", typeStr, host, ttl, result)
}

func printTypeMX(a *D.MX) {
	typeStr := color.HiCyanString("MX")

	host := color.HiBlueString(a.Header().Name)

	ttl := formatTTL(a.Header().Ttl)

	p := a.Preference

	result := fmt.Sprintf("\"%s\"", a.Mx)

	fmt.Printf("%s\t%s\t%s\t%d\t%s\n", typeStr, host, ttl, p, result)
}

func printTypeSOA(ns *D.SOA, isAnsware bool) {

	typeStr := color.HiMagentaString("SOA")
	
	host := color.HiBlueString(ns.Header().Name)
	
	ttl := formatTTL(ns.Header().Ttl)

	var authFlag string
	if !isAnsware {
		authFlag = color.CyanString("A ")
	} else {
		authFlag = ""
	}

	fmt.Printf("%s %s %s %s\"%s\" \"%s\" %d %s %s %s %s \n", typeStr, host, ttl,
		authFlag, ns.Ns, ns.Mbox, ns.Serial, formatSecond(ns.Refresh), formatSecond(ns.Retry),
		formatSecond(ns.Expire), formatSecond(ns.Minttl))
}

func printDefault(m *D.Msg) {
	if len(m.Answer) == 0 {
		for _, ns := range m.Ns {
			switch ns2 := ns.(type) {
			case  *D.SOA:
				printTypeSOA(ns2, false)
			}
		}
		return
	}
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
		case *D.SOA:
			printTypeSOA(a2, true)
		default:
			fmt.Println(a2.String())
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

	if showJSONMsg || fmtJSONMsg {
		printJSONMsg(m)
		return
	}
	if showWholeMsg {
		fmt.Println(m)
		return
	}

	printDefault(m)
}
