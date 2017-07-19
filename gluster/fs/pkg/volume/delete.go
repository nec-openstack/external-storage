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

	"github.com/golang/glog"
	"k8s.io/client-go/pkg/api/v1"
)

func (p *glusterfsProvisioner) Delete(volume *v1.PersistentVolume) error {
	var err error
	class, err := GetClassForVolume(p.client, volume)
	if err != nil {
		glog.Errorf("Fail to get class for volume: %v", volume)
		return err
	}
	cfg, err := NewProvisionerConfig(volume.Name, class.Parameters)
	if err != nil {
		return fmt.Errorf("Parameter is invalid: %s", err)
	}

	pvc := volume.Spec.ClaimRef
	if pvc == nil {
		glog.Errorf("glusterfs: ClaimRef is nil")
		return fmt.Errorf("glusterfs: ClaimRef is nil")
	}
	if pvc.Namespace == "" {
		glog.Errorf("glusterfs: namespace is nil")
		return fmt.Errorf("glusterfs: namespace is nil")
	}
	err = p.deleteVolume(pvc.Namespace, pvc.Name, cfg)
	if err != nil {
		return err
	}

	return nil
}

func (p *glusterfsProvisioner) deleteVolume(
	namespace string, name string, cfg *ProvisionerConfig) error {
	var cmds []string
	var err error
	host := cfg.BrickRootPaths[0].Host

	cmds = []string{
		fmt.Sprintf("gluster --mode=script volume stop %s force", cfg.VolumeName),
	}
	err = p.ExecuteCommands(host, cmds, cfg)
	if err != nil {
		glog.Errorf("glusterfs: failed to stop volume: %s", cfg.VolumeName)
		return fmt.Errorf("Failed to stop volume: %v, error: %s", cfg.VolumeName, err.Error())
	}

	cmds = []string{fmt.Sprintf(
		"gluster --mode=script volume delete %s", cfg.VolumeName,
	)}
	err = p.ExecuteCommands(host, cmds, cfg)
	if err != nil {
		glog.Errorf("glusterfs: failed to delete volume: %s", cfg.VolumeName)
		return fmt.Errorf("Failed to stop volume: %v, error: %s", cfg.VolumeName, err.Error())
	}

	epServiceName := dynamicEpSvcPrefix + name
	err = p.deleteEndpointService(namespace, epServiceName)
	if err != nil {
		return err
	}

	return nil
}

func (p *glusterfsProvisioner) deleteEndpointService(namespace string, epServiceName string) (err error) {
	kubeClient := p.client
	if kubeClient == nil {
		return fmt.Errorf("glusterfs: failed to get kube client when deleting endpoint service")
	}
	err = kubeClient.Core().Services(namespace).Delete(epServiceName, nil)
	if err != nil {
		glog.Errorf("glusterfs: error deleting service %s/%s: %v", namespace, epServiceName, err)
		return fmt.Errorf("error deleting service %s/%s: %v", namespace, epServiceName, err)
	}
	glog.V(1).Infof("glusterfs: service/endpoint %s/%s deleted successfully", namespace, epServiceName)
	return nil
}
