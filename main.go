package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type handler struct {
	clientset *kubernetes.Clientset
}

func main() {
	http.Handle("/", &handler{clientset: connectToK8s()})
	if err := http.ListenAndServe("0.0.0.0:6000", nil); err != nil {
		fmt.Println(err)
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	kind, ok := r.URL.Query()["kind"]
	if !ok {
		log.Println("Url Param 'kind' is missing")
		http.Error(w, "Url Param 'kind' is missing", http.StatusBadRequest)
		return
	}
	group, ok := r.URL.Query()["group"]
	if !ok {
		log.Println("Url Param 'group' is missing")
		http.Error(w, "Url Param 'group' is missing", http.StatusBadRequest)
		return
	}
	version, ok := r.URL.Query()["version"]
	if !ok {
		log.Println("Url Param 'version' is missing")
		http.Error(w, "Url Param 'version' is missing", http.StatusBadRequest)
		return
	}
	name, ok := r.URL.Query()["name"]
	if !ok {
		log.Println("Url Param 'name' is missing")
		http.Error(w, "Url Param 'name' is missing", http.StatusBadRequest)
		return
	}
	namespace, ok := r.URL.Query()["namespace"]
	if !ok {
		log.Println("Url Param 'namespace' is missing")
		http.Error(w, "Url Param 'namespace' is missing", http.StatusBadRequest)
		return
	}

	fmt.Println(r)
	fmt.Printf("Reconciling %s/%s %s, with name %s in namespace %s\n", group[0], version[0], kind[0], name[0], namespace[0])
	err := emitEvent(h.clientset, group[0], version[0], kind[0], name[0], namespace[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func emitEvent(cs *kubernetes.Clientset, group, version, kind, name, namespace string) error {
	events := cs.CoreV1().Events(namespace)
	eventSpec := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s.%d", name, time.Now().Unix()),
			Namespace: namespace,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:       kind,
			Namespace:  namespace,
			Name:       name,
			APIVersion: fmt.Sprintf("%s/%s", group, version),
		},
	}

	_, err := events.Create(context.TODO(), eventSpec, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	//print job details
	log.Println("Created K8s event successfully")

	return nil
}

func connectToK8s() *kubernetes.Clientset {
	home, exists := os.LookupEnv("HOME")
	if !exists {
		home = "/root"
	}

	configPath := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		log.Panicln("failed to create K8s config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panicln("Failed to create K8s clientset")
	}

	return clientset
}
