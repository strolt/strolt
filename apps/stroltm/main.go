package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/strolt/strolt/apps/stroltm/internal/api"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt"
)

func main() {
	log.Println("smanager")

	{
		s := time.Now()
		config, err := strolt.New("127.0.0.1:3333").GetConfig()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(time.Since(s))

		{
			log.Println("200:", config.IsCode(200))
			log.Println("404:", config.IsCode(404))

			data, err := json.Marshal(config)
			if err != nil {
				log.Println(err)
			}
			log.Println(string(data))
		}
	}

	{
		config, err := strolt.New("127.0.0.1:3333").GetPrune("e2e", "local", "restic-local1")
		if err != nil {
			log.Println(err)
			return
		}

		{
			log.Println("200:", config.IsCode(200))
			log.Println("500:", config.IsCode(500))

			data, err := json.Marshal(config)
			if err != nil {
				log.Println(err)
			}
			log.Println(string(data))
		}
	}

	api.Start()
}
