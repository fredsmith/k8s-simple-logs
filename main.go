package main

import (
  "net/http"
  "os"
  "github.com/gin-gonic/gin"
  "context"
  "fmt"
  "strings"
  "strconv"
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
        return
      }
    }

    var output = ""
    var loglines = int64(20)
    loglinesval, err := strconv.Atoi(strings.Join(c.Request.URL.Query()["lines"], " "))
    if err == nil {
      loglines = int64(loglinesval)
    }

    // get all pods in our namespace
    pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        panic(err.Error())
    }

    for i, pod := range pods.Items {
      poddetails, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
      if err != nil {
          panic(err.Error())
      }
      for j, container := range poddetails.Spec.Containers {
        podLogOpts := corev1.PodLogOptions{
        Container: container.Name,
        TailLines: &loglines,
        }

        buf := new(strings.Builder)
        // get the logs here 
        req := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &podLogOpts)
        // 
        output += "\n\n\n-----------------------------\n"
        output += fmt.Sprintf("ID: %d %d, \n Namespace: %s \n Pod: %s:\n Container: %s\n", i, j, namespace, pod.Name, container.Name)
        output += "-----------------------------\n"

        logoutput, err := req.Stream(context.TODO())
        if err != nil {
            panic(err.Error())
        }
        io.Copy(buf,logoutput)
        output += buf.String() 
        output += "-----------------------------\n\n\n\n\n"
      }
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
