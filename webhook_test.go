package main

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
		"successfuly changing annotated pod": {
			input: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub": "0.38.0",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
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
						"secrethub": "0.38.0",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
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
		"ignoring pod without annotation": {
			input:    corev1.Pod{},
			expected: corev1.Pod{},
		},
		"failing pod without command": {
			input: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"secrethub": "0.38.0",
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
