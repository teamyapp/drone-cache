package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/teamyapp/drone-cache/cache"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

func main() {
	cfg := cache.Config{}
	err := fromEnv(&cfg)
	if err != nil {
		log.Println(err)
	}

	ca, err := cache.New(cfg)
	if err != nil {
		log.Println(err)
	}

	err = ca.Execute()
	if err != nil {
		log.Println(err)
	}
}

func fromEnv(config interface{}) error {
	err := autoLoadEnv(".env")
	if err != nil {
		log.Println(err)
		return err
	}

	err = autoLoadEnv(".repo.env")
	if err != nil {
		log.Println(err)
		return err
	}

	err = envconfig.Process("", config)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func autoLoadEnv(fileName string) error {
	_, err := os.Stat(fileName)
	if err == nil {
		return godotenv.Load(fileName)
	} else if os.IsNotExist(err) {
		return nil
	} else {
		return err
	}
}
