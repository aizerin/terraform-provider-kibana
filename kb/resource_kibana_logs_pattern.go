package kb

import (
	"encoding/json"
	"fmt"
	"strings"

	kibana "github.com/disaster37/go-kibana-rest/v7"
	"github.com/disaster37/go-kibana-rest/v7/kbapi"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	log "github.com/sirupsen/logrus"
)

func resourceKibanaLogsPattern() *schema.Resource {
	return &schema.Resource{
		Create: resourceKibanaLogsPatternCreate,
		Read:   resourceKibanaLogsPatternRead,
		Update: resourceKibanaLogsPatternUpdate,
		Delete: resourceKibanaLogsPatternDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"space_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"patterns": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceKibanaLogsPatternCreate(d *schema.ResourceData, meta interface{}) error {
	d.SetId(resource.UniqueId())
	return resourceKibanaLogsPatternUpdate(d, meta)
}

func resourceKibanaLogsPatternUpdate(d *schema.ResourceData, meta interface{}) error {
	var (
		err error
	)
	client := meta.(*kibana.Client)

	patternsRaw := d.Get("patterns").([]interface{})
	patterns := make([]string, len(patternsRaw))
	for i, raw := range patternsRaw {
		patterns[i] = raw.(string)
	}

	spaceId := d.Get("space_id").(string)

	response, err := kibanaLogsPatterPatch(client.Client, spaceId, strings.Join(patterns, ","))
	if err != nil {
		return err
	}
	log.Debug(response)

	return resourceKibanaLogsPatternRead(d, meta)
}

func resourceKibanaLogsPatternRead(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	spaceId := d.Get("space_id").(string)

	log.Debugf("Logs index pattern id:  %s", id)

	client := meta.(*kibana.Client)

	data, err := kibanaLogsPatterGet(client.Client, spaceId)
	if err != nil {
		return err
	}

	log.Debugf("Get Logs index pattern %s successfully:\n%s", id, data)

	splittedPatterns := strings.Split(data.Data.Configuration.LogAlias, ",")
	newPatterns := make([]interface{}, len(splittedPatterns))
	for i, raw := range splittedPatterns {
		newPatterns[i] = raw
	}

	d.Set("patterns", newPatterns)

	log.Infof("Read Logs index %s successfully", id)

	return nil
}

func resourceKibanaLogsPatternDelete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")

	log.Infof("Object is deleted when space is deleted automatically")
	fmt.Printf("[INFO] Object is deleted when space is deleted automatically")
	return nil
}

// todo move to api someday

const (
	logsSourcePath = "/s/%s/api/infra/log_source_configurations/default" // Base URL to access on Kibana save objects
)

type LogsBody struct {
	Data LogsData `json:"data"`
}

type LogsFields struct {
}

type LogsData struct {
	LogAlias string     `json:"logAlias"`
	Fields   LogsFields `json:"fields"`
}

type LogsGet struct {
	Data LogsGetData `json:"data"`
}

type LogsGetData struct {
	Configuration LogsGetConfiguration `json:"configuration"`
}

type LogsGetConfiguration struct {
	LogAlias string `json:"logAlias"`
}

func kibanaLogsPatterPatch(c *resty.Client, kibanaSpace string, patterns string) (map[string]interface{}, error) {

	path := fmt.Sprintf(logsSourcePath, kibanaSpace)
	log.Debugf("URL to update object: %s", path)

	resp, err := c.R().SetBody(LogsBody{
		Data: LogsData{
			LogAlias: patterns,
			Fields:   LogsFields{},
		},
	}).Patch(path)

	if err != nil {
		return nil, err
	}
	log.Debug("Response: ", resp)
	if resp.StatusCode() >= 300 {
		return nil, kbapi.NewAPIError(resp.StatusCode(), resp.Status())
	}
	var dataResponse map[string]interface{}
	err = json.Unmarshal(resp.Body(), &dataResponse)
	if err != nil {
		return nil, err
	}
	log.Debug("Data response: ", dataResponse)

	return dataResponse, nil
}

func kibanaLogsPatterGet(c *resty.Client, kibanaSpace string) (*LogsGet, error) {

	path := fmt.Sprintf(logsSourcePath, kibanaSpace)
	log.Debugf("URL to get object: %s", path)

	resp, err := c.R().Get(path)

	if err != nil {
		return nil, err
	}
	log.Debug("Response: ", resp)
	if resp.StatusCode() >= 300 {
		return nil, kbapi.NewAPIError(resp.StatusCode(), resp.Status())
	}

	var dataResponse LogsGet
	err = json.Unmarshal(resp.Body(), &dataResponse)
	if err != nil {
		return nil, err
	}
	log.Debug("Data response: ", dataResponse)

	return &dataResponse, nil
}
