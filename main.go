package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/kylelemons/go-gypsy/yaml"
)

var (
	configFile = kingpin.Flag("config", "Config file path").Short('c').String()
	mail       = kingpin.Flag("email", "User mail.").Required().Short('e').String()
	title      = kingpin.Flag("title", "Book title.").Required().Short('t').String()
	href       = kingpin.Flag("url", "Url of the book").Required().Short('u').String()
	expire     = kingpin.Flag("expire", "Date time book expire").Short('x').String()
	bookType   = kingpin.Flag("mediatype", "Media Type of book.\nopds: application/atom+xml;type=entry;profile=opds-catalog\nepub: application/epub+zip").Required().Short('m').String()
)

func main() {
	var claims jws.Claims
	var configFilePath string
	var issuer string
	var secret string

	kingpin.Version("0.0.1")
	kingpin.Parse()

	if *configFile == "" {
		configFilePath = "config.yaml"
	} else {
		configFilePath = *configFile
	}

	config, err := yaml.ReadFile(configFilePath)
	if err != nil {
		panic("can't read config file : " + configFilePath)
	}

	issuer, _ = config.Get("issuer")
	if issuer == "" {
		panic("can't get issuer from config file")
	}

	secret, _ = config.Get("secret")
	if secret == "" {
		panic("can't get secret from config file")
	}

	claims = make(jws.Claims)
	claims.Set("iss", issuer)
	claims.Set("sub", *mail)
	claims.Set("title", *title)
	claims.Set("href", *href)
	claims.Set("type", *bookType)

	if *expire != "" {
		claims.Set("expires", *expire)
		fmt.Println("add expire")
	}
	claims.SetIssuedAt(time.Now())
	jwt := jws.NewJWT(claims, crypto.SigningMethodHS256)
	buff, errJWT := jwt.Serialize([]byte(secret))
	if errJWT != nil {
		fmt.Println(errJWT)
	}
	reader := strings.NewReader(string(buff))

	request, err := http.NewRequest("POST", "https://aldiko.feedbooks.com/", reader) //Create request with JSON body

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		fmt.Println(err)
		return
	}

	if res.StatusCode != 200 {
		buffRes, _ := ioutil.ReadAll(res.Body)

		panic("error : \n" + string(buffRes))
	} else {
		fmt.Println("book send")
	}

}
