package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"os"
)

type GlobalConfiguration struct {
	API     APIConfiguration
	Logging LoggingConfig     `envconfig:"LOG"`
	CORS    CORSConfiguration `json:"cors"`
}

func (c *GlobalConfiguration) Validate() error {
	validates := []interface {
		Validate() error
	}{
		&c.API,
		&c.Logging,
	}

	for _, v := range validates {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c *GlobalConfiguration) ApplyDefaults() error {
	return nil
}

func LoadGlobal(fileName string) (*GlobalConfiguration, error) {
	if err := loadEnvironment(fileName); err != nil {
		return nil, err
	}
	conf := new(GlobalConfiguration)

	if err := envconfig.Process("", conf); err != nil {
		return nil, err
	}
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	if err := conf.ApplyDefaults(); err != nil {
		return nil, err
	}

	return conf, nil
}

func loadEnvironment(fileName string) error {
	var err error
	if fileName != "" {
		err = godotenv.Overload(fileName)
	} else {
		err = godotenv.Load()
		// handle if .env file does not exist, this is OK
		if os.IsNotExist(err) {
			return nil
		}
	}
	return err
}

type APIConfiguration struct {
	Host     string
	Port     string `envconfig:"PORT" default:"8000"`
	Endpoint string
}

func (a *APIConfiguration) Validate() error {
	return nil
}
