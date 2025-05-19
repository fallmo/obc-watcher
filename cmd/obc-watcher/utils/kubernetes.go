package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	obcv1alpha1 "github.com/kube-object-storage/lib-bucket-provisioner/pkg/apis/objectbucket.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var cs **dynamic.DynamicClient
var scheme = runtime.NewScheme()

const labelKey = "pending-bind-alert"

var OBCGroupVersionResource = schema.GroupVersionResource{
	Group:    "objectbucket.io",
	Version:  "v1alpha1",
	Resource: "objectbucketclaims",
}

func connectKubernetes() {
	var config *rest.Config
	_ = obcv1alpha1.AddToScheme(scheme)

	if os.Getenv("APP_ENV") == "development" {
		K8S_API_URL := os.Getenv("K8S_API_URL")
		K8S_API_TOKEN := os.Getenv("K8S_API_TOKEN")
		if K8S_API_URL == "" {
			log.Fatal("Variable 'K8S_API_URL' is required during development")
		}
		if K8S_API_TOKEN == "" {
			log.Fatal("Variable 'K8S_API_TOKEN' is required during development")
		}
		log.Println("Connecting to Kubernetes with environment variables..")

		config = &rest.Config{
			Host:        K8S_API_URL,
			BearerToken: K8S_API_TOKEN,
		}
	} else {
		log.Println("Connecting to Kubernetes from inside cluster..")
		cfg, err := rest.InClusterConfig()

		if err != nil {
			log.Fatal(err, "\nFailed to connect to Kubernetes..")
		}

		config = cfg
	}

	client, err := dynamic.NewForConfig(config)

	if err != nil {
		log.Fatal(err, "\nFailed to connect to Kubernetes..")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.Resource(OBCGroupVersionResource).Namespace("openshift").List(ctx, v1.ListOptions{})

	if err != nil {
		log.Fatalln(err, "\nFailed to connect to kubernetes")
	}

	log.Printf("Successfully connected to Kubernetes API...")

	cs = &client

}

func StartWatchingOBCs() {
	if cs == nil {
		log.Fatal("client is nil")
	}

	client := *cs

	watchInterface, err := client.Resource(OBCGroupVersionResource).Watch(context.TODO(), v1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			labelKey: "true",
		}).String(),
	})

	log.Println("Watching objectbucketclaims...")

	if err != nil {
		log.Fatal(err, "\nFailed to start watch stream")
	}

	for event := range watchInterface.ResultChan() {
		unstructuredObj := event.Object.(*unstructured.Unstructured)
		obc, err := convertToOBC(unstructuredObj)

		if err != nil {
			log.Fatalf("Failed to convert to OBC: %v", err)
		}

		fmt.Printf("New event of type %v for '%v' in namespace '%v'\n", event.Type, obc.Name, obc.Namespace)

		if event.Type == "DELETED" {
			fmt.Println("Ignoring event of type deleted")
			continue
		}

		if obc.Status.Phase != "Bound" {
			fmt.Println("OBC is not bound")
			continue
		}
		handleBoundOBC(&obc)
	}

	fmt.Print("Disconnected...")
}

func TriggerReconcile() {
	if cs == nil {
		log.Fatalln("Client is nil")
	}

	client := *cs

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data, err := client.Resource(OBCGroupVersionResource).List(
		ctx,
		v1.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				labelKey: "true",
			}).String(),
		},
	)

	if err != nil {
		log.Println(err, "\n Failed to list OBCs")
	}

	for i := 0; i < len(data.Items); i++ {
		obc, err := convertToOBC(&data.Items[i])

		if err != nil {
			log.Fatalln(err, "\n Failed to convert obj to obc")
		}

		if obc.Status.Phase != "Bound" {
			continue
		}

		handleBoundOBC(&obc)
	}
}

func StartOBCInformer() {
	if cs == nil {
		log.Fatal("No client")
	}

	client := *cs

	informer := dynamicinformer.NewFilteredDynamicInformer(
		client,
		OBCGroupVersionResource,
		"",
		time.Minute,
		cache.Indexers{
			cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
		},
		func(opts *v1.ListOptions) {
			opts.LabelSelector = labels.SelectorFromSet(map[string]string{
				labelKey: "true",
			}).String()
		},
	).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInInitialList bool) {
			obc, err := convertToOBC(obj.(*unstructured.Unstructured))

			if err != nil {
				log.Fatalln(err, "\n Failed to convert obj to obc")
			}

			fmt.Printf("New event of type %v for '%v' in namespace '%v'\n", "ADDED", obc.Name, obc.Namespace)

			if obc.Status.Phase != "Bound" {
				return
			}

			handleBoundOBC(&obc)
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			obc, err := convertToOBC(newObj.(*unstructured.Unstructured))

			if err != nil {
				log.Fatalln(err, "\n Failed to convert obj to obc")
			}

			fmt.Printf("New event of type %v for '%v' in namespace '%v'\n", "EDITED", obc.Name, obc.Namespace)

			if obc.Status.Phase != "Bound" {
				return
			}

			handleBoundOBC(&obc)
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("Ignoring delete event for obj")
		},
	})

	stopCh := make(chan struct{})
	defer close(stopCh)

	informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		fmt.Println("Informer is synced")
		return
	}

}

func handleBoundOBC(obj **obcv1alpha1.ObjectBucketClaim) {
	if cs == nil {
		log.Fatal("client is nil")
	}

	client := *cs
	obc := *obj

	fmt.Printf("OBC '%v' in '%v' is BOUND\n", obc.Name, obc.Namespace)

	PushAMQPMessage(
		amqpMessage{
			Kind: "objectbucketclaim-bound",
			Data: map[string]interface{}{
				"uuid":        string(obc.UID),
				"annotations": obc.Annotations,
			},
		},
	)

	patch := []map[string]interface{}{
		{
			"op":   "remove",
			"path": "/metadata/labels/" + labelKey,
		},
	}

	patchBytes, err := json.Marshal(patch)

	if err != nil {
		log.Fatal(err, "\nFailed to parse patch")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.Resource(OBCGroupVersionResource).Namespace(obc.Namespace).Patch(
		ctx,
		obc.Name,
		types.JSONPatchType,
		patchBytes,
		v1.PatchOptions{},
	)

	if err != nil {
		log.Fatal(err, "\nFailed to patch bound OBC")
	}

	log.Printf("Successfully acknowledged obc '%v' in namespace '%v' ", obc.Name, obc.Namespace)
}

func convertToOBC(obj *unstructured.Unstructured) (*obcv1alpha1.ObjectBucketClaim, error) {
	obc := &obcv1alpha1.ObjectBucketClaim{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, obc)
	return obc, err
}
