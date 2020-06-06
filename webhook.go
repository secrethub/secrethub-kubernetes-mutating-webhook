package main

import (
	"context"
	"fmt"
  "log"
	"net/http"
	"os"
	"strings"

	kwhhttp "github.com/slok/kubewebhook/pkg/http"
	kwhlog "github.com/slok/kubewebhook/pkg/log"
	kwhmutating "github.com/slok/kubewebhook/pkg/webhook/mutating"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
  http.HandleFunc("/", webhookHandler().ServeHTTP)

  log.Fatal(http.ListenAndServeTLS(":443", "/etc/webhook/certs/cert.pem", "/etc/webhook/certs/key.pem", nil))
}

const (
	// binVolumeName is the name of the volume where the SecretHub CLI binary is stored.
	binVolumeName = "secrethub-bin"

	// binVolumeMountPath is the mount path where the SecretHub CLI binary can be found.
	binVolumeMountPath = "/secrethub/bin/"
)

// binVolume is the shared, in-memory volume where the SecretHub CLI binary lives.
var binVolume = corev1.Volume{
	Name: binVolumeName,
	VolumeSource: corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{
			Medium: corev1.StorageMediumMemory,
		},
	},
}

// binVolumeMount is the shared volume mount where the SecretHub CLI binary lives.
var binVolumeMount = corev1.VolumeMount{
	Name:      binVolumeName,
	MountPath: binVolumeMountPath,
	ReadOnly:  true,
}

// SecretHubMutator is a mutator.
type SecretHubMutator struct {
	logger kwhlog.Logger
}

// Mutate implements MutateFunc and provides the top-level entrypoint for object
// mutation.
func (m *SecretHubMutator) Mutate(ctx context.Context, obj metav1.Object) (bool, error) {
	m.logger.Infof("calling mutate")

	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return false, nil
	}

	version, enabled := pod.Annotations["secrethub"]
	if !enabled {
		m.logger.Debugf("Skipping pod %s because it is not annotated with secrethub", pod.Name)
		return false, nil
	}

	mutated := false

	for i, c := range pod.Spec.InitContainers {
		c, didMutate, err := m.mutateContainer(ctx, &c)
		if err != nil {
			return false, err
		}
		if didMutate {
			mutated = true
			pod.Spec.InitContainers[i] = *c
		}
	}

	for i, c := range pod.Spec.Containers {
		c, didMutate, err := m.mutateContainer(ctx, &c)
		if err != nil {
			return false, err
		}
		if didMutate {
			mutated = true
			pod.Spec.Containers[i] = *c
		}
	}

	// binInitContainer is the container that pulls the SecretHub CLI
	// into a shared volume mount.
	var binInitContainer = corev1.Container{
		Name:            "copy-secrethub-bin",
		Image:           "secrethub/cli" + ":" + version,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command: []string{"sh", "-c",
			fmt.Sprintf("cp /usr/bin/secrethub %s", binVolumeMountPath)},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      binVolumeName,
				MountPath: binVolumeMountPath,
			},
		},
	}

	// If any of the containers requested SecretHub secrets, mount the shared volume
	// and ensure the SecretHub CLI is available via an init container.
	if mutated {
		pod.Spec.Volumes = append(pod.Spec.Volumes, binVolume)
		pod.Spec.InitContainers = append([]corev1.Container{binInitContainer},
			pod.Spec.InitContainers...)
	}

	return false, nil
}

// mutateContainer mutates the given container, updating the volume mounts and
// command if it contains SecretHub references.
func (m *SecretHubMutator) mutateContainer(_ context.Context, c *corev1.Container) (*corev1.Container, bool, error) {
	// Ignore if there are no SecretHub references in the container.
	if !m.hasSecretHubReferences(c.Env) {
		return c, false, nil
	}

	// This webhook only attaches SecretHub when a command is specified in the podspec.
	//
	// Note that the command should be defined in the podspec. The ENTRYPOINT or
	// CMD in the Dockerfile does not suffice as this is not visible to the webhook.
	if len(c.Command) == 0 {
		return c, false, fmt.Errorf("not attaching SecretHub to the container %s: the podspec does not define a command", c.Name)
	}

	// Prepend the command with secrethub run --
	c.Command = append([]string{binVolumeMountPath + "secrethub", "run", "--"}, c.Command...)

	// Add the shared volume mount
	c.VolumeMounts = append(c.VolumeMounts, binVolumeMount)

	return c, true, nil
}

// hasSecretHubReferences parses the environment and returns true if any of the
// environment variables includes a SecretHub reference.
func (m *SecretHubMutator) hasSecretHubReferences(env []corev1.EnvVar) bool {
	for _, e := range env {
		if strings.HasPrefix(e.Value, "secrethub://") {
			return true
		}
	}
	return false
}

// webhookHandler is the http.Handler that responds to webhooks
func webhookHandler() http.Handler {
	logger := &kwhlog.Std{Debug: true}

	mutator := &SecretHubMutator{logger: logger}

	mcfg := kwhmutating.WebhookConfig{
		Name: "SecretHubMutator",
		Obj:  &corev1.Pod{},
	}

	// Create the wrapping webhook
	wh, err := kwhmutating.NewWebhook(mcfg, mutator, nil, nil, logger)
	if err != nil {
		logger.Errorf("error creating webhook: %s", err)
		os.Exit(1)
	}

	// Get the handler for our webhook.
	whhandler, err := kwhhttp.HandlerFor(wh)
	if err != nil {
		logger.Errorf("error creating webhook handler: %s", err)
		os.Exit(1)
	}
	return whhandler
}