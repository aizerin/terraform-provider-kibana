package kb

import (
	"encoding/json"
	"fmt"

	kibana "github.com/disaster37/go-kibana-rest/v7"
	kbapi "github.com/disaster37/go-kibana-rest/v7/kbapi"
	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	log "github.com/sirupsen/logrus"
)

func resourceKibanaIndexPattern() *schema.Resource {
	return &schema.Resource{
		Create: resourceKibanaIndexPatternCreate,
		Read:   resourceKibanaIndexPatternRead,
		Update: resourceKibanaIndexPatternUpdate,
		Delete: resourceKibanaIndexPatternDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"space_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceKibanaIndexPatternCreate(d *schema.ResourceData, meta interface{}) error {
	var (
		err error
	)
	client := meta.(*kibana.Client)

	name := d.Get("name").(string)
	spaceId := d.Get("space_id").(string)

	response, err := kibanaIndexPatternCreate(client.Client, spaceId, name)
	if err != nil {
		return err
	}
	log.Debug(response)
	d.SetId(response.IndexPattern.Id)

	log.Infof("Created index pattern %s successfully", name)

	return resourceKibanaIndexPatternRead(d, meta)
}

func resourceKibanaIndexPatternRead(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	name := d.Get("name").(string)
	spaceId := d.Get("space_id").(string)

	log.Debugf("Index pattern id:  %s", id)

	client := meta.(*kibana.Client)

	data, err := kibanaIndexPatternGet(client.Client, spaceId, id)
	if err != nil {
		if err.(kbapi.APIError).Code == 404 {
			fmt.Printf("[WARN] Index pattern %s in space %s not found - removing from state", name, spaceId)
			log.Warnf("Index pattern %s in space %s not found - removing from state", name, spaceId)
			d.SetId("")
			return nil
		}
		return err
	}

	log.Debugf("Get index pattern %s successfully:\n%s", id, data)

	d.Set("name", data.IndexPattern.Title)

	log.Infof("Read index pattern %s successfully", id)

	return nil
}

func resourceKibanaIndexPatternUpdate(d *schema.ResourceData, meta interface{}) error {
	var (
		err error
	)
	client := meta.(*kibana.Client)

	id := d.Id()
	name := d.Get("name").(string)
	spaceId := d.Get("space_id").(string)

	response, err := kibanaIndexPatternUpdate(client.Client, spaceId, name, id)
	if err != nil {
		return err
	}
	log.Debug(response)

	log.Infof("Updated index pattern %s successfully", name)

	return resourceKibanaIndexPatternRead(d, meta)
}

func resourceKibanaIndexPatternDelete(d *schema.ResourceData, meta interface{}) error {

	name := d.Get("name").(string)
	id := d.Id()
	log.Debugf("Index pattern id: %s", id)
	spaceId := d.Get("space_id").(string)

	client := meta.(*kibana.Client)

	err := kibanaIndexPatternDelete(client.Client, spaceId, name)

	if err != nil {
		if err.(kbapi.APIError).Code == 404 {
			fmt.Printf("[WARN] Index pattern %s in space %s not found - removing from state", name, spaceId)
			log.Warnf("Index pattern %s in space %s not found - removing from state", name, spaceId)
			d.SetId("")
			return nil
		}
		return err

	}

	d.SetId("")

	log.Infof("Deleted index pattern %s successfully", id)
	return nil

}

// todo move to api someday

const (
	indexPatternPath       = "/s/%s/api/index_patterns/index_pattern/%s"
	indexPatternCreatePath = "/s/%s/api/index_patterns/index_pattern"
)

type IndexPatternBody struct {
	IndexPattern IndexPatternData `json:"index_pattern"`
}

type IndexPatternData struct {
	Id            string `json:"id"`
	Title         string `json:"title"`
	TimeFieldName string `json:"timeFieldName"`
}

func kibanaIndexPatternCreate(c *resty.Client, kibanaSpace string, indexPatterName string) (*IndexPatternBody, error) {

	path := fmt.Sprintf(indexPatternCreatePath, kibanaSpace)
	log.Debugf("URL to create index pattern: %s", path)

	resp, err := c.R().SetBody(IndexPatternBody{
		IndexPattern: IndexPatternData{
			Id:            indexPatterName,
			Title:         indexPatterName,
			TimeFieldName: "@timestamp",
		},
	}).Post(path)

	if err != nil {
		return nil, err
	}
	log.Debug("Index Pattern Create Response: ", resp)
	if resp.StatusCode() >= 300 {
		return nil, kbapi.NewAPIError(resp.StatusCode(), resp.Status())
	}
	var dataResponse IndexPatternBody
	err = json.Unmarshal(resp.Body(), &dataResponse)
	if err != nil {
		return nil, err
	}
	log.Debug("Data response: ", dataResponse)

	return &dataResponse, nil
}

func kibanaIndexPatternUpdate(c *resty.Client, kibanaSpace string, indexPatterName string, id string) (*IndexPatternBody, error) {

	path := fmt.Sprintf(indexPatternPath, kibanaSpace, id)
	log.Debugf("URL to update index pattern: %s", path)

	resp, err := c.R().SetBody(IndexPatternBody{
		IndexPattern: IndexPatternData{
			Title: indexPatterName,
		},
	}).Post(path)

	if err != nil {
		return nil, err
	}
	log.Debug("Intex Pattern Update Response: ", resp)
	if resp.StatusCode() >= 300 {
		return nil, kbapi.NewAPIError(resp.StatusCode(), resp.Status())
	}
	var dataResponse IndexPatternBody
	err = json.Unmarshal(resp.Body(), &dataResponse)
	if err != nil {
		return nil, err
	}
	log.Debug("Data response: ", dataResponse)

	return &dataResponse, nil
}

func kibanaIndexPatternGet(c *resty.Client, kibanaSpace string, id string) (*IndexPatternBody, error) {

	path := fmt.Sprintf(indexPatternPath, kibanaSpace, id)
	log.Debugf("URL to get pattern: %s", path)

	resp, err := c.R().Get(path)

	if err != nil {
		return nil, err
	}
	log.Debug("Index Pattern Get Response: ", resp)
	if resp.StatusCode() >= 300 {
		return nil, kbapi.NewAPIError(resp.StatusCode(), resp.Status())
	}

	var dataResponse IndexPatternBody
	err = json.Unmarshal(resp.Body(), &dataResponse)
	if err != nil {
		return nil, err
	}
	log.Debug("Data response: ", dataResponse)

	return &dataResponse, nil
}

func kibanaIndexPatternDelete(c *resty.Client, kibanaSpace string, id string) error {

	path := fmt.Sprintf(indexPatternPath, kibanaSpace, id)
	log.Debugf("URL to delete pattern: %s", path)

	resp, err := c.R().Delete(path)

	if err != nil {
		return err
	}
	log.Debug("Index Pattern Delete Response: ", resp)
	if resp.StatusCode() >= 300 {
		return kbapi.NewAPIError(resp.StatusCode(), resp.Status())
	}

	return nil
}
