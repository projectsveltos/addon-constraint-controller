package controllers_test

import (
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/projectsveltos/addon-compliance-controller/controllers"
	"github.com/projectsveltos/addon-compliance-controller/pkg/scope"
	libsveltosv1alpha1 "github.com/projectsveltos/libsveltos/api/v1alpha1"
	libsveltosset "github.com/projectsveltos/libsveltos/lib/set"
)

var (
	scheme *runtime.Scheme
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controllers Suite")
}

var _ = BeforeSuite(func() {
	By("bootstrapping test environment")

	var err error
	scheme, err = setupScheme()
	Expect(err).To(BeNil())
})

func getAddonComplianceScope(c client.Client, logger logr.Logger,
	addonConstraint *libsveltosv1alpha1.AddonCompliance,
) *scope.AddonComplianceScope {

	addonConstraintScope, err := scope.NewAddonComplianceScope(scope.AddonComplianceScopeParams{
		Client:          c,
		Logger:          logger,
		AddonCompliance: addonConstraint,
		ControllerName:  "addoncompliance",
	})
	Expect(err).To(BeNil())
	return addonConstraintScope
}

func getAddonComplianceReconciler(c client.Client) *controllers.AddonComplianceReconciler {
	return &controllers.AddonComplianceReconciler{
		Client:                        c,
		Scheme:                        scheme,
		AddonCompliances:              make(map[types.NamespacedName]libsveltosv1alpha1.Selector),
		ClusterLabels:                 make(map[corev1.ObjectReference]map[string]string),
		ClusterMap:                    make(map[corev1.ObjectReference]*libsveltosset.Set),
		AddonComplianceToClusterMap:   make(map[types.NamespacedName]*libsveltosset.Set),
		ReferenceMap:                  make(map[corev1.ObjectReference]*libsveltosset.Set),
		AddonComplianceToReferenceMap: make(map[types.NamespacedName]*libsveltosset.Set),
		PolicyMux:                     sync.Mutex{},
	}
}

func randomString() string {
	const length = 10
	return util.RandomString(length)
}

func setupScheme() (*runtime.Scheme, error) {
	s := runtime.NewScheme()
	if err := clusterv1.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := clientgoscheme.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := apiextensionsv1.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := libsveltosv1alpha1.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := sourcev1.AddToScheme(s); err != nil {
		return nil, err
	}

	return s, nil
}
