package main

import (
	"log"

	"github.com/strolt/strolt/apps/stroltm/internal/api"
	"github.com/strolt/strolt/apps/stroltm/internal/config"
)

func main() {
	if err := config.Scan(); err != nil {
		log.Println(err)
		return
	}

	log.Println(config.Get())

	log.Println("smanager")

	// {
	// 	s := time.Now()
	// 	config, err := strolt.New("127.0.0.1:3333").GetConfig()
	// 	if err != nil {
	// 		log.Println(err)
	// 		return
	// 	}
	// 	log.Println(time.Since(s))

	// 	{
	// 		data, err := json.Marshal(config)
	// 		if err != nil {
	// 			log.Println(err)
	// 		}
	// 		log.Println(string(data))
	// 	}
	// }

	// {
	// 	config, err := strolt.New("127.0.0.1:3333").GetPrune("e2e", "local", "restic-local1")
	// 	if err != nil {
	// 		log.Println(err)
	// 		return
	// 	}

	// 	{
	// 		data, err := json.Marshal(config)
	// 		if err != nil {
	// 			log.Println(err)
	// 		}
	// 		log.Println(string(data))
	// 	}
	// }

	api.Start()
}
