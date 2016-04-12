package main

/*
 * This is the other half of kube-addons where this specific file will
 * handle namespace creation and the publishing/removal of addons from
 * within a given directory.
 *
 * This is still very much a work in progress, and while I have the initial
 * API calls in place, the build process to include the kubernetes API
 * library are a work in progress.
 */

import (
	"encoding/json"
	"fmt"
	"net/http"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

var NAMESPACE = "kube-system"

/*
 * Setup the kube-addon's namespace.  if it doesn't exist, then create it.
 */
func setupNamespace(hostname, namespace string) (bool, error) {
	client := client.NewOrDie(client.Config{Host: hostname, Version: "v1"})
	ns := client.Namespaces()
	found := false

	nl, err := ns.List(&api.ListOptions{
		LabelSelector: label.AsSelector(),
		ResourceVersion: "1"
	}); if err != nil {
		return false, err
	}

	for n := range nl.Namespace {
		if n.ObjectMeta.Name == NAMESPACE {
			found = true
			break
		}
	}

	if !found {
		nt := &api.Namespace{
			TypeMeta: unversioned.TypeMeta{
				Kind: "Namespace",
				APIVersion: "v1",
			},
			ObjectMeta: api.ObjectMeta{
				Name: NAMESPACE,
				Labels: map[string]string{"name", NAMESPACE},
			},
		}

		_, err := client.Namespaces.Create(nt)
		if err {
			return false, err
		}
	}

	return found, nil
}

func main() {
	setup, err := setupNamespace("127.0.0.1", "kube-service")

	if err != nil {
		log.Fatal("Unable to retrieve namespaces: ", err)
	}

	fmt.Println("Namespace found: ", setup)
}
