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

package main

import (
	"flag"

	vol "github.com/kubernetes-incubator/external-storage/gluster/fs/pkg/volume"
)

var (
	provisioner = flag.String("provisioner", "gluster.org/gluster-simple", "Name of the provisioner. The provisioner will only provision volumes for claims that request a StorageClass with a provisioner field set equal to this name.")
	master      = flag.String("master", "", "Master URL to build a client config from. Either this or kubeconfig needs to be set if the provisioner is being run out of cluster.")
	kubeconfig  = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file. Either this or master needs to be set if the provisioner is being run out of cluster.")
)

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	glusterfsProvisioner := vol.NewGlusterfsProvisioner(nil)
	print(glusterfsProvisioner)
}
