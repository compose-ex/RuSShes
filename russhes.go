package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	elastigo "github.com/mattbaird/elastigo/lib"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	feedurl = kingpin.Arg("feed", "RSS Feed URL").Required().String()
	esurl   = kingpin.Arg("es", "Elasticsearch URL").String()
	esindex = kingpin.Arg("index", "ES index name").Default("rss").String()
	estype  = kingpin.Arg("type", "ES type name").Default("rssitem").String()
)

// Rss2 defines the XML preamble describing the feed
type Rss2 struct {
	XMLName     xml.Name `xml:"rss"`
	Version     string   `xml:"version,attr"`
	Title       string   `xml:"channel>title"`
	Link        string   `xml:"channel>link"`
	Description string   `xml:"channel>description"`
	PubDate     string   `xml:"channel>pubDate"`
	ItemList    []Item   `xml:"channel>item"`
}

// Item describes the items within the feed
type Item struct {
	// Required
	Title       string        `xml:"title" json:"title"`
	Link        string        `xml:"link" json:"link"`
	Description template.HTML `xml:"description" json:"description"`
	Content     template.HTML `xml:"encoded" json:"content"`
	PubDate     string        `xml:"pubDate" json:"pubdate"`
	Comments    string        `xml:"comments" json:"comments"`
	GUID        string        `xml:"guid" json:"guid"`
	Category    []string      `xml:"category" json:"category"`
	Creator     string        `xml:"creator" json:"creator"`
}

func main() {
	kingpin.Parse()

	r := Rss2{}

	fmt.Println(*feedurl)

	response, err := http.DefaultClient.Get(*feedurl)

	if err != nil {
		log.Fatal(err)
	}
	xmlContent, err := ioutil.ReadAll(response.Body)
	err = xml.Unmarshal(xmlContent, &r)
	if err != nil {
		panic(err)
	}

	client := elastigo.NewConn()

	if *esurl != "" {
		client.SetFromUrl(*esurl)
	}

	for _, item := range r.ItemList {
		jsonValue, _ := json.MarshalIndent(item, "", "    ")
		if *esurl != "" {
			client.Index(*esindex, *estype, item.GUID, nil, jsonValue)
		} else {
			fmt.Println(string(jsonValue))
		}
	}
}
