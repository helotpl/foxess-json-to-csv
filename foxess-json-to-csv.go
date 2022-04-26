package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	Errno      int              `json:"errno"`
	ResultVars []ResultVariable `json:"result"`
}

type ResultVariable struct {
	Variable string `json:"variable"`
	Unit     string `json:"unit"`
	Name     string `json:"name"`
	//Data     []struct {
	//	Time  string  `json:"time"`
	//	Value float64 `json:"value"`
	//} `json:"data"`
	Data DataStore `json:"data"`
}

type DataStore struct {
	Data map[time.Time]float64
}

func (ds *DataStore) UnmarshalJSON(b []byte) error {
	if ds.Data == nil {
		ds.Data = make(map[time.Time]float64)
	}
	var tmp []struct {
		Time  string  `json:"time"`
		Value float64 `json:"value"`
	}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	for _, v := range tmp {
		t, err := time.Parse(timeFormat, v.Time)
		if err != nil {
			return err
		}
		ds.Data[t] = v.Value
	}
	return nil
}

const timeFormat = "2006-01-02 15:04:05 MST-0700"
const timeFormatShorter = "2006-01-02 15:04:05"

func (r *Result) GetTimes() []time.Time {
	ret := make(map[time.Time]bool)
	for _, rs := range r.ResultVars {
		for t := range rs.Data.Data {
			ret[t] = true
		}
	}

	var ret2 []time.Time
	for t := range ret {
		ret2 = append(ret2, t)
	}
	sort.Slice(ret2, func(i, j int) bool {
		return ret2[i].Before(ret2[j])
	})
	return ret2
}

func (r *Result) GetRow(t time.Time) string {
	var values []string

	values = append(values, t.Format(timeFormatShorter))
	for _, rv := range r.ResultVars {
		if value, ok := rv.Data.Data[t]; ok {
			strvalue := strconv.FormatFloat(value, 'f', -1, 64)
			strvalue = strings.ReplaceAll(strvalue, ".", ",")
			values = append(values, strvalue)
		} else {
			values = append(values, "")
		}
	}
	return strings.Join(values, ";")
}

func (r *Result) GetHeaders() string {
	var values []string
	values = append(values, "time")
	for _, rv := range r.ResultVars {
		values = append(values, fmt.Sprintf("%s[%s]", rv.Variable, rv.Unit))
	}
	return strings.Join(values, ";")
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatalln("first command line argument is a input file name")
	}

	infile := flag.Args()[0]
	in, err := os.Open(infile)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		err = in.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	unparsed, err := ioutil.ReadAll(in)
	if err != nil {
		log.Fatalln(err)
	}

	var result Result
	err = json.Unmarshal(unparsed, &result)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(result.GetHeaders())
	for _, t := range result.GetTimes() {
		fmt.Println(result.GetRow(t))
	}

}
