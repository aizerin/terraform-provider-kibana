package kb

import (
	"encoding/json"
	"fmt"
	"reflect"

	kibana "github.com/disaster37/go-kibana-rest/v7"
	kbapi "github.com/disaster37/go-kibana-rest/v7/kbapi"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	log "github.com/sirupsen/logrus"
)

const (
	indexPattern = "index-pattern"
)

func diffSuppressIndexPattern(k, old, new string, d *schema.ResourceData) bool {
	var oo, no interface{}
	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		return false
	}

	if om, ok := oo.(map[string]interface{}); ok {
		normalizeIndexPattern(om)
	}

	if nm, ok := no.(map[string]interface{}); ok {
		normalizeIndexPattern(nm)
	}

	return reflect.DeepEqual(oo, no)
}

func normalizeIndexPattern(pol map[string]interface{}) {
	delete(pol, "coreMigrationVersion")
	delete(pol, "migrationVersion")
	delete(pol, "namespaces")
	delete(pol, "references")
	delete(pol, "type")
	delete(pol, "updated_at")
	delete(pol, "id")
	delete(pol, "version")
}

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
				ForceNew: true,
			},
			"space_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"body": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: diffSuppressIndexPattern,
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
	body := d.Get("body").(string)

	jsonMap := make(map[string]interface{})

	err = json.Unmarshal([]byte(body), &jsonMap)
	if err != nil {
		panic(err)
	}

	response, err := client.API.KibanaSavedObject.Create(jsonMap, indexPattern, name, true, spaceId)
	if err != nil {
		return err
	}
	log.Debug(response)
	d.SetId(name + spaceId)

	log.Infof("Created user space %s successfully", name)

	return resourceKibanaIndexPatternRead(d, meta)
}

func resourceKibanaIndexPatternRead(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	name := d.Get("name").(string)
	spaceId := d.Get("space_id").(string)

	log.Debugf("Index pattern id:  %s", id)

	client := meta.(*kibana.Client)

	data, err := client.API.KibanaSavedObject.Get(indexPattern, name, spaceId)
	if err != nil {
		return err
	}

	if data == nil {
		fmt.Printf("[WARN] Index pattern %s in space %s not found - removing from state", name, spaceId)
		log.Warnf("Index pattern %s in space %s not found - removing from state", name, spaceId)
		d.SetId("")
		return nil
	}

	log.Debugf("Get index pattern %s successfully:\n%s", id, data)

	tj, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("fail to unmarshal: %v", err)
	}

	d.Set("body", string(tj))

	log.Infof("Read patter index %s successfully", id)

	return nil
}

func resourceKibanaIndexPatternUpdate(d *schema.ResourceData, meta interface{}) error {
	var (
		err error
	)
	client := meta.(*kibana.Client)

	name := d.Get("name").(string)
	spaceId := d.Get("space_id").(string)
	body := d.Get("body").(string)

	jsonMap := make(map[string]interface{})

	err = json.Unmarshal([]byte(body), &jsonMap)
	if err != nil {
		panic(err)
	}

	response, err := client.API.KibanaSavedObject.Update(jsonMap, indexPattern, name, spaceId)
	log.Debug(response)
	if err != nil {
		return err
	}

	log.Infof("Updated index pattern %s successfully", name)

	return resourceKibanaIndexPatternRead(d, meta)
}

func resourceKibanaIndexPatternDelete(d *schema.ResourceData, meta interface{}) error {

	name := d.Get("name").(string)
	id := d.Id()
	log.Debugf("Index pattern id: %s", id)
	spaceId := d.Get("space_id").(string)

	client := meta.(*kibana.Client)

	err := client.API.KibanaSavedObject.Delete(indexPattern, name, spaceId)

	if err != nil {
		if err.(kbapi.APIError).Code == 404 {
			fmt.Printf("[WARN] User space %s not found - removing from state", id)
			log.Warnf("User space %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return err

	}

	d.SetId("")

	log.Infof("Deleted user space %s successfully", id)
	return nil

}
