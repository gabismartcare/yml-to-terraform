package main

func getTemplate() string {
	return `terraform {
  backend "http" {
  }
}
{{ $projectName:=.GabiProject.Gcp.Name}}
{{ $sqlName:= fillEnv .GabiProject.Sql.Name}}
{{ $sqlResource:= formatAction "sql" $sqlName }}
provider "google" {
	credentials = file("{{ .GabiProject.Gcp.ServiceKeyFile }}")
	project     = {{ fillEnv .GabiProject.Gcp.Name }}
	region      = "{{ .GabiProject.Gcp.Location }}"
}
provider "google-beta" {
	credentials = file("{{ .GabiProject.Gcp.ServiceKeyFile }}")
	project     = {{ fillEnv .GabiProject.Gcp.Name }}
	region      = "{{ .GabiProject.Gcp.Location }}"
}
data "google_project" "project" {}
{{range .GabiProject.Apis }}
resource "google_project_service" "{{formatAction "gcp_services" .}}" {
  project = "{{ $projectName }}"
  service = "{{.}}"
  disable_on_destroy = false
}
{{end}}
{{ range .GabiProject.ServiceAccounts }}
    {{ $name:=.Name }}

    resource "google_service_account" "{{ formatAction "service_account" $name}}" {
      project = "{{ $projectName }}"
      account_id = "{{ $name }}"
      display_name = "{{ $name }}"
    }
{{end}}
{{ range .GabiProject.CustomRole }}
	resource "google_project_iam_custom_role" "{{ formatAction "custom_role" .Name}}" {
    role_id     = "{{ .Name }}"
    title       = "{{ .Name }}"
    permissions = [{{ join .Roles }}]
}
{{end}}
{{ range $role, $account := .ServiceAccountMap }}
        resource "google_project_iam_binding" "{{formatAction "role" $role}}" {
          project = "{{ $projectName }}"
          role    =  "{{ $role }}"
          members = [
           {{range $account }}"serviceAccount:{{.}}@{{ $projectName }}.iam.gserviceaccount.com",
           {{end}}
          ]
        }
{{ end }}

{{if .GabiProject.Sql }}
    resource "google_sql_database_instance" "{{ $sqlResource }}" {
      name             =  {{ $sqlName}}
      database_version = "{{.GabiProject.Sql.Version}}"
      region           = "{{.GabiProject.Sql.Location }}"

      settings {
        tier = "{{ .GabiProject.Sql.Tier }}"
        {{ if .GabiProject.Sql.Configuration }}
            {{ if .GabiProject.Sql.Configuration.QueryInsight }}
            insights_config {
                query_insights_enabled = true
             }
            {{ end }}
	   
            {{ if .GabiProject.Sql.Configuration.SSL }}
			ip_configuration {
				require_ssl = true 
			}
			{{end}}
            {{ if .GabiProject.Sql.Configuration.AutoStorageIncrease }} disk_autoresize = true {{ end }}
            {{ if .GabiProject.Sql.Configuration.HighAvailability }} availability_type = "REGIONAL" 
			backup_configuration {
			  enabled = true
			}
		{{ end }}
            {{ if .GabiProject.Sql.Configuration.MaintenanceWindow }}
             maintenance_window {
                day ={{ .GabiProject.Sql.Configuration.MaintenanceWindow.Day }}
                hour ={{ .GabiProject.Sql.Configuration.MaintenanceWindow.Hour }}
             }
             {{ end }}
             {{ if .GabiProject.Sql.Configuration.BackupSqlOption }}
             backup_configuration {
                enabled = true
                location = "{{ .GabiProject.Sql.Configuration.BackupSqlOption.Region }}"
                backup_retention_settings {
                    retained_backups = "{{ .GabiProject.Sql.Configuration.BackupSqlOption.Duration }}"
                }
             }
             {{ end }}
        {{ end }}
      }
    }
  {{if .GabiProject.Sql.Database }}
  resource "google_sql_database" "{{ formatAction "sqldb" $sqlName }}" {
    name     = "{{.GabiProject.Sql.Database}}"
    instance = google_sql_database_instance.{{ $sqlResource }}.name
  }
  {{end}}

    resource "google_sql_user" "{{ formatAction  "sql_users"  $sqlName }}" {
      name     = "{{  .GabiProject.Sql.User }}"
      instance = google_sql_database_instance.{{ $sqlResource }}.name
      password = "{{ .GabiProject.Sql.Password }}"
    }
{{end}}

{{ if .GabiProject.Secret }}
    {{ $loc:= .GabiProject.Secret.Location }}
    {{ range $key, $value := .GabiProject.Secret.Values }}
        resource "google_secret_manager_secret" "{{formatAction "secret" $key }}" {
          secret_id = "{{$key}}"
          replication {
            user_managed {
              replicas {
                location = "{{$loc }}"
              }
            }
          }
        }
        resource "google_secret_manager_secret_version" "{{formatAction "secret_version" $key }}" {
          secret = google_secret_manager_secret.{{formatAction "secret" $key }}.id
          secret_data = {{fillEnv $value}}
        }
    {{ end }}
{{ end }}

{{ range .GabiProject.CloudRuns }}		
    {{ $serviceAccount := .ServiceAccount }}
    {{if .Secret }}
        {{ range $key, $value := .Secret }}
            resource "google_secret_manager_secret_iam_member" "{{formatAction "secret_access" (formatAction $serviceAccount $key)   }}" {
              secret_id = google_secret_manager_secret.{{formatAction "secret" $value }}.id
              role      = "roles/secretmanager.secretAccessor"
              member    = "serviceAccount:{{ $serviceAccount }}@{{ $projectName }}.iam.gserviceaccount.com"
              depends_on = [google_secret_manager_secret.{{formatAction "secret" $value }}, google_service_account.{{ formatAction "service_account" $serviceAccount}}]
            }
        {{end }}
    {{ end }}
    resource "google_cloud_run_service" "{{getResourceName "cloudrun" .Name }}" {
    name     = "{{ .Name }}"
    location = "{{ .Location }}"
	
	depends_on =[{{if .Secret }}{{ range $key, $value := .Secret }}
	 	google_secret_manager_secret_iam_member.{{formatAction "secret_access" (formatAction $serviceAccount $key)   }},
		{{end}}{{end}}
	 google_service_account.{{ formatAction "service_account" .ServiceAccount}}
	]
  	autogenerate_revision_name = true
    traffic {
      percent         = 100
	  latest_revision = true
    }
	template {
     spec {
      service_account_name = "{{ .ServiceAccount }}@{{ $projectName }}.iam.gserviceaccount.com"
      containers {
        image = {{ fillEnv .Image }}
        {{ range $key, $value := .Env }}env {
          name = "{{ $key }}"
          value = {{ fillEnv $value }}
        }
        {{end}}  {{ range $key, $value := .Secret }}env {
          name = "{{ $key }}"
          value_from {
              secret_key_ref {
                name = google_secret_manager_secret.{{formatAction "secret" $value }}.secret_id
                key = "latest"
            }
          }
        }
        {{end}}
        }
  }
   metadata {
        annotations = {
          "autoscaling.knative.dev/maxScale"      = "{{ if eq (.MaxInstances) 0 }}10{{else}}{{.MaxInstances}}{{end}}"
          {{if .SqlInstance}}"run.googleapis.com/cloudsql-instances" = google_sql_database_instance.{{ unquote ( fillEnv $sqlResource) }}.connection_name{{end}}
        }
      }
  }
}
{{ end }}

{{ if .GabiProject.PubSub }}
resource "google_pubsub_topic" "deadletter_pubsub_topic" {
  name = "deadletter_topic"
}
{{ range  .GabiProject.PubSub.Topics }}
resource "google_pubsub_topic" "{{formatAction "pubsub_topic" .Name}}" {
  name = "{{ .Name }}"
}
{{ $topicName := .Name }}
{{ range .Subscriptions }}
resource "google_pubsub_subscription" "{{formatAction "pubsub_subscription" .Name }}" {
  name  = "{{.Name}}"
  topic = "{{$topicName}}"

  ack_deadline_seconds = {{ if eq .AckTimeOut 0}}20{{else}}{{.AckTimeOut}}{{end}}
  retry_policy {
    minimum_backoff = "10s"
  }
  dead_letter_policy {
      dead_letter_topic = google_pubsub_topic.deadletter_pubsub_topic.id
      max_delivery_attempts = 10
    }
  push_config {
    push_endpoint = {{ cloudRunUrl .CloudRun }}
     {{ if .ServiceAccount }}
      oidc_token {
        service_account_email = "{{.ServiceAccount}}@{{ $projectName }}.iam.gserviceaccount.com"
        audience = {{ cloudRunUrl .CloudRun }}
      }
      {{end}}
  }
}
{{ end }}
{{ end }}
{{ end }}
{{range .GabiProject.Storage }}
{{ $name := fillEnv .Name }}
resource "google_storage_bucket" "{{formatAction "storage" $name}}" {
  name          = {{$name}}
  location      = "{{.Location}}"
  storage_class = "{{.StorageClass}}"
  {{ if eq .Permission.AccessControl "UNIFORM" }} uniform_bucket_level_access = true{{ end}}
  {{ if .ObjectVersioning}}
  versioning {
     enabled = true
  }
   lifecycle_rule {
       action {
            type = "Delete"
       }
       condition {
           num_newer_versions = {{ .ObjectVersioning.MaxVersion }}
           with_state = "ARCHIVED"
           }
    }
   lifecycle_rule {
       action {
            type = "Delete"
       }
       condition {
          days_since_noncurrent_time =  {{ .ObjectVersioning.Duration }}
          with_state = "ANY"
       }
    }
  {{end}}
}
{{end}}

{{if .GabiProject.Gcp.Firebase }}
resource "google_firebase_project" "default" {
  provider = google-beta
  project  = "{{ $projectName }}"
}{{ end }}`
}
