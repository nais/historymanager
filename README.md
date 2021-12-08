# historymanager
Let's scrape the Alertmanager history


## Prerequisites

Tools needed: `gcloud`, og `bq`

```
export PROJECT_ID=aura-prod-d7e3
export EMAIL=alert-history
export DATASET=alert_history
export TABLE=alerts
gcloud iam --project "$PROJECT_ID" \
  service-accounts create "$EMAIL" \
  --description="Manually created service-account for $EMAIL"

bq mk --project_id "$PROJECT_ID" --location europe-north1 "$DATASET"

bq mk --table "$PROJECT_ID:$DATASET.$TABLE" schema.json

bq add-iam-policy-binding \
  --member="serviceAccount:$EMAIL@$PROJECT_ID.iam.gserviceaccount.com" \
  --role=roles/bigquery.dataEditor \
  "$PROJECT_ID:$DATASET.$TABLE"

gcloud iam service-accounts keys create sa.json \
    --iam-account=$EMAIL@$PROJECT_ID.iam.gserviceaccount.com
```
