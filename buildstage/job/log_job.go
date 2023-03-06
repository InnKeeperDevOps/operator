package job

import (
	"bytes"
	"context"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func (j *JobPod) Log(ctx context.Context) (error, string) {

	podLogOpts := corev1.PodLogOptions{}
	config, err := config.GetConfigWithContext("")
	if err != nil {
		return err, "error in getting config"
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err, "error in getting access to K8S"
	}
	req := clientset.CoreV1().Pods(j.pod.Namespace).GetLogs(j.pod.Name, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return err, "error in opening stream"
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return err, "error in copy information from podLogs to buf"
	}
	str := buf.String()

	return nil, str
}
