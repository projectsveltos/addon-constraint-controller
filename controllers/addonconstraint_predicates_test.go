/*
Copyright 2023. projectsveltos.io. All rights reserved.

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

package controllers_test

import (
	"encoding/base64"

	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2/klogr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/projectsveltos/addon-constraint-controller/controllers"
	libsveltosv1alpha1 "github.com/projectsveltos/libsveltos/api/v1alpha1"
)

var _ = Describe("AddonConstraint Predicates: SvelotsClusterPredicates", func() {
	var logger logr.Logger
	var cluster *libsveltosv1alpha1.SveltosCluster

	BeforeEach(func() {
		logger = klogr.New()
		cluster = &libsveltosv1alpha1.SveltosCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      randomString(),
				Namespace: "predicates" + randomString(),
			},
		}
	})

	It("Create reprocesses when sveltos Cluster is unpaused", func() {
		clusterPredicate := controllers.SveltosClusterPredicates(logger)

		cluster.Spec.Paused = false

		e := event.CreateEvent{
			Object: cluster,
		}

		result := clusterPredicate.Create(e)
		Expect(result).To(BeTrue())
	})
	It("Create does not reprocess when sveltos Cluster is paused", func() {
		clusterPredicate := controllers.SveltosClusterPredicates(logger)

		cluster.Spec.Paused = true
		cluster.Annotations = map[string]string{clusterv1.PausedAnnotation: "true"}

		e := event.CreateEvent{
			Object: cluster,
		}

		result := clusterPredicate.Create(e)
		Expect(result).To(BeFalse())
	})
	It("Delete does reprocess ", func() {
		clusterPredicate := controllers.SveltosClusterPredicates(logger)

		e := event.DeleteEvent{
			Object: cluster,
		}

		result := clusterPredicate.Delete(e)
		Expect(result).To(BeTrue())
	})
	It("Update reprocesses when sveltos Cluster paused changes from true to false", func() {
		clusterPredicate := controllers.SveltosClusterPredicates(logger)

		cluster.Spec.Paused = false

		oldCluster := &libsveltosv1alpha1.SveltosCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
		}
		oldCluster.Spec.Paused = true
		oldCluster.Annotations = map[string]string{clusterv1.PausedAnnotation: "true"}

		e := event.UpdateEvent{
			ObjectNew: cluster,
			ObjectOld: oldCluster,
		}

		result := clusterPredicate.Update(e)
		Expect(result).To(BeTrue())
	})
	It("Update does not reprocess when sveltos Cluster paused changes from false to true", func() {
		clusterPredicate := controllers.SveltosClusterPredicates(logger)

		cluster.Spec.Paused = true
		cluster.Annotations = map[string]string{clusterv1.PausedAnnotation: "true"}
		oldCluster := &libsveltosv1alpha1.SveltosCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
		}
		oldCluster.Spec.Paused = false

		e := event.UpdateEvent{
			ObjectNew: cluster,
			ObjectOld: oldCluster,
		}

		result := clusterPredicate.Update(e)
		Expect(result).To(BeFalse())
	})
	It("Update does not reprocess when sveltos Cluster paused has not changed", func() {
		clusterPredicate := controllers.SveltosClusterPredicates(logger)

		cluster.Spec.Paused = false
		oldCluster := &libsveltosv1alpha1.SveltosCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
		}
		oldCluster.Spec.Paused = false

		e := event.UpdateEvent{
			ObjectNew: cluster,
			ObjectOld: oldCluster,
		}

		result := clusterPredicate.Update(e)
		Expect(result).To(BeFalse())
	})
	It("Update reprocesses when sveltos Cluster labels change", func() {
		clusterPredicate := controllers.SveltosClusterPredicates(logger)

		cluster.Labels = map[string]string{"department": "eng"}

		oldCluster := &libsveltosv1alpha1.SveltosCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
				Labels:    map[string]string{},
			},
		}

		e := event.UpdateEvent{
			ObjectNew: cluster,
			ObjectOld: oldCluster,
		}

		result := clusterPredicate.Update(e)
		Expect(result).To(BeTrue())
	})
	It("Update reprocesses when sveltos Cluster Status Ready changes", func() {
		clusterPredicate := controllers.SveltosClusterPredicates(logger)

		cluster.Status.Ready = true

		oldCluster := &libsveltosv1alpha1.SveltosCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
				Labels:    map[string]string{},
			},
			Status: libsveltosv1alpha1.SveltosClusterStatus{
				Ready: false,
			},
		}

		e := event.UpdateEvent{
			ObjectNew: cluster,
			ObjectOld: oldCluster,
		}

		result := clusterPredicate.Update(e)
		Expect(result).To(BeTrue())
	})
})

var _ = Describe("AddonConstraint Predicates: ClusterPredicates", func() {
	var logger logr.Logger
	var cluster *clusterv1.Cluster

	BeforeEach(func() {
		logger = klogr.New()
		cluster = &clusterv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      randomString(),
				Namespace: "predicates" + randomString(),
			},
		}
	})

	It("Create reprocesses when v1Cluster is unpaused", func() {
		clusterPredicate := controllers.ClusterPredicates(logger)

		cluster.Spec.Paused = false

		e := event.CreateEvent{
			Object: cluster,
		}

		result := clusterPredicate.Create(e)
		Expect(result).To(BeTrue())
	})
	It("Create does not reprocess when v1Cluster is paused", func() {
		clusterPredicate := controllers.ClusterPredicates(logger)

		cluster.Spec.Paused = true
		cluster.Annotations = map[string]string{clusterv1.PausedAnnotation: "true"}

		e := event.CreateEvent{
			Object: cluster,
		}

		result := clusterPredicate.Create(e)
		Expect(result).To(BeFalse())
	})
	It("Delete does reprocess ", func() {
		clusterPredicate := controllers.ClusterPredicates(logger)

		e := event.DeleteEvent{
			Object: cluster,
		}

		result := clusterPredicate.Delete(e)
		Expect(result).To(BeTrue())
	})
	It("Update reprocesses when v1Cluster paused changes from true to false", func() {
		clusterPredicate := controllers.ClusterPredicates(logger)

		cluster.Spec.Paused = false

		oldCluster := &clusterv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
		}
		oldCluster.Spec.Paused = true
		oldCluster.Annotations = map[string]string{clusterv1.PausedAnnotation: "true"}

		e := event.UpdateEvent{
			ObjectNew: cluster,
			ObjectOld: oldCluster,
		}

		result := clusterPredicate.Update(e)
		Expect(result).To(BeTrue())
	})
	It("Update does not reprocess when v1Cluster paused changes from false to true", func() {
		clusterPredicate := controllers.ClusterPredicates(logger)

		cluster.Spec.Paused = true
		cluster.Annotations = map[string]string{clusterv1.PausedAnnotation: "true"}
		oldCluster := &clusterv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
		}
		oldCluster.Spec.Paused = false

		e := event.UpdateEvent{
			ObjectNew: cluster,
			ObjectOld: oldCluster,
		}

		result := clusterPredicate.Update(e)
		Expect(result).To(BeFalse())
	})
	It("Update does not reprocess when v1Cluster paused has not changed", func() {
		clusterPredicate := controllers.ClusterPredicates(logger)

		cluster.Spec.Paused = false
		oldCluster := &clusterv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
		}
		oldCluster.Spec.Paused = false

		e := event.UpdateEvent{
			ObjectNew: cluster,
			ObjectOld: oldCluster,
		}

		result := clusterPredicate.Update(e)
		Expect(result).To(BeFalse())
	})
	It("Update reprocesses when v1Cluster labels change", func() {
		clusterPredicate := controllers.ClusterPredicates(logger)

		cluster.Labels = map[string]string{"department": "eng"}

		oldCluster := &clusterv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
				Labels:    map[string]string{},
			},
		}

		e := event.UpdateEvent{
			ObjectNew: cluster,
			ObjectOld: oldCluster,
		}

		result := clusterPredicate.Update(e)
		Expect(result).To(BeTrue())
	})
})

var _ = Describe("AddonConstraint Predicates: MachinePredicates", func() {
	var logger logr.Logger
	var machine *clusterv1.Machine

	BeforeEach(func() {
		logger = klogr.New()
		machine = &clusterv1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      randomString(),
				Namespace: "predicates" + randomString(),
			},
		}
	})

	It("Create reprocesses when v1Machine is Running", func() {
		machinePredicate := controllers.MachinePredicates(logger)

		machine.Status.Phase = string(clusterv1.MachinePhaseRunning)

		e := event.CreateEvent{
			Object: machine,
		}

		result := machinePredicate.Create(e)
		Expect(result).To(BeTrue())
	})
	It("Create does not reprocess when v1Machine is not Running", func() {
		machinePredicate := controllers.MachinePredicates(logger)

		e := event.CreateEvent{
			Object: machine,
		}

		result := machinePredicate.Create(e)
		Expect(result).To(BeFalse())
	})
	It("Delete does not reprocess ", func() {
		machinePredicate := controllers.MachinePredicates(logger)

		e := event.DeleteEvent{
			Object: machine,
		}

		result := machinePredicate.Delete(e)
		Expect(result).To(BeFalse())
	})
	It("Update reprocesses when v1Machine Phase changed from not running to running", func() {
		machinePredicate := controllers.MachinePredicates(logger)

		machine.Status.Phase = string(clusterv1.MachinePhaseRunning)

		oldMachine := &clusterv1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      machine.Name,
				Namespace: machine.Namespace,
			},
		}

		e := event.UpdateEvent{
			ObjectNew: machine,
			ObjectOld: oldMachine,
		}

		result := machinePredicate.Update(e)
		Expect(result).To(BeTrue())
	})
	It("Update does not reprocess when v1Machine Phase changes from not Phase not set to Phase set but not running", func() {
		machinePredicate := controllers.MachinePredicates(logger)

		machine.Status.Phase = "Provisioning"

		oldMachine := &clusterv1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      machine.Name,
				Namespace: machine.Namespace,
			},
		}

		e := event.UpdateEvent{
			ObjectNew: machine,
			ObjectOld: oldMachine,
		}

		result := machinePredicate.Update(e)
		Expect(result).To(BeFalse())
	})
	It("Update does not reprocess when v1Machine Phases does not change", func() {
		machinePredicate := controllers.MachinePredicates(logger)
		machine.Status.Phase = string(clusterv1.MachinePhaseRunning)

		oldMachine := &clusterv1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      machine.Name,
				Namespace: machine.Namespace,
			},
		}
		oldMachine.Status.Phase = machine.Status.Phase

		e := event.UpdateEvent{
			ObjectNew: machine,
			ObjectOld: oldMachine,
		}

		result := machinePredicate.Update(e)
		Expect(result).To(BeFalse())
	})
})

var _ = Describe("AddonConstraint Predicates: ConfigMapPredicates", func() {
	var logger logr.Logger
	var configMap *corev1.ConfigMap

	BeforeEach(func() {
		logger = klogr.New()
		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
			},
		}
	})

	It("Create returns true", func() {
		configMapPredicate := controllers.ConfigMapPredicates(logger)

		e := event.CreateEvent{
			Object: configMap,
		}

		result := configMapPredicate.Create(e)
		Expect(result).To(BeTrue())
	})

	It("Delete returns true", func() {
		configMapPredicate := controllers.ConfigMapPredicates(logger)

		e := event.DeleteEvent{
			Object: configMap,
		}

		result := configMapPredicate.Delete(e)
		Expect(result).To(BeTrue())
	})

	It("Update returns true when Data has changed", func() {
		configMapPredicate := controllers.ConfigMapPredicates(logger)
		configMap.Data = map[string]string{randomString(): randomString()}

		oldConfigMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: configMap.Name,
			},
		}

		e := event.UpdateEvent{
			ObjectNew: configMap,
			ObjectOld: oldConfigMap,
		}

		result := configMapPredicate.Update(e)
		Expect(result).To(BeTrue())
	})

	It("Update returns true when BinaryData has changed", func() {
		configMapPredicate := controllers.ConfigMapPredicates(logger)
		configMap.BinaryData = map[string][]byte{randomString(): []byte(randomString())}

		oldConfigMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: configMap.Name,
			},
		}

		e := event.UpdateEvent{
			ObjectNew: configMap,
			ObjectOld: oldConfigMap,
		}

		result := configMapPredicate.Update(e)
		Expect(result).To(BeTrue())
	})

	It("Update returns false when BinaryData has not changed", func() {
		configMapPredicate := controllers.ConfigMapPredicates(logger)
		configMap.BinaryData = map[string][]byte{randomString(): []byte(randomString())}

		oldConfigMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:   configMap.Name,
				Labels: map[string]string{"env": "testing"},
			},
			BinaryData: configMap.BinaryData,
		}

		e := event.UpdateEvent{
			ObjectNew: configMap,
			ObjectOld: oldConfigMap,
		}

		result := configMapPredicate.Update(e)
		Expect(result).To(BeFalse())
	})
})

var _ = Describe("Clustersummary Predicates: SecretPredicates", func() {
	var logger logr.Logger
	var secret *corev1.Secret

	BeforeEach(func() {
		logger = klogr.New()
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
			},
		}
	})

	It("Create returns true", func() {
		secretPredicate := controllers.SecretPredicates(logger)

		e := event.CreateEvent{
			Object: secret,
		}

		result := secretPredicate.Create(e)
		Expect(result).To(BeTrue())
	})

	It("Delete returns true", func() {
		secretPredicate := controllers.SecretPredicates(logger)

		e := event.DeleteEvent{
			Object: secret,
		}

		result := secretPredicate.Delete(e)
		Expect(result).To(BeTrue())
	})

	It("Update returns true when data has changed", func() {
		secretPredicate := controllers.SecretPredicates(logger)
		str := base64.StdEncoding.EncodeToString([]byte("password"))
		secret.Data = map[string][]byte{"change": []byte(str)}

		oldSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: secret.Name,
			},
		}

		e := event.UpdateEvent{
			ObjectNew: secret,
			ObjectOld: oldSecret,
		}

		result := secretPredicate.Update(e)
		Expect(result).To(BeTrue())
	})

	It("Update returns false when Data has not changed", func() {
		secretPredicate := controllers.SecretPredicates(logger)

		oldSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   secret.Name,
				Labels: map[string]string{"env": "testing"},
			},
		}

		e := event.UpdateEvent{
			ObjectNew: secret,
			ObjectOld: oldSecret,
		}

		result := secretPredicate.Update(e)
		Expect(result).To(BeFalse())
	})
})

var _ = Describe("AddonConstraint Predicates: FluxSourcePredicates", func() {
	var logger logr.Logger
	var gitRepository *sourcev1.GitRepository

	BeforeEach(func() {
		logger = klogr.New()
		gitRepository = &sourcev1.GitRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "flux" + randomString(),
				Namespace: "predicates" + randomString(),
			},
		}

		controllers.AddTypeInformationToObject(scheme, gitRepository)
	})

	It("Create reprocesses", func() {
		sourcePredicate := controllers.FluxSourcePredicates(scheme, logger)

		e := event.CreateEvent{
			Object: gitRepository,
		}

		result := sourcePredicate.Create(e)
		Expect(result).To(BeTrue())
	})
	It("Delete does reprocess", func() {
		sourcePredicate := controllers.FluxSourcePredicates(scheme, logger)

		e := event.DeleteEvent{
			Object: gitRepository,
		}

		result := sourcePredicate.Delete(e)
		Expect(result).To(BeTrue())
	})
	It("Update reprocesses when artifact has changed", func() {
		sourcePredicate := controllers.FluxSourcePredicates(scheme, logger)

		gitRepository.Status.Artifact = &sourcev1.Artifact{
			Revision: randomString(),
		}

		oldGitRepository := &sourcev1.GitRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      gitRepository.Name,
				Namespace: gitRepository.Namespace,
			},
		}

		controllers.AddTypeInformationToObject(scheme, oldGitRepository)

		e := event.UpdateEvent{
			ObjectNew: gitRepository,
			ObjectOld: oldGitRepository,
		}

		result := sourcePredicate.Update(e)
		Expect(result).To(BeTrue())
	})
	It("Update does not reprocess when artifact has not changed", func() {
		sourcePredicate := controllers.FluxSourcePredicates(scheme, logger)

		gitRepository.Status.Artifact = &sourcev1.Artifact{
			Revision: randomString(),
		}

		oldGitRepository := &sourcev1.GitRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      gitRepository.Name,
				Namespace: gitRepository.Namespace,
			},
		}
		oldGitRepository.Status.Artifact = gitRepository.GetArtifact()

		controllers.AddTypeInformationToObject(scheme, oldGitRepository)

		e := event.UpdateEvent{
			ObjectNew: gitRepository,
			ObjectOld: oldGitRepository,
		}

		result := sourcePredicate.Update(e)
		Expect(result).To(BeFalse())
	})
})
