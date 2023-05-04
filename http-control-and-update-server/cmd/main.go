package main

import (
	"fmt"
	"github.com/adamhoof/MedunkaOPBarcode2.0/config"
	"log"
	"net/http"
)

func main() {
	conf, err := config.LoadConfig("/home/adamhoof/MedunkaOPBarcode2.0/Config.json")
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("config ok")

	http.Handle("/", http.FileServer(http.Dir(conf.ControlAndUpdate.PathToUiStuff)))
	err = http.ListenAndServe(fmt.Sprintf("%s:%s", conf.ControlAndUpdate.Host, conf.ControlAndUpdate.Port), nil)
	if err != nil {
		log.Fatal("unable to start http server: ", err)
	}
}
