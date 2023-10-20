package logging

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	kapkube "github.com/kapetacom/insight-api/kubernetes"
	"github.com/labstack/echo/v4"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

func LogByInstanceID(c echo.Context) error {
	return logBlockById(c)
}

func LogByInstanceName(c echo.Context) error {
	return logBlockByName(c)
}

func logBlockById(c echo.Context) error {
	tail := c.QueryParam("tail") != ""
	previous := c.QueryParam("previous") != ""
	podName := c.Param("instance")
	namespace := "services"
	if c.QueryParam("namespace") != "" {
		namespace = c.QueryParam("namespace")
	}
	ctx := c.Request().Context()

	clientset, err := kapkube.KubernetesClient()
	if err != nil {
		return fmt.Errorf("error getting kubernetes client: %v", err)
	}

	// find the pod(s) with the given block id
	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "kapeta.com/block-id=" + podName,
	})
	if err != nil {
		return fmt.Errorf("error getting pods: %v", err)
	}
	if len(podList.Items) == 0 {
		return fmt.Errorf("no pods found with label kapeta.com/block-id=%s", podName)

	}
	return writeLog(ctx, c, podList, namespace, clientset, tail, previous)
}

func logBlockByName(c echo.Context) error {
	tail := c.QueryParam("tail") != ""
	previous := c.QueryParam("previous") != ""
	podName := c.Param("name")
	namespace := "services"
	if c.QueryParam("namespace") != "" {
		namespace = c.QueryParam("namespace")
	}

	ctx := c.Request().Context()
	clientset, err := kapkube.KubernetesClient()
	if err != nil {
		return fmt.Errorf("error getting kubernetes client: %v", err)
	}

	// find the pod(s) with the given block id
	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "instance=" + podName,
	})
	if err != nil {
		return fmt.Errorf("error getting pods: %v", err)
	}
	if len(podList.Items) == 0 {
		return fmt.Errorf("no pods found with label instance=%s", podName)
	}
	return writeLog(ctx, c, podList, namespace, clientset, tail, previous)
}

func writeLog(ctx context.Context, c echo.Context, podList *corev1.PodList, namespace string, clientset *kubernetes.Clientset, tail bool, previous bool) error {
	for _, pod := range podList.Items {
		podName := pod.Name
		req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
			Follow:     tail,
			Previous:   previous,
			Timestamps: true,
		})
		readCloser, err := req.Stream(ctx)
		if err != nil {
			return fmt.Errorf("error opening stream to pod logs: %v", err)

		}

		defer func(readCloser io.ReadCloser) {
			err := readCloser.Close()
			if err != nil {
				fmt.Printf("error closing stream to pod logs: %v", err)
			}
		}(readCloser)

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		enc := json.NewEncoder(c.Response())

		lineReader := bufio.NewScanner(readCloser)

		entries := make([]*LogEntry, 0)

		for lineReader.Scan() {
			line := lineReader.Text()
			// Split the line into timestamp and message
			timestamp := line[0:30]
			message := line[31:]

			miliseconds := int64(0)
			parsedTime, err := time.Parse(time.RFC3339, timestamp)
			if err == nil {
				miliseconds = parsedTime.UnixMilli()
			}

			logEntry := LogEntry{
				Entity:    podName,
				Severity:  "INFO",
				Timestamp: miliseconds,
				Message:   message,
			}
			entries = append(entries, &logEntry)
		}

		err = enc.Encode(entries)
		if err != nil {
			return fmt.Errorf("error writing to response: %v", err)
		}
	}
	return nil
}
