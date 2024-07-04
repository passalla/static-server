package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

type Config struct {
	HostToDir map[string]string `yaml:"host"`
}

var (
	config     *Config
	configFile string
	configLock sync.RWMutex
)

func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var newConfig Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&newConfig); err != nil {
		return nil, err
	}

	return &newConfig, nil
}

func reloadConfig() error {
	newConfig, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	configLock.Lock()
	config = newConfig
	configLock.Unlock()

	fmt.Printf("Config reloaded: %+v\n", config)
	return nil
}

func main() {
	// Menentukan default values
	defaultPort := "8080"
	defaultStaticDir := "./"
	defaultConfigFile := "config.yml"
	defaultDebug := false
	defaultStrict := false

	// Mengambil argumen dari command line
	port := flag.String("port", defaultPort, "Port to run the server on")
	configFile = *flag.String("config", defaultConfigFile, "Configuration file for host to directory mapping")
	strict := flag.Bool("strict", defaultStrict, "Return 404 if file not found without searching all folders")
	isDebug := flag.Bool("debug", defaultDebug, "Is debug mode?")
	flag.Parse()

	strictMode := *strict

	if *isDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Memuat konfigurasi awal
	var err error
	config, err = loadConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Membuat router Gin
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		host := c.Request.Host

		configLock.RLock()
		dir, ok := config.HostToDir[host]
		configLock.RUnlock()

		fmt.Println(c.Request.URL.Path)

		if !ok {
			fs := http.FileServer(http.Dir(defaultStaticDir))
			if strictMode {
				requestedPath := filepath.Join(defaultStaticDir, c.Request.URL.Path)
				fileInfo, err := os.Stat(requestedPath)
				if os.IsNotExist(err) {
					c.String(http.StatusNotFound, "File not found")
					c.Abort()
					return
				}

				if fileInfo.IsDir() {
					indexPath := filepath.Join(requestedPath, "index.html")
					_, err := os.Stat(indexPath)
					if os.IsNotExist(err) {
						c.String(http.StatusNotFound, "File not exists")
						c.Abort()
						return
					}
				}
			}
			fs.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}
		fs := http.FileServer(http.Dir(dir))
		if strictMode {
			_, err := os.Stat(dir + c.Request.URL.Path)
			if os.IsNotExist(err) {
				c.String(http.StatusNotFound, "File not found")
				c.Abort()
				return
			}
		}
		fs.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	})

	r.GET("/reload", func(c *gin.Context) {
		if err := reloadConfig(); err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to reload config: %v", err))
			return
		}
		c.String(http.StatusOK, "Config reloaded successfully")
	})

	// Menjalankan server di port yang diberikan
	fmt.Println("Listening on: ", *port)
	r.Run(fmt.Sprintf(":%s", *port))
}
