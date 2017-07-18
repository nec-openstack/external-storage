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
	"fmt"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
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
	if options.PVC.Spec.Selector != nil {
		return nil, fmt.Errorf("claim Selector is not supported")
	}
	glog.V(4).Infof("Start Provisioning volume: VolumeOptions %v", options)

	pvcNamespace := options.PVC.Namespace
	pvcName := options.PVC.Name
	cfg, err := NewProvisionerConfig(options.Parameters)
	if err != nil {
		return nil, fmt.Errorf("Parameter is invalid: %s", err)
	}

	vol, err := p.createVolume(pvcNamespace, pvcName, cfg)
	if err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("Test provision is faild %v", vol)
}

func (p *glusterfsProvisioner) createVolume(
	namespace string, name string, cfg *ProvisionerConfig) (*glusterVolume, error) {
	var err error

	err = p.createBricks(namespace, name, cfg)
	if err != nil {
		glog.Errorf("Creating bricks is failed: %s,%s", namespace, name)
		return nil, err
	}

	return nil, nil
}

func (p *glusterfsProvisioner) createBricks(
	namespace string, name string, cfg *ProvisionerConfig) error {
	var commands []string

	for _, brick := range cfg.BrickRootPaths {
		host := brick.Host
		parentDir := filepath.Join(brick.Path, namespace)
		path := filepath.Join(parentDir, name)

		glog.Infof("mkdir -p %s:%s", host, path)
		commands = []string{
			// Create parent directory
			fmt.Sprintf("mkdir -p %s", parentDir),
			// Run mkdir (if path is already existed then this command will fail)
			fmt.Sprintf("mkdir %s", path),
			// TODO: Assign GID
		}
		err := p.ExecuteCommands(host, commands, cfg)
		if err != nil {
			// TODO: Cleanup created directories if commands failed
			return err
		}
	}

	return nil
}
