// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kube

import (
	"bytes"
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type CLIClient interface {
	// RESTConfig returns the Kubernetes rest.Config used to configure the clients.
	RESTConfig() *rest.Config

	// PodExec takes a command and the pod data to run the command in the specified pod.
	PodExec(podName, podNamespace, container string, command string) (stdout string, stderr string, err error)
}

type client struct {
	config     *rest.Config
	restClient *rest.RESTClient
	kube       kubernetes.Interface
}

func NewCLIClient(clientConfig clientcmd.ClientConfig) (CLIClient, error) {
	return newClientInternal(clientConfig)
}

func newClientInternal(clientConfig clientcmd.ClientConfig) (*client, error) {
	var (
		c   client
		err error
	)

	c.config, err = clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	c.restClient, err = rest.RESTClientFor(c.config)
	if err != nil {
		return nil, err
	}

	c.kube, err = kubernetes.NewForConfig(c.config)
	if err != nil {
		return nil, err
	}

	return &c, err
}

func (c *client) RESTConfig() *rest.Config {
	if c.config == nil {
		return nil
	}
	cpy := *c.config
	return &cpy
}

func (c *client) PodExec(podName, podNamespace, container string, command string) (stdout string, stderr string, err error) {
	defer func() {
		if err != nil {
			if len(stderr) > 0 {
				err = fmt.Errorf("error exec'ing into %s/%s %s container: %v\n%s",
					podNamespace, podName, container, err, stderr)
			} else {
				err = fmt.Errorf("error exec'ing into %s/%s %s container: %v",
					podNamespace, podName, container, err)
			}
		}
	}()

	req := c.restClient.Post().
		Resource("pods").
		Name(podName).
		Namespace(podNamespace).
		SubResource("exec").
		Param("container", container).
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   []string{command},
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, kubescheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return "", "", err
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
		Tty:    false,
	})

	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	return
}
