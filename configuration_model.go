package main

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"text/template"
	"time"
)

type ProjectConfiguration struct {
	GabiProject       GabiProject `yaml:"gabi_project"`
	AsMap             map[string]interface{}
	ServiceAccountMap map[string][]string
	HasBackup         bool
}
type Secret struct {
	Location string            `yaml:"location"`
	Values   map[string]string `yaml:"values"`
}
type GabiProject struct {
	Gcp             Project          `yaml:"gcp"`
	CustomRole      []CustomRole     `yaml:"custom_role"`
	Apis            []string         `yaml:"api"`
	ServiceAccounts []ServiceAccount `yaml:"service_accounts"`
	CloudRuns       []CloudRun       `yaml:"cloud_runs"`
	PubSub          PubSub           `yaml:"pubsub"`
	Storage         []Storage        `yaml:"storage"`
	Sql             Sql              `yaml:"sql"`
	Secret          Secret           `yaml:"secret"`
}
type CustomRole struct {
	Name  string   `yaml:"name"`
	Roles []string `yaml:"roles"`
}

type Project struct {
	Name           string `yaml:"name"`
	ServiceKeyFile string `yaml:"service_key_file"`
	Location       string `yaml:"location"`
	Firebase       bool   `yaml:"firebase"`
}

type ServiceAccount struct {
	Name  string   `yaml:"name"`
	Roles []string `yaml:"roles"`
}

type CloudRun struct {
	Name                 string            `yaml:"name"`
	ServiceAccount       string            `yaml:"service_account"`
	AllowUnauthenticated bool              `yaml:"allow_unauthenticated"`
	Location             string            `yaml:"location"`
	Env                  map[string]string `yaml:"env"`
	Secret               map[string]string `yaml:"secret"`
	Image                string            `yaml:"image"`
	Memory               string            `yaml:"memory"`
	Cpu                  int               `yaml:"cpu"`
	MaxInstances         int               `yaml:"max_instances"`
	SqlInstance          string            `yaml:"sql_instance"`
}
type PubSub struct {
	Topics []Topic `yaml:"topics"`
}
type Topic struct {
	Name          string         `yaml:"name"`
	Subscriptions []Subscription `yaml:"subscriptions"`
}
type Subscription struct {
	Name           string `yaml:"name"`
	CloudRun       string `yaml:"cloud_run"`
	AckTimeOut     int    `yaml:"ack_time_out"`
	ServiceAccount string `yaml:"service_account"`
}

type Storage struct {
	Name             string            `yaml:"name"`
	Location         string            `yaml:"location"`
	StorageClass     string            `yaml:"storage_class"`
	Firebase         bool              `yaml:"firebase"`
	Permission       Permission        `yaml:"permission"`
	NotPublic        bool              `yaml:"not_public"`
	Backup           []StorageBackup   `yaml:"backup"`
	ObjectVersioning *ObjectVersioning `yaml:"object_versioning"`
}
type StorageBackup struct {
	DstBucket string `yaml:"dst_bucket"`
	DstPath   string `yaml:"dst_path"`
	StartDate string `yaml:"start_date"`
	RunEvery  string `yaml:"run_every"`
}
type Permission struct {
	AccessControl string `yaml:"access_control"`
}
type ObjectVersioning struct {
	MaxVersion int    `yaml:"max_version"`
	Duration   string `yaml:"duration"`
}

type Sql struct {
	Name          string            `yaml:"name"`
	Version       string            `yaml:"version"`
	Location      string            `yaml:"location"`
	Tier          string            `yaml:"tier"`
	User          string            `yaml:"user"`
	Password      string            `yaml:"password"`
	Database      string            `yaml:"database"`
	Configuration *SqlConfiguration `yaml:"configuration"`
}
type SqlConfiguration struct {
	SSL                 bool               `yaml:"ssl"`
	QueryInsight        bool               `yaml:"query_insight"`
	AutoStorageIncrease bool               `yaml:"auto_storage_increase"`
	HighAvailability    bool               `yaml:"high_availability"`
	MaintenanceWindow   *MaintenanceWindow `yaml:"maintenance_window"`
	BackupSqlOption     *BackupSqlOption   `yaml:"backup_sql_option"`
}
type BackupSqlOption struct {
	region   string `yaml:"region"`
	duration string `yaml:"duration"`
}
type MaintenanceWindow struct {
	Day  int `yaml:"day"`
	Hour int `yaml:"hour"`
}

func (s *ProjectConfiguration) ToTerraform() string {
	return s.createTerraformScript()
}

func (s *ProjectConfiguration) getReplacers() *strings.Replacer {
	keys := parse("", s.AsMap)

	replacerKeys := make([]string, len(keys)*2)
	i := 0
	for k, v := range keys {
		replacerKeys[i], replacerKeys[i+1] = fmt.Sprintf("{%s}", k), v
		i += 2
	}
	newReplacer := strings.NewReplacer(replacerKeys...)
	return newReplacer
}

func getResourceName(resourceType string, resource string) string {
	return format(resourceType, resource)
}
func databaseUrl(cloudRunName string) string {
	return fmt.Sprintf("google_sql_database_instance.%s.public_ip_address", getResourceName("sql", cloudRunName))
}
func cloudRunUrl(cloudRunName string) string {
	return fmt.Sprintf("google_cloud_run_service.%s.status[0].url", getResourceName("cloudrun", cloudRunName))
}
func (s *ProjectConfiguration) createTerraformScript() string {
	replacer := s.getReplacers()

	t, err := template.New("main").Funcs(template.FuncMap{
		"unquote":         unquote,
		"cloudRunUrl":     cloudRunUrl,
		"getResourceName": format,
		"formatAction":    format,
		"join": func(elems []string) string {
			return "\"" + strings.ReplaceAll(strings.Join(elems, ","), ",", "\",\"") + "\""
		},
		"date": func(fullDate string, field string) string {
			t, err := time.Parse("2006-01-02", fullDate)
			if err != nil {
				log.Fatal(err)
			}
			yyyy, mm, dd := t.Date()
			switch field {
			case "yyyy":
				return fmt.Sprintf("%d", yyyy)
			case "mm":
				return fmt.Sprintf("%d", mm)
			case "dd":
				return fmt.Sprintf("%d", dd)
			}
			return "NOT_FOUND"
		},
		"fillEnv": fillEnv(replacer),
		"trim": func(variable string) string {
			return strings.ReplaceAll(replacer.Replace(variable[1:len(variable)-1]), "-", "_")
		},
	}).Parse(getTemplate())
	if err != nil {
		log.Fatal(err)
	}
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, s); err != nil {
		log.Fatal(err)
	}
	return tpl.String()
}

func fillEnv(replacer *strings.Replacer) func(variable string) string {
	return func(variable string) string {
		variable = replacer.Replace(variable)
		if strings.HasPrefix(variable, "{{") {
			function := variable[2:strings.Index(variable, " ")]
			if function == "cloudRunUrl" {
				return cloudRunUrl(strings.TrimSpace(variable[14 : len(variable)-2]))
			} else if function == "databaseUrl" {
				return databaseUrl(replacer.Replace(strings.TrimSpace(variable[14 : len(variable)-2])))
			}
		}

		return fmt.Sprintf("\"%s\"", variable)
	}
}

func unquote(str string) string {
	return str[1 : len(str)-1]
}
func format(role string, name string) string {
	if strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"") {
		name = name[1 : len(name)-1]
	}
	str := ""
	if role == "" {
		str = name
	} else {
		str = fmt.Sprintf("%s_%s", name, role)
	}
	reg, err := regexp.Compile("[^A-Za-z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	return reg.ReplaceAllString(str, "_")
}

func parse(str string, m map[string]interface{}) map[string]string {
	retour := make(map[string]string, 100)
	for k, v := range m {
		k := k
		v := v
		newKey := newKey(str, k)
		rt := reflect.TypeOf(v)
		switch rt.Kind() {
		case reflect.Slice, reflect.Array:
			retour[newKey] = fmt.Sprintf("%v", v)
		case reflect.String:
			retour[newKey] = v.(string)
		case reflect.Bool:
			retour[newKey] = fmt.Sprintf("%v", v)
		case reflect.Int:
			retour[newKey] = fmt.Sprintf("%d", v)
		default:
			for k, v := range parse(newKey, m[k].(map[string]interface{})) {
				k := k
				v := v
				retour[k] = v
			}
		}
	}
	return retour
}

func newKey(str string, k string) string {
	if str == "" {
		return k
	}
	return fmt.Sprintf("%s.%s", str, k)
}
