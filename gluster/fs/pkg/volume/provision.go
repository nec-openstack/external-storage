/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package volume

import (
	"bytes"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	apiRemotecommand "k8s.io/apimachinery/pkg/util/remotecommand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	// are we allowed to set this? else make up our own
	annCreatedBy = "kubernetes.io/createdby"
	createdBy    = "glusterfs-simple-provisioner"

	// A PV annotation for the identity of the s3fsProvisioner that provisioned it
	annProvisionerID = "Provisioner_Id"
)

// NewGlusterfsProvisioner creates a new glusterfs simple provisioner
func NewGlusterfsProvisioner(config *rest.Config, client kubernetes.Interface) controller.Provisioner {
	glog.Infof("Creating NewGlusterfsProvisioner.")
	return newGlusterfsProvisionerInternal(config, client)
}

func newGlusterfsProvisionerInternal(config *rest.Config, client kubernetes.Interface) *glusterfsProvisioner {
	var identity types.UID

	restClient := client.Core().RESTClient()
	provisioner := &glusterfsProvisioner{
		config:     config,
		client:     client,
		restClient: restClient,
		identity:   identity,
	}

	return provisioner
}

type glusterfsProvisioner struct {
	client     kubernetes.Interface
	restClient rest.Interface
	config     *rest.Config
	identity   types.UID
}

var _ controller.Provisioner = &glusterfsProvisioner{}

func (p *glusterfsProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	glog.Infoln("Start Provisioning volume...")
	pod, err := p.client.CoreV1().Pods("default").Get("busybox", meta_v1.GetOptions{})
	if err != nil {
		if err != nil {
			glog.Fatalf("Failed to retrive Pod: %v", err)
		}
		return nil, err
	}
	containerName := pod.Spec.Containers[0].Name
	req := p.restClient.Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		Param("container", containerName).
		Param("stdout", "true").
		Param("stderr", "true")

	commands := []string{"nslookup", "kubernetes"}
	for _, command := range commands {
		req.Param("command", command)
	}

	exec, err := remotecommand.NewExecutor(p.config, "POST", req.URL())
	if err != nil {
		glog.Fatalf("Failed to create NewExecutor: %v", err)
		return nil, err
	}

	var b bytes.Buffer
	var berr bytes.Buffer

	err = exec.Stream(remotecommand.StreamOptions{
		SupportedProtocols: apiRemotecommand.SupportedStreamingProtocols,
		Stdout:             &b,
		Stderr:             &berr,
		Tty:                false,
	})
	if err != nil {
		glog.Fatalf("Failed to create Stream: %v", err)
		return nil, err
	}

	return nil, nil
}
