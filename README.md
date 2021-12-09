# gabi-yml-to-terraform

YML File to describe the environment that will be transformed to a terraform file to allow automatic deployment

It will generate the main.tf needed by terraform, and an import script to allow to synchronize the cloud with the
current main

### Steps

- Create the project on GCP
- Create a service account with role owner
- Log in with gcloud tool into your project 
- run ````make prepare```` or activate the needed api (See Makefile)
- run ````make build````
- run ````./main -output . -input sample.yml````

### Sample Yml

````
gabi_project:
  gcp:
    name: "project"
    location: "europe-west1"
    firebase: yes
  api:
    - "iam.googleapis.com"
    - "pubsub.googleapis.com"
  service_accounts:
    - name: "account1"
      roles:
        - "roles/pubsub.editor"
  cloud_runs:
    - name: "sample service"
      service_account: "account1"
      allow_unauthenticated: no
      location: "europe-west1"
      memory: "512Mi"
      cpu: 1
      max_instances: 1
      image: "eu.gcr.io/{gabi_project.gcp.name}/service1:a9d8aae4"
      env:
        GCP_PROJECT: "{gabi_project.gcp.name}"
      secret:
        POSTGRES_PASSWORD: "POSTGRES_PASSWORD"
  pubsub:
    topics:
      - name: "hcp"
        subscriptions:
          - name: "pubsub_subscription"
            cloud_run: "account1"
  storage:
    gcp:
      - name: "{gabi_project.gcp.name}-backend"
        location: "EU"
  sql:
    name: "cloud_sql_database"
    version: "POSTGRES_13"
    location: "europe-west1"
    tier: "db-f1-micro"
    user: "postgres"
    password: "postgres"
  secret:
    location: "europe-west1"
    values:
      POSTGRES_PASSWORD: "{gabi_project.sql.password}"

````
