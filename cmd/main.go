package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

type Config struct {
	HostToDir map[string]string `yaml:"host"`
}

const (
	Version = "1.0.4"
)

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

	defaultProxy := []string{"127.0.0.1"}

	// Mengambil argumen dari command line
	port := flag.String("port", defaultPort, "Port to run the server on")
	configFile = *flag.String("config", defaultConfigFile, "Configuration file for host to directory mapping")
	strict := flag.Bool("strict", defaultStrict, "Return 404 if file not found without searching all folders")
	isDebug := flag.Bool("debug", defaultDebug, "Is debug mode?")
	proxy := flag.String("proxy", "", "string Proxies separated by commas")

	flag.Parse()

	proxies := defaultProxy
	if *proxy != "" {
		proxies = strings.Split(*proxy, ",")
	}

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
	r := gin.New()
	r.SetTrustedProxies(proxies)
	// Middleware untuk mencatat log dengan IP asli klien dan hostname
	r.Use(gin.Recovery()) // Menambahkan middleware recovery bawaan gin
	r.Use(func(c *gin.Context) {
		startTime := time.Now()
		c.Next()

		// Mendapatkan alamat IP asli klien
		clientIP := c.ClientIP()
		// Mendapatkan hostname dari permintaan
		host := c.Request.Host
		// Mendapatkan durasi waktu proses
		duration := time.Since(startTime)

		// Mencatat log
		fmt.Printf("[LOG] %d | %s | %s | %s%s | %v\n",
			c.Writer.Status(),
			clientIP,
			c.Request.Method,
			host,
			c.Request.RequestURI,
			duration)
	})

	r.GET("/reload", func(c *gin.Context) {
		if err := reloadConfig(); err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to reload config: %v", err))
			return
		}
		c.String(http.StatusOK, "Config reloaded successfully")
	})

	r.Use(func(c *gin.Context) {
		host := c.Request.Host

		configLock.RLock()
		dir, ok := config.HostToDir[host]
		configLock.RUnlock()

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

	// Menjalankan server di port yang diberikan
	fmt.Printf("Static Server Version %s Listening on: %s\n", Version, *port)
	r.Run(fmt.Sprintf(":%s", *port))
}
