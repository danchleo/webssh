package main

import (
	"os"
	"path"
	"strings"
	"webssh/webssh"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Config struct {
	env map[string][]string
}

func (c *Config) LoadEnv() {
	c.env = map[string][]string{}
	env := os.Environ()
	for i := 0; i < len(env); i++ {
		pair := strings.SplitN(env[i], "=", 2)
		k := pair[0]
		v := pair[1]
		if _, ok := c.env[k]; !ok {
			c.env[k] = []string{}
		}
		c.env[k] = append(c.env[k], v)
	}
}

func (c *Config) getValues(key string, defVal []string) []string {
	if v, ok := c.env[key]; ok {
		return v
	}
	return defVal
}

func (c *Config) getValue(key string, defVal string) string {
	return c.getValues(key, []string{defVal})[0]
}

func (c *Config) RemoveAddr() string {
	return c.getValue("SSH_REMOTE_ADDR", "127.0.0.1:22")
}

func (c *Config) User() string {
	return c.getValue("SSH_USER", "root")
}

func (c *Config) Password() string {
	return c.getValue("SSH_PASSWORD", "root")
}

func (c *Config) AllowOrigins() []string {
	var origins []string
	cfgOrigins := c.getValues("HTTP_ALLOW_ORIGINS", []string{"*"})
	for i := 0; i < len(cfgOrigins); i++ {
		opt := strings.Split(cfgOrigins[i], ",")
		origins = append(origins, opt...)
	}
	return origins
}

func (c *Config) Addr() string {
	return c.Host() + ":" + c.Port()
}

func (c *Config) Host() string {
	return c.getValue("HTTP_HOST", "")
}

func (c *Config) Port() string {
	return c.getValue("HTTP_PORT", "8080")
}

func (c *Config) StaticDir() string {
	return c.getValue("HTTP_PUBLIC", "./frontend/dist")
}

func main() {
	cfg := Config{}
	cfg.LoadEnv()

	r := gin.Default()

	//跨域设置
	r.Use(
		cors.New(
			cors.Config{
				AllowOrigins:     cfg.AllowOrigins(),
				AllowHeaders:     []string{"Origin"},
				ExposeHeaders:    []string{"Content-Length"},
				AllowCredentials: true,
			},
		),
	)

	sshCfg := &webssh.WebSSHConfig{
		Record:     false,
		RemoteAddr: cfg.RemoveAddr(),
		User:       cfg.User(),
		Password:   cfg.Password(),
		AuthModel:  webssh.PASSWORD,
	}

	handle := webssh.NewWebSSH(sshCfg)

	staticDir := cfg.StaticDir()
	r.GET("/ws/:id", handle.ServeConn)
	r.Static("/static", staticDir)
	r.LoadHTMLFiles(path.Join(staticDir, "index.html"))
	r.GET(
		"/", func(c *gin.Context) {
			c.HTML(200, "index.html", nil)
		},
	)
	if err := r.Run(cfg.Addr()); err != nil {
		panic(err)
	}
}
