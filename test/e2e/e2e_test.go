// +build e2e

package e2e

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	nfdv1alpha1 "github.com/openshift/cluster-nfd-operator/pkg/apis/nfd/v1alpha1"
	nfdclient "github.com/openshift/cluster-nfd-operator/pkg/client"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	crdc "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	deploymentTimeout = 5 * time.Minute
	apiTimeout        = 10 * time.Second
)

type assetsFromFile []byte

var manifests []assetsFromFile

func TestCreateOperator(t *testing.T) {

	clientSet, err := nfdclient.GetClientSet()
	if clientSet == nil {
		t.Errorf("failed to get a clientSet: %v", err)
	}

	cfgv1client, err := nfdclient.GetCfgV1Client()
	if cfgv1client == nil {
		t.Errorf("failed to get a cfgv1client: %v", err)
	}

	apiclient, err := nfdclient.GetApiClient()
	if apiclient == nil {
		t.Errorf("failed to get a apiclient: %v", err)
	}

	crdclient, err := nfdclient.NewClient()
	if crdclient == nil {
		t.Errorf("failed to get a crdclient: %v", err)
	}

	path := "manifests"
	namespace := "openshift-nfd-operator"

	manifests := getAssetsFrom(path)

	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme,
		scheme.Scheme)
	reg, _ := regexp.Compile(`\b(\w*kind:\w*)\B.*\b`)

	for _, m := range manifests {
		kind := reg.FindString(string(m))
		slce := strings.Split(kind, ":")
		kind = strings.TrimSpace(slce[1])

		switch kind {
		case "Namespace":
			var res corev1.Namespace
			t.Logf("Resource: Kind %s", kind)
			_, _, err := s.Decode(m, nil, &res)

			client := clientSet.CoreV1().Namespaces()

			

			_, err = client.Get(res.GetName(), metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				_, err = client.Create(&res)
				if err != nil {
					t.Errorf("Creating %s %s object failed", kind, res.GetName())
				}
			}
			if err != nil {
				t.Errorf("Retrieving %s %s object failed", kind, res.GetName())
			}

			_, err = client.Update(&res)
			if err != nil {
				t.Errorf("Updating %s %s object failed", kind, res.GetName())
			}

		case "ServiceAccount":
			var res corev1.ServiceAccount
			t.Logf("Resource: Kind %s", kind)
			_, _, err := s.Decode(m, nil, &res)

			client := clientSet.CoreV1().ServiceAccounts(namespace)

			_, err = client.Get(res.GetName(), metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				_, err = client.Create(&res)
				if err != nil {
					t.Errorf("Creating %s %s object failed", kind, res.GetName())
				}
			}
			if err != nil {
				t.Errorf("Retrieving %s %s object failed", kind, res.GetName())
			}

			_, err = client.Update(&res)
			if err != nil {
				t.Errorf("Updating %s %s object failed", kind, res.GetName())
			}

		case "ClusterRole":
			var res rbacv1.ClusterRole
			t.Logf("Resource: Kind %s", kind)
			_, _, err := s.Decode(m, nil, &res)

			client := clientSet.RbacV1().ClusterRoles()

			_, err = client.Get(res.GetName(), metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				_, err = client.Create(&res)
				if err != nil {
					t.Errorf("Creating %s %s object failed", kind, res.GetName())
				}
			}
			if err != nil {
				t.Errorf("Retrieving %s %s object failed", kind, res.GetName())
			}

			_, err = client.Update(&res)
			if err != nil {
				t.Errorf("Updating %s %s object failed", kind, res.GetName())
			}
		case "ClusterRoleBinding":
			var res rbacv1.ClusterRoleBinding
			t.Logf("Resource: Kind %s", kind)
			_, _, err := s.Decode(m, nil, &res)
			client := clientSet.RbacV1().ClusterRoleBindings()

			_, err = client.Get(res.GetName(), metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				_, err = client.Create(&res)
				if err != nil {
					t.Errorf("Creating %s %s object failed", kind, res.GetName())
				}
			}
			if err != nil {
				t.Errorf("Retrieving %s %s object failed", kind, res.GetName())
			}

			_, err = client.Update(&res)
			if err != nil {
				t.Errorf("Updating %s %s object failed", kind, res.GetName())
			}
		case "CustomResourceDefinition":
			var res crdc.CustomResourceDefinition
			t.Logf("Resource: Kind %s", kind)
			_, _, err := s.Decode(m, nil, &res)

			t.Logf("CRD: %s", string(m))

			client := apiclient.ApiextensionsV1beta1().CustomResourceDefinitions()

			existing, err := client.Get(res.GetName(), metav1.GetOptions{})

			t.Logf("EXISTING: %s", string(existing.GetName()))

			if apierrors.IsNotFound(err) {
				_, err = client.Create(&res)
				if err != nil {
					t.Errorf("Creating %s %s object failed", kind, res.GetName())
				}
				t.Logf("Resource: Kind %s created", kind)
			}
			if err != nil {
				t.Errorf("Retrieving %s %s object failed", kind, res.GetName())
			}

			required := res.DeepCopy()
			required.ResourceVersion = existing.ResourceVersion

			_, err = client.Update(required)
			if err != nil {
				t.Errorf("Updating %s %s object failed: %s", kind, res.GetName(), err)
			}
			t.Logf("Resource: Kind %s updated", kind)

		case "Deployment":
			var res appsv1.Deployment
			t.Logf("Resource: Kind %s", kind)
			_, _, err := s.Decode(m, nil, &res)

			client := clientSet.AppsV1().Deployments(namespace)

			_, err = client.Get(res.GetName(), metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				_, err = client.Create(&res)
				if err != nil {
					t.Errorf("Creating %s %s object failed", kind, res.GetName())
				}
			}
			if err != nil {
				t.Errorf("Retrieving %s %s object failed", kind, res.GetName())
			}

			_, err = client.Update(&res)
			if err != nil {
				t.Errorf("Updating %s %s object failed", kind, res.GetName())
			}

		case "NodeFeatureDiscovery!!":
			var res nfdv1alpha1.NodeFeatureDiscovery
			t.Logf("Resource: Kind %s", kind)
			_, _, err := s.Decode(m, nil, &res)

			existing, err := crdclient.NodeFeatureDiscoveries("openshift-nfd").Get("nfd-master-server")
			if apierrors.IsNotFound(err) {
				_, err = crdclient.NodeFeatureDiscoveries("openshift-nfd").Create(&res)
				if err != nil {
					t.Errorf("Creating %s %s object failed", kind, res.GetName())
				}
			}
			if err != nil {
				t.Errorf("Retrieving %s %s object failed", kind, res.GetName())
			}

			required := res.DeepCopy()
			required.ResourceVersion = existing.ResourceVersion

			//			_, err = crdclient.NodeFeatureDiscoveries("openshift-nfd").Update(required)
			//			if err != nil {
			//				t.Errorf("Updating %s %s object failed", kind, res.GetName())
			//			}
		default:
			t.Errorf("Unknown Resource: Kind %s", kind)
		}
	}
}

func filePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func getAssetsFrom(path string) []assetsFromFile {

	manifests := []assetsFromFile{}
	assets := path
	files, err := filePathWalkDir(assets)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		buffer, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		manifests = append(manifests, buffer)
	}
	return manifests
}
