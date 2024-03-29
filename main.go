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
	"go.uber.org/zap"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
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

func validateNecessaryEnvs() {
	if os.Getenv("NAIS_CLUSTER_NAME") == "" {
		panic("Missing env NAIS_CLUSTER_NAME")
	}

	if os.Getenv("PROJECT_ID") == "" {
		panic("Missing env PROJECT_ID")
	}
}

func main() {
	validateNecessaryEnvs()

	r := Root{}
	logger, _ := zap.NewProduction()
	r.Logger = logger
	defer logger.Sync()

	logger.Info("The ancient books slowly crumbled, their secrets turning to dust. But their every word sings within the BigQuery's head.")

	http.HandleFunc("/history", r.history)
	http.ListenAndServe(":8090", nil)

}
