package python

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GKERunner executes Python code in a GKE Sandbox (gVisor).
type GKERunner struct {
	clientset    *kubernetes.Clientset
	namespace    string
	runtimeClass string
	image        string
}

// NewGKERunner creates a new GKERunner.
func NewGKERunner(ctx context.Context, namespace, image string) (*GKERunner, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig for local development
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		config, err = kubeConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &GKERunner{
		clientset:    clientset,
		namespace:    namespace,
		runtimeClass: "gvisor", // Pre-configured in GKE
		image:        image,
	}, nil
}

// ExecuteScript runs a Python script in a GKE Sandbox by creating an ephemeral Job.
func (r *GKERunner) ExecuteScript(ctx context.Context, code string, contextFileName string, handler IPCHandler) (string, error) {
	id := time.Now().UnixNano()
	jobName := fmt.Sprintf("python-sandbox-%d", id)
	configMapName := fmt.Sprintf("python-code-%d", id)
	slog.Info("Creating GKE Sandbox Job and ConfigMap", "name", jobName)

	fullCode := HelperCode + "\n\n" + code
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: r.namespace,
		},
		Data: map[string]string{
			"script.py": fullCode,
		},
	}

	_, err := r.clientset.CoreV1().ConfigMaps(r.namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create configmap: %w", err)
	}
	defer r.clientset.CoreV1().ConfigMaps(r.namespace).Delete(ctx, configMapName, metav1.DeleteOptions{})

	// In a real implementation, we would mount the context file via GCS Fuse.
	// For this prototype, we assume the file is available or small enough to inject.
	
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: r.namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            new(int32),
			TTLSecondsAfterFinished: int32Ptr(60),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RuntimeClassName: &r.runtimeClass,
					RestartPolicy:    corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "worker",
							Image: r.image,
							Command: []string{"python", "/mnt/code/script.py"},
							Env: []corev1.EnvVar{
								{Name: "CONTEXT_FILE", Value: contextFileName},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "code-volume",
									MountPath: "/mnt/code",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "code-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMapName,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = r.clientset.BatchV1().Jobs(r.namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create job: %w", err)
	}
	defer r.cleanup(ctx, jobName)

	return r.waitForJobAndGetLogs(ctx, jobName)
}

func (r *GKERunner) waitForJobAndGetLogs(ctx context.Context, jobName string) (string, error) {
	// Polling for job completion (simplified for prototype)
	for {
		job, err := r.clientset.BatchV1().Jobs(r.namespace).Get(ctx, jobName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		if job.Status.Succeeded > 0 {
			break
		}
		if job.Status.Failed > 0 {
			return "", fmt.Errorf("job failed")
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(2 * time.Second):
			continue
		}
	}

	// Get Pod logs
	pods, err := r.clientset.CoreV1().Pods(r.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil || len(pods.Items) == 0 {
		return "", fmt.Errorf("failed to find pods for job: %w", err)
	}

	podName := pods.Items[0].Name
	logReq := r.clientset.CoreV1().Pods(r.namespace).GetLogs(podName, &corev1.PodLogOptions{})
	logStream, err := logReq.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to stream logs: %w", err)
	}
	defer logStream.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, logStream)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}

	return buf.String(), nil
}

func (r *GKERunner) cleanup(ctx context.Context, jobName string) {
	propagationPolicy := metav1.DeletePropagationBackground
	r.clientset.BatchV1().Jobs(r.namespace).Delete(ctx, jobName, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
}

func int32Ptr(i int32) *int32 { return &i }
