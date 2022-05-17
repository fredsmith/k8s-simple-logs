package main

import (
  "net/http"
  "os"
  "github.com/gin-gonic/gin"
  "context"
  "fmt"
  "strings"
  "io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func setupRouter() *gin.Engine {
  // Disable Console Color
  gin.DisableConsoleColor()
  if os.Getenv("DEBUG") != "" {
    gin.SetMode(gin.DebugMode)
  } else {
    gin.SetMode(gin.ReleaseMode)
  }
  logkey := os.Getenv("LOGKEY")
  fmt.Println("Logkey is: ", logkey)
  // k8s client setup
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
  clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
  // get the namespace we're in.
  namespaceByte, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
  namespace := fmt.Sprintf("%s", namespaceByte)
  // globally set pod logs options
  podLogOpts := corev1.PodLogOptions{}

  if err != nil {
    panic(err)
  }
  r := gin.New()
  r.Use(
        gin.LoggerWithWriter(gin.DefaultWriter, "/healthcheck"),
        gin.Recovery(),
  )
  // Ping
  r.GET("/healthcheck", func(c *gin.Context) {
    c.String(http.StatusOK, "still alive")
  })

  r.GET("/logs", func(c *gin.Context) {
    // check if there's a key in the environment, if so, make sure it's in the request
    if os.Getenv("LOGKEY") != "" {
      if os.Getenv("LOGKEY") != strings.Join(c.Request.URL.Query()["key"], " ") {
        c.String(http.StatusForbidden, "Key Required")
        c.Abort()
      }
    }

    var output = ""
    // get all pods in our namespace
    pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        panic(err.Error())
    }

    for i, pod := range pods.Items {
      buf := new(strings.Builder)
      // get the logs here 
      req := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &podLogOpts)
      // 
      output += "\n\n\n-----------------------------\n"
      output += fmt.Sprintf("ID: %d, \n Namespace: %s \n Pod: %s:\n", i, namespace, pod.Name)
      output += "-----------------------------\n"

      logoutput, err := req.Stream(context.TODO())
      if err != nil {
          panic(err.Error())
      }
      io.Copy(buf,logoutput)
      output += buf.String() 
      output += "-----------------------------\n\n\n\n\n"
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
