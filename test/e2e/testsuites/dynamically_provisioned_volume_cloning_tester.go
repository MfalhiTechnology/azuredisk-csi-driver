/*
Copyright 2019 The Kubernetes Authors.

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

package testsuites

import (
	"sigs.k8s.io/azuredisk-csi-driver/test/e2e/driver"

	"github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// DynamicallyProvisionedVolumeCloningTest will provision required StorageClass(es), PVC(s) and Pod(s)
type DynamicallyProvisionedVolumeCloningTest struct {
	CSIDriver           driver.DynamicPVTestDriver
	Pod                 PodDetails
	PodWithClonedVolume PodDetails
}

func (t *DynamicallyProvisionedVolumeCloningTest) Run(client clientset.Interface, namespace *v1.Namespace) {
	// create the storageClass
	tsc, tscCleanup := t.Pod.Volumes[0].CreateStorageClass(client, namespace, t.CSIDriver)
	defer tscCleanup()

	// create the pod
	t.Pod.Volumes[0].StorageClass = tsc.storageClass
	tpod, cleanups := t.Pod.SetupWithDynamicVolumes(client, namespace, t.CSIDriver)
	for i := range cleanups {
		defer cleanups[i]()
	}

	ginkgo.By("deploying the pod")
	tpod.Create()
	defer tpod.Cleanup()
	ginkgo.By("checking that the pod's command exits with no error")
	tpod.WaitForSuccess()

	ginkgo.By("cloning existing volume")
	clonedVolume := t.Pod.Volumes[0]
	clonedVolume.DataSource = &DataSource{
		Name: tpod.pod.Spec.Volumes[0].VolumeSource.PersistentVolumeClaim.ClaimName,
		Kind: VolumePVCKind,
	}
	clonedVolume.StorageClass = tsc.storageClass
	t.PodWithClonedVolume.Volumes = []VolumeDetails{clonedVolume}
	tpod, cleanups = t.PodWithClonedVolume.SetupWithDynamicVolumes(client, namespace, t.CSIDriver)
	for i := range cleanups {
		defer cleanups[i]()
	}

	ginkgo.By("deploying a second pod with cloned volume")
	tpod.Create()
	defer tpod.Cleanup()
	ginkgo.By("checking that the pod's command exits with no error")
	tpod.WaitForSuccess()
}
