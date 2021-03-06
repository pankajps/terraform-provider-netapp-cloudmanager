package cloudmanager

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceOCCMGCP() *schema.Resource {
	return &schema.Resource{
		Create: resourceOCCMGCPCreate,
		Read:   resourceOCCMGCPRead,
		Delete: resourceOCCMGCPDelete,
		Exists: resourceOCCMGCPExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_account_email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_account_path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"machine_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "n1-standard-4",
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
				ForceNew: true,
			},
			"firewall_tags": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"company": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"proxy_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"proxy_user_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"proxy_password": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"associate_public_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
		},
	}
}

func resourceOCCMGCPCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("Creating OCCM: %#v", d)

	client := meta.(*Client)

	occmDetails := createOCCMDetails{}

	occmDetails.GCPCommonSuffixName = "-vm-boot-deployment"
	occmDetails.Name = d.Get("name").(string)
	occmDetails.GCPProject = d.Get("project_id").(string)
	occmDetails.Zone = d.Get("zone").(string)
	occmDetails.Region = string(occmDetails.Zone[0 : len(occmDetails.Zone)-2])
	occmDetails.SubnetID = d.Get("subnet_id").(string)
	occmDetails.MachineType = d.Get("machine_type").(string)
	occmDetails.ServiceAccountEmail = d.Get("service_account_email").(string)
	client.GCPServiceAccountPath = d.Get("service_account_path").(string)
	occmDetails.FirewallTags = d.Get("firewall_tags").(bool)
	occmDetails.AssociatePublicIP = d.Get("associate_public_ip").(bool)
	occmDetails.Company = d.Get("company").(string)

	if o, ok := d.GetOk("proxy_url"); ok {
		occmDetails.ProxyURL = o.(string)
	}

	if o, ok := d.GetOk("proxy_user_name"); ok {
		occmDetails.ProxyUserName = o.(string)
	}

	if o, ok := d.GetOk("proxy_password"); ok {
		occmDetails.ProxyPassword = o.(string)
	}

	if o, ok := d.GetOk("account_id"); ok {
		client.AccountID = o.(string)
	}

	res, err := client.deployGCPVM(occmDetails)
	if err != nil {
		log.Print("Error creating instance")
		return err
	}

	d.SetId(occmDetails.Name)
	if err := d.Set("client_id", res.ClientID); err != nil {
		return fmt.Errorf("Error reading occm client_id: %s", err)
	}

	if err := d.Set("account_id", res.AccountID); err != nil {
		return fmt.Errorf("Error reading occm account_id: %s", err)
	}

	log.Printf("Created occm: %v", res)

	return resourceOCCMGCPRead(d, meta)
}

func resourceOCCMGCPRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("Reading OCCM: %#v", d)
	client := meta.(*Client)

	occmDetails := createOCCMDetails{}

	occmDetails.GCPCommonSuffixName = "-vm-boot-deployment"
	occmDetails.Name = d.Get("name").(string)
	occmDetails.GCPProject = d.Get("project_id").(string)
	occmDetails.Region = d.Get("zone").(string)
	occmDetails.SubnetID = d.Get("subnet_id").(string)
	client.GCPServiceAccountPath = d.Get("service_account_path").(string)
	occmDetails.Company = d.Get("company").(string)

	id := d.Id() + "-vm-boot-deployment"

	resID, err := client.getdeployGCPVM(occmDetails, id)
	if err != nil {
		log.Print("Error getting occm")
		return err
	}

	if resID != id {
		return fmt.Errorf("Expected occm ID %v, Response could not find", id)
	}

	return nil
}

func resourceOCCMGCPDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("Deleting OCCM: %#v", d)

	client := meta.(*Client)

	occmDetails := deleteOCCMDetails{}

	occmDetails.GCPCommonSuffixName = "-vm-boot-deployment"
	id := d.Id() + occmDetails.GCPCommonSuffixName
	occmDetails.InstanceID = id
	occmDetails.Name = d.Get("name").(string)
	occmDetails.Project = d.Get("project_id").(string)
	client.GCPServiceAccountPath = d.Get("service_account_path").(string)
	occmDetails.Region = d.Get("zone").(string)
	client.ClientID = d.Get("client_id").(string)
	client.AccountID = d.Get("account_id").(string)

	deleteErr := client.deleteOCCMGCP(occmDetails)
	if deleteErr != nil {
		return deleteErr
	}

	return nil
}

func resourceOCCMGCPExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	log.Printf("Checking existence of OCCM: %#v", d)
	client := meta.(*Client)

	occmDetails := createOCCMDetails{}

	occmDetails.GCPCommonSuffixName = "-vm-boot-deployment"
	occmDetails.Name = d.Get("name").(string)
	occmDetails.GCPProject = d.Get("project_id").(string)
	occmDetails.Region = d.Get("zone").(string)
	occmDetails.SubnetID = d.Get("subnet_id").(string)
	client.GCPServiceAccountPath = d.Get("service_account_path").(string)
	occmDetails.Company = d.Get("company").(string)

	id := d.Id() + occmDetails.GCPCommonSuffixName

	resID, err := client.getdeployGCPVM(occmDetails, id)
	if err != nil {
		log.Print("Error getting occm")
		return false, err
	}

	if resID != id {
		d.SetId("")
		return false, nil
	}

	return true, nil
}
