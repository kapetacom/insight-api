package logging

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"

	kapkube "github.com/kapetacom/insight-api/kubernetes"
	"github.com/labstack/echo/v4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
			Follow:   tail,
			Previous: previous,
		})
		readCloser, err := req.Stream(ctx)
		if err != nil {
			return fmt.Errorf("error opening stream to pod logs: %v", err)

		}
		defer readCloser.Close()

		buf := make([]byte, 4096)
		for {
			n, err := readCloser.Read(buf)
			if err != nil {
				break
			}
			_, err = c.Response().Writer.Write(prefixLine(buf[:n], podName+": "))
			if err != nil {
				return fmt.Errorf("error writing to response: %v", err)
			}
		}
	}
	return nil
}

func prefixLine(buffer []byte, prefix string) []byte {

	// Convert the byte buffer to a string
	inputString := string(buffer)

	// Split the input string into lines
	scanner := bufio.NewScanner(strings.NewReader(inputString))

	// Create a buffer to store the modified multiline string
	var outputBuffer bytes.Buffer

	// Iterate through each line and add the "test:" prefix
	for scanner.Scan() {
		line := scanner.Text()
		modifiedLine := prefix + line + "\n"
		outputBuffer.WriteString(modifiedLine)
	}

	if scanner.Err() != nil {
		fmt.Println("Error:", scanner.Err())
		return nil
	}

	return outputBuffer.Bytes()
}
