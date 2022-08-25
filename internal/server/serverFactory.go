package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const envFile = ".env"

type Server struct {
	config *ServerConfig
	e      *echo.Echo
}

// server config used to create the server struct.
type ServerConfig struct {
	url        string
	port       int
	isSecure   bool
	serverCert string
	serverKey  string
}

func verifyPath(path string) error {
	_, err := os.Stat(path)
	return err
}

func loadConfig() (*ServerConfig, error) {
	var config ServerConfig
	if err := verifyPath(envFile); err != nil {
		file, err := os.Open(envFile)
		if err == nil {
			viper.ReadConfig(file)
		}
	}
	viper.AutomaticEnv()

	name, err := os.Hostname()
	if err != nil {
		// TODO: implement global logger
		log.Printf("error getting hostname, defaulting to localhost: %v", err)
		config.url = "127.0.0.1"
	} else {
		config.url = name
	}

	port := viper.GetInt("SERVER_PORT")
	if port == 0 {
		// TODO: implement global logger
		log.Print("no port found in environment, defaulting to 3000")
		config.port = 3000
	} else {
		config.port = port
	}

	tlsEnabled := viper.GetBool("SERVER_TLS")

	if tlsEnabled {
		serverKeyPath := viper.GetString("SERVER_KEY")
		if err := verifyPath(serverKeyPath); err != nil {
			return nil, err
		}
		config.serverKey = serverKeyPath

		serverCertPath := viper.GetString("SERVER_CERT")
		if err := verifyPath(serverCertPath); err != nil {
			return nil, err
		}
		config.serverCert = serverCertPath
		config.isSecure = true
	} else {
		config.isSecure = false
	}

	return &config, nil

}

func new(logger *zap.Logger) (*Server, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}
	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request",
				zap.String("URI", v.URI),
				zap.Int("status", v.Status),
			)

			return nil
		},
	}))
	// TODO: Add routes once implemented...
	return &Server{
		config: config,
		e:      e,
	}, nil
}

func Serve(logger *zap.Logger) {
	s, err := new(logger)
	if err != nil {
		log.Fatal(err)
	}
	url := fmt.Sprintf("%s:%d", s.config.url, s.config.port)
	log.Print(url)
	if s.config.isSecure {
		if err := s.e.StartTLS(url, s.config.serverCert, s.config.serverKey); err != http.ErrServerClosed {
			// TODO: implement global logger
			log.Fatal(err)
		}
	} else {
		if err := s.e.Start(url); err != http.ErrServerClosed {
			// TODO: implement global logger
			log.Fatal(err)
		}
	}
}
