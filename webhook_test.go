package webhook

import (
	"context"
	"errors"
	"testing"

	kwhlog "github.com/slok/kubewebhook/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/secrethub/secrethub-go/internals/assert"
)

func TestMutate(t *testing.T) {
	cases := map[string]struct {
		input    corev1.Pod
		expected corev1.Pod
		err      error
	}{
		"one annotated container and version set": {
			input: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub.io/mutate":  "app",
						"secrethub.io/version": "0.38.0",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app",
							Command: []string{"foo"},
							Env: []corev1.EnvVar{
								{
									Name:  "API_KEY",
									Value: "secrethub://path/to/api/key",
								},
							},
						},
					},
				},
			},
			expected: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub.io/mutate":  "app",
						"secrethub.io/version": "0.38.0",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app",
							Command: []string{"/secrethub/bin/secrethub", "run", "--", "foo"},
							Env: []corev1.EnvVar{
								{
									Name:  "API_KEY",
									Value: "secrethub://path/to/api/key",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrethub-bin",
									MountPath: "/secrethub/bin/",
									ReadOnly:  true,
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:            "copy-secrethub-bin",
							Image:           "secrethub/cli:0.38.0",
							Command:         []string{"sh", "-c", "cp /usr/bin/secrethub /secrethub/bin/"},
							ImagePullPolicy: corev1.PullIfNotPresent,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrethub-bin",
									MountPath: "/secrethub/bin/",
									ReadOnly:  false,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "secrethub-bin",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: corev1.StorageMediumMemory,
								},
							},
						},
					},
				},
			},
		},
		"mutate one of two containers": {
			input: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub.io/mutate":  "app",
						"secrethub.io/version": "0.38.0",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app",
							Command: []string{"foo"},
							Env: []corev1.EnvVar{
								{
									Name:  "API_KEY",
									Value: "secrethub://path/to/api/key",
								},
							},
						},
						{
							Name:    "app2",
							Command: []string{"foo"},
							Env: []corev1.EnvVar{
								{
									Name:  "API_KEY",
									Value: "secrethub://path/to/api/key",
								},
							},
						},
					},
				},
			},
			expected: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub.io/mutate":  "app",
						"secrethub.io/version": "0.38.0",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app",
							Command: []string{"/secrethub/bin/secrethub", "run", "--", "foo"},
							Env: []corev1.EnvVar{
								{
									Name:  "API_KEY",
									Value: "secrethub://path/to/api/key",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrethub-bin",
									MountPath: "/secrethub/bin/",
									ReadOnly:  true,
								},
							},
						},
						{
							Name:    "app2",
							Command: []string{"foo"},
							Env: []corev1.EnvVar{
								{
									Name:  "API_KEY",
									Value: "secrethub://path/to/api/key",
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:            "copy-secrethub-bin",
							Image:           "secrethub/cli:0.38.0",
							Command:         []string{"sh", "-c", "cp /usr/bin/secrethub /secrethub/bin/"},
							ImagePullPolicy: corev1.PullIfNotPresent,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrethub-bin",
									MountPath: "/secrethub/bin/",
									ReadOnly:  false,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "secrethub-bin",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: corev1.StorageMediumMemory,
								},
							},
						},
					},
				},
			},
		},
		"mutate two containers": {
			input: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub.io/mutate":  "app,app2",
						"secrethub.io/version": "0.38.0",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app",
							Command: []string{"foo"},
							Env: []corev1.EnvVar{
								{
									Name:  "API_KEY",
									Value: "secrethub://path/to/api/key",
								},
							},
						},
						{
							Name:    "app2",
							Command: []string{"foo"},
							Env: []corev1.EnvVar{
								{
									Name:  "API_KEY",
									Value: "secrethub://path/to/api/key",
								},
							},
						},
					},
				},
			},
			expected: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub.io/mutate":  "app,app2",
						"secrethub.io/version": "0.38.0",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app",
							Command: []string{"/secrethub/bin/secrethub", "run", "--", "foo"},
							Env: []corev1.EnvVar{
								{
									Name:  "API_KEY",
									Value: "secrethub://path/to/api/key",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrethub-bin",
									MountPath: "/secrethub/bin/",
									ReadOnly:  true,
								},
							},
						},
						{
							Name:    "app2",
							Command: []string{"/secrethub/bin/secrethub", "run", "--", "foo"},
							Env: []corev1.EnvVar{
								{
									Name:  "API_KEY",
									Value: "secrethub://path/to/api/key",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrethub-bin",
									MountPath: "/secrethub/bin/",
									ReadOnly:  true,
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:            "copy-secrethub-bin",
							Image:           "secrethub/cli:0.38.0",
							Command:         []string{"sh", "-c", "cp /usr/bin/secrethub /secrethub/bin/"},
							ImagePullPolicy: corev1.PullIfNotPresent,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrethub-bin",
									MountPath: "/secrethub/bin/",
									ReadOnly:  false,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "secrethub-bin",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: corev1.StorageMediumMemory,
								},
							},
						},
					},
				},
			},
		},
		"default to latest version": {
			input: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub.io/mutate": "app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app",
							Command: []string{"foo"},
							Env: []corev1.EnvVar{
								{
									Name:  "API_KEY",
									Value: "secrethub://path/to/api/key",
								},
							},
						},
					},
				},
			},
			expected: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub.io/mutate": "app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app",
							Command: []string{"/secrethub/bin/secrethub", "run", "--", "foo"},
							Env: []corev1.EnvVar{
								{
									Name:  "API_KEY",
									Value: "secrethub://path/to/api/key",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrethub-bin",
									MountPath: "/secrethub/bin/",
									ReadOnly:  true,
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:            "copy-secrethub-bin",
							Image:           "secrethub/cli:latest",
							Command:         []string{"sh", "-c", "cp /usr/bin/secrethub /secrethub/bin/"},
							ImagePullPolicy: corev1.PullIfNotPresent,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrethub-bin",
									MountPath: "/secrethub/bin/",
									ReadOnly:  false,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "secrethub-bin",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: corev1.StorageMediumMemory,
								},
							},
						},
					},
				},
			},
		},
		"support image override and ignore version": {
			input: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub.io/mutate":        "app",
						"secrethub.io/version":       "123",
						"secrethub.io/imageOverride": "123456789.dkr.ecr.us-west-2.amazonaws.com/secrethub/cli:latest",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app",
							Command: []string{"foo"},
						},
					},
				},
			},
			expected: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub.io/mutate":        "app",
						"secrethub.io/version":       "123",
						"secrethub.io/imageOverride": "123456789.dkr.ecr.us-west-2.amazonaws.com/secrethub/cli:latest",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app",
							Command: []string{"/secrethub/bin/secrethub", "run", "--", "foo"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrethub-bin",
									MountPath: "/secrethub/bin/",
									ReadOnly:  true,
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:            "copy-secrethub-bin",
							Image:           "123456789.dkr.ecr.us-west-2.amazonaws.com/secrethub/cli:latest",
							Command:         []string{"sh", "-c", "cp /usr/bin/secrethub /secrethub/bin/"},
							ImagePullPolicy: corev1.PullIfNotPresent,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrethub-bin",
									MountPath: "/secrethub/bin/",
									ReadOnly:  false,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "secrethub-bin",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: corev1.StorageMediumMemory,
								},
							},
						},
					},
				},
			},
		},
		"mutate container without secret references": {
			input: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub.io/mutate": "app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app",
							Command: []string{"foo"},
						},
					},
				},
			},
			expected: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub.io/mutate": "app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app",
							Command: []string{"/secrethub/bin/secrethub", "run", "--", "foo"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrethub-bin",
									MountPath: "/secrethub/bin/",
									ReadOnly:  true,
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:            "copy-secrethub-bin",
							Image:           "secrethub/cli:latest",
							Command:         []string{"sh", "-c", "cp /usr/bin/secrethub /secrethub/bin/"},
							ImagePullPolicy: corev1.PullIfNotPresent,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrethub-bin",
									MountPath: "/secrethub/bin/",
									ReadOnly:  false,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "secrethub-bin",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: corev1.StorageMediumMemory,
								},
							},
						},
					},
				},
			},
		},
		"ignoring pod without annotation": {
			input:    corev1.Pod{},
			expected: corev1.Pod{},
		},
		"failing pod without command": {
			input: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub.io/mutate": "foo",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "foo",
							Env: []corev1.EnvVar{
								{
									Name:  "API_KEY",
									Value: "secrethub://path/to/api/key",
								},
							},
						},
					},
				},
			},
			err: errors.New("not attaching SecretHub to the container foo: the podspec does not define a command"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			logger := &kwhlog.Std{Debug: true}
			mutator := &SecretHubMutator{logger: logger}

			actual := &tc.input

			_, err := mutator.Mutate(context.Background(), actual)

			if tc.err == nil {
				assert.Equal(t, actual, tc.expected)
			}
			assert.Equal(t, err, tc.err)
		})
	}
}
