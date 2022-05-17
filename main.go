package main

import (
  "net/http"
  "os"
  "github.com/gin-gonic/gin"
  "crypto/tls"
  "crypto/x509"
  "context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func setupRouter() *gin.Engine {
// Disable Console Color
// gin.DisableConsoleColor()
  // k8s client setup
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
  clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

  r := gin.Default()

  // Ping
  r.GET("/healthcheck", func(c *gin.Context) {
    c.String(http.StatusOK, "still alive")
  })

  r.GET("/logs", func(c *gin.Context) {
   // get the logs here 
    if err != nil {
      c.String(http.StatusServiceUnavailable, err)
    } else {
      c.String(http.StatusOK, response)
    }
  })

  return r
}

func main() {
  r := setupRouter()
  // Listen and Server in 0.0.0.0:8080
  r.Run(":8080")
}
