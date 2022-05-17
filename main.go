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

  namespace, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
  if err != nil {
    panic(err)
  }
  r := gin.Default()

  // Ping
  r.GET("/healthcheck", func(c *gin.Context) {
    c.String(http.StatusOK, "still alive")
  })

  r.GET("/logs", func(c *gin.Context) {
    var output :=
  // get all pods in our namespace
  pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
  if err != nil {
			panic(err.Error())
	}

  for i, podID := range pods {
     // get the logs here 
     req := clientset.RESTClient.Get().
          Namespace(namespace).
          Name(podID).
          Resource("pods").
          SubResource("log").
          
     req.param(tailLines, "40")

     output += req.Stream()
   }

   c.String(http.StatusOK, output)
  })

  return r
}

func main() {
  r := setupRouter()
  // Listen and Server in 0.0.0.0:8080
  r.Run(":8080")
}
