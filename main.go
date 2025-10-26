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
  "time"
  "bufio"
  "runtime/debug"
  _ "embed"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/gorilla/websocket"
)

// Embed VERSION file contents
//go:embed VERSION
var versionFile string

// version can be set at build time via -ldflags "-X main.version=X.Y.Z"
var version = ""

// Version returns the application version
var Version = getVersion()

func getVersion() string {
	// If version was explicitly set via ldflags (Docker/release builds), use it
	if version != "" {
		return version
	}

	// Use embedded VERSION file (for go install @latest or go build)
	if versionFile != "" {
		return strings.TrimSpace(versionFile)
	}

	// Try to get version from build info (for go install @vX.Y.Z)
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}

	return "dev"
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity
	},
}

type PodContainer struct {
	PodName       string `json:"podName"`
	ContainerName string `json:"containerName"`
	Namespace     string `json:"namespace"`
	ID            string `json:"id"`
}

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

  // k8s client setup - try in-cluster config first, fall back to kubeconfig
  var config *rest.Config
  var err error

  config, err = rest.InClusterConfig()
  if err != nil {
    // Not in cluster, try kubeconfig
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    configOverrides := &clientcmd.ConfigOverrides{}
    kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
    config, err = kubeConfig.ClientConfig()
    if err != nil {
      panic(fmt.Sprintf("Failed to load kubernetes config: %v", err))
    }
  }

  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
    panic(err.Error())
  }

  // Get the namespace - try in-cluster first, fall back to kubeconfig namespace or "default"
  var namespace string
  namespaceByte, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
  if err == nil {
    namespace = string(namespaceByte)
  } else {
    // Not in cluster, get namespace from kubeconfig or use default
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    configOverrides := &clientcmd.ConfigOverrides{}
    kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
    namespace, _, err = kubeConfig.Namespace()
    if err != nil || namespace == "" {
      namespace = "default"
    }
  }
  fmt.Println("Using namespace:", namespace)

  r := gin.New()
  r.Use(
        gin.LoggerWithWriter(gin.DefaultWriter, "/healthcheck"),
        gin.Recovery(),
  )

  // Authentication middleware
  authMiddleware := func(c *gin.Context) {
    if os.Getenv("LOGKEY") != "" {
      key := c.Query("key")
      if key == "" {
        key = c.GetHeader("X-API-Key")
      }
      if os.Getenv("LOGKEY") != key {
        c.JSON(http.StatusForbidden, gin.H{"error": "Invalid or missing API key"})
        c.Abort()
        return
      }
    }
    c.Next()
  }

  // Health check
  r.GET("/healthcheck", func(c *gin.Context) {
    c.String(http.StatusOK, "still alive")
  })

  // Version endpoint
  r.GET("/version", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "version": Version,
      "namespace": namespace,
    })
  })

  // Serve the UI
  r.GET("/", func(c *gin.Context) {
    c.Header("Content-Type", "text/html")
    c.String(http.StatusOK, getHTMLUI())
  })

  // API: List all pods and containers
  r.GET("/api/containers", authMiddleware, func(c *gin.Context) {
    pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }

    var containers []PodContainer
    for _, pod := range pods.Items {
      for _, container := range pod.Spec.Containers {
        id := fmt.Sprintf("%s/%s", pod.Name, container.Name)
        containers = append(containers, PodContainer{
          PodName:       pod.Name,
          ContainerName: container.Name,
          Namespace:     namespace,
          ID:            id,
        })
      }
    }

    c.JSON(http.StatusOK, gin.H{
      "namespace": namespace,
      "containers": containers,
    })
  })

  // API: Get logs for a specific container
  r.GET("/api/logs/:pod/:container", authMiddleware, func(c *gin.Context) {
    podName := c.Param("pod")
    containerName := c.Param("container")

    var loglines = int64(100)
    if linesParam := c.Query("lines"); linesParam != "" {
      if linesVal, err := strconv.Atoi(linesParam); err == nil {
        loglines = int64(linesVal)
      }
    }

    podLogOpts := corev1.PodLogOptions{
      Container: containerName,
      TailLines: &loglines,
    }

    req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
    logStream, err := req.Stream(context.TODO())
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    defer logStream.Close()

    buf := new(strings.Builder)
    io.Copy(buf, logStream)

    c.JSON(http.StatusOK, gin.H{
      "pod":       podName,
      "container": containerName,
      "logs":      buf.String(),
    })
  })

  // WebSocket: Stream logs in real-time
  r.GET("/ws/logs/:pod/:container", func(c *gin.Context) {
    // Check authentication for WebSocket
    if os.Getenv("LOGKEY") != "" {
      key := c.Query("key")
      if os.Getenv("LOGKEY") != key {
        c.JSON(http.StatusForbidden, gin.H{"error": "Invalid or missing API key"})
        return
      }
    }

    podName := c.Param("pod")
    containerName := c.Param("container")

    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
      fmt.Println("WebSocket upgrade failed:", err)
      return
    }
    defer conn.Close()

    // Stream logs with follow enabled
    podLogOpts := corev1.PodLogOptions{
      Container: containerName,
      Follow:    true,
      TailLines: func(i int64) *int64 { return &i }(100),
    }

    req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
    logStream, err := req.Stream(context.TODO())
    if err != nil {
      conn.WriteJSON(gin.H{"error": err.Error()})
      return
    }
    defer logStream.Close()

    // Read logs line by line and send over WebSocket
    scanner := bufio.NewScanner(logStream)
    for scanner.Scan() {
      line := scanner.Text()
      if err := conn.WriteJSON(gin.H{
        "timestamp": time.Now().Format(time.RFC3339),
        "log":       line,
      }); err != nil {
        break
      }
    }

    if err := scanner.Err(); err != nil {
      conn.WriteJSON(gin.H{"error": err.Error()})
    }
  })

  // Legacy endpoint - keep for backward compatibility
  r.GET("/logs", authMiddleware, func(c *gin.Context) {
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
  fmt.Printf("k8s-simple-logs version %s\n", Version)
  r := setupRouter()
  // Listen and Server in 0.0.0.0:8080
  r.Run(":8080")
}
