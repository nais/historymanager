package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/nais/historymanager/pkg/models"
	"github.com/nais/historymanager/pkg/persister"
	naisiov1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Root struct {
	Logger *zap.Logger
}

func (r *Root) history(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		fmt.Fprintf(w, "Only POST supported")
		return
	}

	defer req.Body.Close()
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var alertRequest models.AlertRequest
	err = json.Unmarshal(bodyBytes, &alertRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	out, err := alertRequest.ToBigQuery()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		r.Logger.Error("Can't convert alert to BigQquery model", zap.Any("alert", json.RawMessage(bodyBytes)), zap.Error(err))
		return
	}

	err = persister.Persist(context.Background(), out)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		r.Logger.Error("Can't persist alert to BigQuery", zap.Any("alert", json.RawMessage(bodyBytes)), zap.Error(err))
		return
	}

	r.Logger.Info("Alerts persisted", zap.Int("alerts", len(out)))
	fmt.Fprintln(w, "Alerts persisted")
}

func main() {
	r := Root{}
	logger, _ := zap.NewProduction()
	r.Logger = logger
	defer logger.Sync()

	goClient()

	logger.Info("The ancient books slowly crumbled, their secrets turning to dust. But their every word sings within the BigQuery's head.")

	http.HandleFunc("/history", r.history)
	http.ListenAndServe(":8090", nil)

}

func goClient() {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic(err.Error())
	}

	naisiov1.AddToScheme(scheme.Scheme)

	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: "nais.io", Version: "v1"}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	exampleRestClient, err := rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		panic(err)
	}

	result := naisiov1.AlertList{}
	err = exampleRestClient.
		Get().
		Resource("alerts").
		Do(context.Background()).
		Into(&result)
	if err != nil {
		panic(err.Error())
	}

	for _, a := range result.Items {
		fmt.Printf("%v-%v\n", a.Namespace, a.Name)
	}
}
