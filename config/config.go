package config

import "github.com/gin-gonic/gin"

type Config struct {

}

func NewConfig(r *gin.Engine) *Config {
	config := Config{}
	config.initWebHandler()
	config.initGrpcHandler()
	return &config
}


func (c *Config) initWebHandler() {

}

func (c *Config) initGrpcHandler() {

}
