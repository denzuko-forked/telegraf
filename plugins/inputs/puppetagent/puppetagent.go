package puppetagent

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// PuppetAgent is a PuppetAgent plugin
type PuppetAgent struct {
	Location string
}

var sampleConfig = `
  ## Location of puppet last run summary file
  location = "/var/lib/puppet/state/last_run_summary.yaml"
`

type State struct {
	Events    event
	Resources resource
	Changes   change
	Time      time
	Version   version
}

type event struct {
	Failure int64 `yaml:"failure"`
	Noop    int64 `yaml:"noop"`
	Total   int64 `yaml:"total"`
	Success int64 `yaml:"success"`
}

type resource struct {
	Changed          int64 `yaml:"changed"`
	CorrectiveChange int64 `yaml:"corrective_change"`
	Failed           int64 `yaml:"failed"`
	FailedToRestart  int64 `yaml:"failed_to_restart"`
	OutOfSync        int64 `yaml:"out_of_sync"`
	Restarted        int64 `yaml:"restarted"`
	Scheduled        int64 `yaml:"scheduled"`
	Skipped          int64 `yaml:"skipped"`
	Total            int64 `yaml:"total"`
}

type change struct {
	Total int64 `yaml:"total"`
}

type time struct {
	Anchor                float64 `yaml:"anchor"`
	CataLogApplication    float64 `yaml:"catalog_application"`
	ConfigRetrieval       float64 `yaml:"config_retrieval"`
	ConvertCatalog        float64 `yaml:"convert_catalog"`
	Cron                  float64 `yaml:"cron"`
	Exec                  float64 `yaml:"exec"`
	FactGeneration        float64 `yaml:"fact_generation"`
	File                  float64 `yaml:"file"`
	FileBucket            float64 `yaml:"filebucket"`
	Group                 float64 `yaml:"group"`
	LastRun               int64   `yaml:"last_run"`
	NodeRetrieval         float64 `yaml:"node_retrieval"`
	Notify                float64 `yaml:"notify"`
	Package               float64 `yaml:"package"`
	PluginSync            float64 `yaml:"plugin_sync"`
	Schedule              float64 `yaml:"schedule"`
	Service               float64 `yaml:"service"`
	SSHAuthorizedKey      float64 `yaml:"ssh_authorized_key"`
	Total                 float64 `yaml:"total"`
	TransactionEvaluation float64 `yaml:"transaction_evaluation"`
	User                  float64 `yaml:"user"`
}

type version struct {
	ConfigString string `yaml:"config"`
	Puppet       string `yaml:"puppet"`
}

// SampleConfig returns sample configuration message
func (pa *PuppetAgent) SampleConfig() string {
	return sampleConfig
}

// Description returns description of PuppetAgent plugin
func (pa *PuppetAgent) Description() string {
	return `Reads last_run_summary.yaml file and converts to measurements`
}

// Gather reads stats from all configured servers accumulates stats
func (pa *PuppetAgent) Gather(acc telegraf.Accumulator) error {
	if len(pa.Location) == 0 {
		pa.Location = "/var/lib/puppet/state/last_run_summary.yaml"
	}

	if _, err := os.Stat(pa.Location); err != nil {
		return fmt.Errorf("%s", err)
	}

	fh, err := ioutil.ReadFile(pa.Location)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	var puppetState State

	err = yaml.Unmarshal(fh, &puppetState)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	tags := map[string]string{"location": pa.Location}
	structPrinter(&puppetState, acc, tags)

	return nil
}

func structPrinter(s *State, acc telegraf.Accumulator, tags map[string]string) {
	e := reflect.ValueOf(s).Elem()

	fields := make(map[string]interface{})
	for tLevelFNum := 0; tLevelFNum < e.NumField(); tLevelFNum++ {
		name := e.Type().Field(tLevelFNum).Name
		nameNumField := e.FieldByName(name).NumField()

		for sLevelFNum := 0; sLevelFNum < nameNumField; sLevelFNum++ {
			sName := e.FieldByName(name).Type().Field(sLevelFNum).Name
			sValue := e.FieldByName(name).Field(sLevelFNum).Interface()

			lname := strings.ToLower(name)
			lsName := strings.ToLower(sName)
			fields[fmt.Sprintf("%s_%s", lname, lsName)] = sValue
		}
	}
	acc.AddFields("puppetagent", fields, tags)
}

func init() {
	inputs.Add("puppetagent", func() telegraf.Input {
		return &PuppetAgent{}
	})
}
