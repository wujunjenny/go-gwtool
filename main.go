// QueryAcStatus project main.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	//"github.com/fatih/color"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app      = kingpin.New("smgwtool", "A command-line monitor smgw server.")
	serverIP = app.Flag("server", "Server address.").Short('s').Default("127.0.0.1:6789").TCP()

	queryAccounts = app.Command("accounts", "Query all account names.")

	QAS       = app.Command("account-status", "query account status")
	QASName   = QAS.Arg("name", "query account name").Required().String()
	QASFormat = QAS.Flag("format", "format file").Short('f').File()

	QSTable     = app.Command("ac-convert", "query account convert table")
	QSTableName = QSTable.Arg("name", "query account name").Required().String()

	SETSP   = app.Command("setsp", "set sp info")
	SPFILES = SETSP.Arg("files", "json files for sp setting").Required().ExistingFiles()
	DelSP   = app.Command("delsp", "del sp")
	Delname = DelSP.Arg("name", "account name").Required().String()

	SETSECTION   = app.Command("setsection", "set section table from files")
	SECTIONFILES = SETSECTION.Arg("files", "json files for sections").Required().ExistingFiles()

	TESTROUTE   = app.Command("testroute", "test route")
	TESTDEST    = TESTROUTE.Arg("dest", "destaddress").Required().String()
	TESTORG     = TESTROUTE.Arg("org", "orgaddress").String()
	TESTAC      = TESTROUTE.Arg("ac", "account name").String()
	TESTSERVICE = TESTROUTE.Arg("service", "service code").String()
)

type FormatTag struct {
	Name        string `json:"name"`
	DisplayName string `json:"display"`
	Format      string `json:"format"`
}

func main() {

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case queryAccounts.FullCommand():
		QueryAccounts()
	case QAS.FullCommand():
		QueryAccountStatus()
	case SETSP.FullCommand():
		SetSpInfo()
	case DelSP.FullCommand():
		DelSp()
	case SETSECTION.FullCommand():
		SetSection()
	case TESTROUTE.FullCommand():
		TestRoute()
	}

}

func QueryAccounts() {
	url := "http://" + (*serverIP).String() + "/OmcManager.QueryAccounts"
	fmt.Println(url)
	rs, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	} else {
		buf, _ := ioutil.ReadAll(rs.Body)
		PrintJson(buf)
	}

}

func QueryAccountStatus() {
	url := "http://" + (*serverIP).String() + "/OmcManager.QueryAccountStatus"

	content := &struct {
		Name string `json:"accountName"`
	}{}

	content.Name = *QASName

	js, _ := json.Marshal(content)

	fmt.Println(string(js))
	for {

		select {
		case <-time.After(time.Second):
			bd := strings.NewReader(string(js))

			rs, err := http.Post(url, "", bd)
			if err != nil {
				fmt.Println(err)
			} else {
				buf, _ := ioutil.ReadAll(rs.Body)
				PrintJson(buf)
			}

		}
	}
}

func QueryServiceConvertTable() {
	url := "http://" + (*serverIP).String() + "/OmcManager.QueryServiceConvertTable"

	content := &struct {
		Name string `json:"accountName"`
	}{}

	content.Name = *QASName

	js, _ := json.Marshal(content)

	fmt.Println(string(js))
	bd := strings.NewReader(string(js))

	rs, err := http.Post(url, "", bd)
	if err != nil {
		fmt.Println(err)
	} else {
		buf, _ := ioutil.ReadAll(rs.Body)
		PrintJson(buf)
	}
}

func PrintJson(v []byte) {
	printJson(v, "")
}

func printJson(value []byte, prefix string) error {

	var objmap map[string]*json.RawMessage
	err := json.Unmarshal(value, &objmap)
	if err == nil {
		prefix1 := prefix + "  "
		fmt.Println("")
		fmt.Println(prefix1 +
			"{")
		for key, v := range objmap {
			fmt.Print(prefix1+key, ":")
			pe := printJson(*v, prefix1)
			if pe != nil {
				fmt.Println(string(*v))
			}
		}
		fmt.Println(prefix1 + "}")
		return nil
	}
	var objarray []*json.RawMessage
	err = json.Unmarshal(value, &objarray)

	if err == nil {
		prefix1 := prefix + "  "
		fmt.Println("")
		fmt.Println(prefix1 + "[")
		for key, v := range objarray {
			fmt.Print(prefix1+strconv.Itoa(key), ":")
			pe := printJson(*v, prefix1)
			if pe != nil {
				fmt.Println(string(*v))
			}

		}
		fmt.Println(prefix1 + "]")
		return nil
	}

	return err
}

func SetSpInfo() {
	url := "http://" + (*serverIP).String() + "/SetSPInfo"

	for _, file := range *SPFILES {
		bf, _ := ioutil.ReadFile(file)
		rs, err := http.Post(url, "", strings.NewReader(string(bf)))
		if err != nil {
			fmt.Println(err)
		} else {
			buf, _ := ioutil.ReadAll(rs.Body)
			p := &struct {
				Result  int    `json:"result"`
				Message string `json:"msg"`
			}{}

			err := json.Unmarshal(buf, p)

			fmt.Println(err)
			fmt.Println("result:", p.Result)
			fmt.Println("msg:", p.Message)
		}

	}

	//bd := strings.NewReader(string(`{"acname":"yy"}`))

}

func DelSp() {
	url := "http://" + (*serverIP).String() + "/RemoveSP"
	data := &struct {
		Acname string `json:"acname"`
		SpCode string `json:"spcode"`
	}{}
	data.Acname = *Delname
	buf, _ := json.Marshal(data)
	rs, err := http.Post(url, "", bytes.NewReader(buf))
	if err != nil {
		fmt.Println(err)
	} else {
		buf, _ := ioutil.ReadAll(rs.Body)
		p := &struct {
			Result  int    `json:"result"`
			Message string `json:"msg"`
		}{}

		err := json.Unmarshal(buf, p)

		fmt.Println(err)
		fmt.Println("result:", p.Result)
		fmt.Println("msg:", p.Message)
	}

}

func SetSection() {
	url := "http://" + (*serverIP).String() + "/SetSectionTable"
	var ivars []interface{}

	for _, file := range *SECTIONFILES {
		bf, _ := ioutil.ReadFile(file)
		var tmp []interface{}
		err := json.Unmarshal(bf, &tmp)
		if err != nil {
			fmt.Println("error json format ", file, err)
			return
		}
		ivars = append(ivars, tmp...)
	}
	bf, _ := json.Marshal(ivars)
	rs, err := http.Post(url, "", bytes.NewReader(bf))
	if err != nil {
		fmt.Println(err)
	} else {
		buf, _ := ioutil.ReadAll(rs.Body)
		p := &struct {
			Result  int    `json:"result"`
			Message string `json:"msg"`
		}{}

		err := json.Unmarshal(buf, p)

		fmt.Println(err)
		fmt.Println("result:", p.Result)
		fmt.Println("msg:", p.Message)
	}

}

func TestRoute() {
	fmt.Println(*TESTDEST, *TESTORG, *TESTAC)
	url := "http://" + (*serverIP).String() + "/TestRoute"
	var data = struct {
		Dest    string `json:"destAddr"`
		Org     string `json:"orgAddr"`
		Ac      string `json:"acname"`
		Service string `json:"service"`
	}{
		Dest:    *TESTDEST,
		Org:     *TESTORG,
		Ac:      *TESTAC,
		Service: *TESTSERVICE,
	}

	buf, _ := json.Marshal(&data)
	rs, err := http.Post(url, "", bytes.NewReader(buf))
	if err != nil {
		fmt.Println(err)
	} else {
		buf, _ := ioutil.ReadAll(rs.Body)
		p := &struct {
			Result  int    `json:"result"`
			Message string `json:"msg"`
		}{}

		err := json.Unmarshal(buf, p)

		fmt.Println(err)
		fmt.Println("result:", p.Result)
		fmt.Println("msg:", p.Message)
	}

}
