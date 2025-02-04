package instaclustr

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceKafkaAcl() *schema.Resource {
	return &schema.Resource{
		Create: resourceKafkaAclCreate,
		Read:   resourceKafkaAclRead,
		Delete: resourceKafkaAclDelete,

		Importer: &schema.ResourceImporter{
			State: resourceKafkaAclStateImport,
		},

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"principal": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"host": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"operation": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"permission_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pattern_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

// mainly for coverage testing
type KafkaAclResourceDataInterface interface {
	Get(key string) interface{}
	Set(key string, value interface{}) error
	SetId(value string)
	Id() string
}

func parseResourceData(d KafkaAclResourceDataInterface) (string, KafkaAcl) {
	cluster_id := d.Get("cluster_id").(string)

	return cluster_id, KafkaAcl {
		Principal:	d.Get("principal").(string),
		Host:		d.Get("host").(string),
		ResourceType:	d.Get("resource_type").(string),
		ResourceName: 	d.Get("resource_name").(string),
		Operation: 	d.Get("operation").(string),
		PermissionType:	d.Get("permission_type").(string),
		PatternType: 	d.Get("pattern_type").(string),
	}
}

func doResourceKafkaAclCreate(d KafkaAclResourceDataInterface, client KafkaAclAPIClientInterface) error {
	cluster_id, acl := parseResourceData(d)

	log.Printf("[INFO] Creating Kafka ACL in %s.", cluster_id)

	// Cluster has to reach running state first
	cluster, err := client.ReadCluster(cluster_id)
	if err != nil {
		return fmt.Errorf("[Error] Error in getting the status of the cluster: %w", err)
	}
	if cluster.ClusterStatus != "RUNNING" {
		return fmt.Errorf("[Error] Cluster %s is not RUNNING. Currently in %s state", cluster_id, cluster.ClusterStatus)
	}

	var jsonStr []byte
	jsonStr, err = json.Marshal(acl)

	if err != nil {
		return fmt.Errorf("[Error] Error creating kafka ACL creation request: %w", err)
	}

	remoteAcls, err := client.ReadKafkaAcls(cluster_id, jsonStr)
	if err != nil {
		return fmt.Errorf("[Error] Error reading kafka ACL: %w", err)
	}
	
	// When we pass the exact parameters, there should be at most 1 ACL in the list (or none if no ACL match).
	// If there is no ACL in the list, we can proceed to the ACL creation.
	// If there is an ACL in the list, then we abort the create ACL resource as we don't want to deal with duplicate resources.
	if len(remoteAcls) > 0 {
		return fmt.Errorf("[Error] Error creating kafka ACL: the resource already exists, use terraform import instead.")
	} 

	err = client.CreateKafkaAcl(cluster_id, jsonStr)
	if err != nil {
		return fmt.Errorf("[Error] Error creating kafka ACL: %w", err)
	}

	d.SetId(fmt.Sprintf("%s&%s&%s&%s&%s&%s&%s&%s", cluster_id, acl.Principal, acl.Host, acl.ResourceType, 
		acl.ResourceName, acl.Operation, acl.PermissionType, acl.PatternType))

	log.Printf("[INFO] Kafka ACL (principal=%s,host=%s,resourceType=%s,resourceName=%s,operation=%s,permissionType=%s,patternType=%s) has been created.", 
		acl.Principal, acl.Host, acl.ResourceType, acl.ResourceName, acl.Operation, acl.PermissionType, acl.PatternType)
	return nil
}

func removeKafkaAclResource(d KafkaAclResourceDataInterface) {
	d.SetId("")
	d.Set("cluster_id", "")
	d.Set("principal", "")
	d.Set("host", "")
	d.Set("resource_type", "")
	d.Set("resource_name", "")
	d.Set("operation", "")
	d.Set("permission_type", "")
	d.Set("pattern_type", "")
}

func doResourceKafkaAclRead(d KafkaAclResourceDataInterface, client KafkaAclAPIClientInterface) error {
	cluster_id, data := parseResourceData(d)

	log.Printf("[INFO] Reading Kafka ACL in %s.", cluster_id)

	// Cluster has to reach running state first
	cluster, err := client.ReadCluster(cluster_id)
	if err != nil {
		return fmt.Errorf("[Error] Error in getting the status of the cluster: %w", err)
	}
	if cluster.ClusterStatus != "RUNNING" {
		return fmt.Errorf("[Error] Cluster %s is not RUNNING. Currently in %s state", cluster_id, cluster.ClusterStatus)
	}

	var jsonStr []byte
	jsonStr, err = json.Marshal(data)
	if err != nil {
		return fmt.Errorf("[Error] Error creating kafka ACL read request: %w", err)
	}

	remoteAcls, err := client.ReadKafkaAcls(cluster_id, jsonStr)
	if err != nil {
		return fmt.Errorf("[Error] Error reading kafka ACL: %w", err)
	}
	
	// When we pass the exact parameters, there should be at most 1 ACL in the list (or none if no ACL match).
	// If there is no ACL in the list, we can assume that someone has deleted this particular entry from Kafka.
	// If there is an ACL in the list, then we can say that this particular resource exists in Kafka and we don't need to change anything.
	if len(remoteAcls) == 0 {
		removeKafkaAclResource(d)
	} 

	return nil
}

func doResourceKafkaAclDelete(d KafkaAclResourceDataInterface, client KafkaAclAPIClientInterface) error {
	cluster_id, acl := parseResourceData(d)
	
	log.Printf("[INFO] Deleting Kafka ACL in %s.", cluster_id)

	var jsonStr []byte
	jsonStr, err := json.Marshal(acl)
	if err != nil {
		return fmt.Errorf("[Error] Error creating kafka ACL delete request: %w", err)
	}

	err = client.DeleteKafkaAcl(cluster_id, jsonStr)
	if err != nil {
		return fmt.Errorf("[Error] Error deleting Kafka ACL: %w", err)
	}

	removeKafkaAclResource(d)

	log.Printf("[INFO] Kafka ACL (principal=%s,host=%s,resourceType=%s,resourceName=%s,operation=%s,permissionType=%s,patternType=%s) has been deleted.", 
		acl.Principal, acl.Host, acl.ResourceType, acl.ResourceName, acl.Operation, acl.PermissionType, acl.PatternType)
	return nil
}

func resourceKafkaAclStateImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "&")
	if len(idParts) != 8 || stringInSlice("", idParts) {
		return nil, fmt.Errorf("[Error] Unexpected format of ID (%q), expected <CLUSTER-ID>&<PRINCIPAL>&<HOST>&<RESOURCE-TYPE>&<RESOURCE-NAME>&<OPERATION>&<PERMISSION-TYPE>&<PATTERN-TYPE>", d.Id())
	}
	d.SetId(d.Id())
	d.Set("cluster_id", idParts[0])
	d.Set("principal", idParts[1])
	d.Set("host", idParts[2])
	d.Set("resource_type", idParts[3])
	d.Set("resource_name", idParts[4])
	d.Set("operation", idParts[5])
	d.Set("permission_type", idParts[6])
	d.Set("pattern_type", idParts[7])
	return []*schema.ResourceData{d}, nil
}

// Ugly hacks for unit testing
func resourceKafkaAclCreate(d *schema.ResourceData, meta interface{}) error {
	return doResourceKafkaAclCreate(d, meta.(*Config).Client)
} 

func resourceKafkaAclRead(d *schema.ResourceData, meta interface{}) error {
	return doResourceKafkaAclRead(d, meta.(*Config).Client)
}

func resourceKafkaAclDelete(d *schema.ResourceData, meta interface{}) error {
	return doResourceKafkaAclDelete(d, meta.(*Config).Client)
}
