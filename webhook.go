package webhook

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	kwhhttp "github.com/slok/kubewebhook/pkg/http"
	kwhlog "github.com/slok/kubewebhook/pkg/log"
	kwhmutating "github.com/slok/kubewebhook/pkg/webhook/mutating"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

	containersStr, enabled := pod.Annotations["secrethub.io/mutate"]
	if !enabled {
		m.logger.Debugf("Skipping pod %s because it is not annotated with secrethub", pod.Name)
		return false, nil
	}

	containers := map[string]struct{}{}

	for _, container := range strings.Split(containersStr, ",") {
		containers[container] = struct{}{}
	}

	version, ok := pod.Annotations["secrethub.io/version"]
	if !ok {
		version = "latest"
	}

	mutated := false

	for i, c := range pod.Spec.InitContainers {
		_, mutate := containers[c.Name]
		if !mutate {
			continue
		}

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
		_, mutate := containers[c.Name]
		if !mutate {
			continue
		}

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

	// Set app info
	c.Env = append(c.Env, appInfo...)

	return c, true, nil
}

// Handler is the http.Handler that responds to webhooks
func Handler() http.Handler {
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

// F is the exported webhook for the function to bind.
var F = Handler().ServeHTTP

