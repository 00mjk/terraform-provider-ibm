// Copyright IBM Corp. 2023 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package secretsmanager

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
	"github.com/IBM/secrets-manager-go-sdk/secretsmanagerv2"
)

func ResourceIbmSmEnRegistration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIbmSmEnRegistrationCreate,
		ReadContext:   resourceIbmSmEnRegistrationRead,
		UpdateContext: resourceIbmSmEnRegistrationUpdate,
		DeleteContext: resourceIbmSmEnRegistrationDelete,
		Importer:      &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"event_notifications_instance_crn": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.InvokeValidator("ibm_sm_en_registration", "event_notifications_instance_crn"),
				Description:  "A CRN that uniquely identifies an IBM Cloud resource.",
			},
			"event_notifications_source_name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.InvokeValidator("ibm_sm_en_registration", "event_notifications_source_name"),
				Description:  "The name that is displayed as a source that is in your Event Notifications instance.",
			},
			"event_notifications_source_description": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validate.InvokeValidator("ibm_sm_en_registration", "event_notifications_source_description"),
				Description:  "An optional description for the source  that is in your Event Notifications instance.",
			},
		},
	}
}

func ResourceIbmSmEnRegistrationValidator() *validate.ResourceValidator {
	validateSchema := make([]validate.ValidateSchema, 0)
	validateSchema = append(validateSchema,
		validate.ValidateSchema{
			Identifier:                 "event_notifications_instance_crn",
			ValidateFunctionIdentifier: validate.ValidateRegexpLen,
			Type:                       validate.TypeString,
			Required:                   true,
			Regexp:                     `^crn:v[0-9](:([A-Za-z0-9-._~!$&'()*+,;=@\/]|%[0-9A-Z]{2})*){8}$`,
			MinValueLength:             9,
			MaxValueLength:             512,
		},
		validate.ValidateSchema{
			Identifier:                 "event_notifications_source_name",
			ValidateFunctionIdentifier: validate.ValidateRegexpLen,
			Type:                       validate.TypeString,
			Required:                   true,
			Regexp:                     `(.*?)`,
			MinValueLength:             2,
			MaxValueLength:             256,
		},
		validate.ValidateSchema{
			Identifier:                 "event_notifications_source_description",
			ValidateFunctionIdentifier: validate.ValidateRegexpLen,
			Type:                       validate.TypeString,
			Optional:                   true,
			Regexp:                     `(.*?)`,
			MinValueLength:             0,
			MaxValueLength:             1024,
		},
	)

	resourceValidator := validate.ResourceValidator{ResourceName: "ibm_sm_en_registration", Schema: validateSchema}
	return &resourceValidator
}

func resourceIbmSmEnRegistrationCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	secretsManagerClient, err := meta.(conns.ClientSession).SecretsManagerV2()
	if err != nil {
		return diag.FromErr(err)
	}

	region := getRegion(secretsManagerClient, d)
	instanceId := d.Get("instance_id").(string)
	secretsManagerClient = getClientWithInstanceEndpoint(secretsManagerClient, instanceId, region, getEndpointType(secretsManagerClient, d))

	createNotificationsRegistrationOptions := &secretsmanagerv2.CreateNotificationsRegistrationOptions{}

	createNotificationsRegistrationOptions.SetEventNotificationsInstanceCrn(d.Get("event_notifications_instance_crn").(string))
	createNotificationsRegistrationOptions.SetEventNotificationsSourceName(d.Get("event_notifications_source_name").(string))
	if _, ok := d.GetOk("event_notifications_source_description"); ok {
		createNotificationsRegistrationOptions.SetEventNotificationsSourceDescription(d.Get("event_notifications_source_description").(string))
	}

	_, response, err := secretsManagerClient.CreateNotificationsRegistrationWithContext(context, createNotificationsRegistrationOptions)
	if err != nil {
		log.Printf("[DEBUG] CreateNotificationsRegistrationWithContext failed %s\n%s", err, response)
		return diag.FromErr(fmt.Errorf("CreateNotificationsRegistrationWithContext failed %s\n%s", err, response))
	}

	d.SetId(fmt.Sprintf("%s/%s", region, instanceId))

	return resourceIbmSmEnRegistrationRead(context, d, meta)
}

func resourceIbmSmEnRegistrationRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	secretsManagerClient, err := meta.(conns.ClientSession).SecretsManagerV2()
	if err != nil {
		return diag.FromErr(err)
	}

	id := strings.Split(d.Id(), "/")
	region := id[0]
	instanceId := id[1]
	secretsManagerClient = getClientWithInstanceEndpoint(secretsManagerClient, instanceId, region, getEndpointType(secretsManagerClient, d))

	getNotificationsRegistrationOptions := &secretsmanagerv2.GetNotificationsRegistrationOptions{}

	notificationsRegistration, response, err := secretsManagerClient.GetNotificationsRegistrationWithContext(context, getNotificationsRegistrationOptions)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		log.Printf("[DEBUG] GetNotificationsRegistrationWithContext failed %s\n%s", err, response)
		return diag.FromErr(fmt.Errorf("GetNotificationsRegistrationWithContext failed %s\n%s", err, response))
	}

	if err = d.Set("instance_id", instanceId); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting instance_id: %s", err))
	}
	if err = d.Set("region", region); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting region: %s", err))
	}
	if err = d.Set("event_notifications_instance_crn", notificationsRegistration.EventNotificationsInstanceCrn); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting event_notifications_instance_crn: %s", err))
	}

	return nil
}

func resourceIbmSmEnRegistrationUpdate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	secretsManagerClient, err := meta.(conns.ClientSession).SecretsManagerV2()
	if err != nil {
		return diag.FromErr(err)
	}

	id := strings.Split(d.Id(), "/")
	region := id[0]
	instanceId := id[1]
	secretsManagerClient = getClientWithInstanceEndpoint(secretsManagerClient, instanceId, region, getEndpointType(secretsManagerClient, d))

	createNotificationsRegistrationOptions := &secretsmanagerv2.CreateNotificationsRegistrationOptions{}

	hasChange := false

	if d.HasChange("event_notifications_instance_crn") || d.HasChange("event_notifications_source_name") {
		createNotificationsRegistrationOptions.SetEventNotificationsInstanceCrn(d.Get("event_notifications_instance_crn").(string))
		createNotificationsRegistrationOptions.SetEventNotificationsSourceName(d.Get("event_notifications_source_name").(string))
		hasChange = true
	}
	if d.HasChange("event_notifications_source_description") {
		createNotificationsRegistrationOptions.SetEventNotificationsSourceDescription(d.Get("event_notifications_source_description").(string))
		hasChange = true
	}

	if hasChange {
		_, response, err := secretsManagerClient.CreateNotificationsRegistrationWithContext(context, createNotificationsRegistrationOptions)
		if err != nil {
			log.Printf("[DEBUG] CreateNotificationsRegistrationWithContext failed %s\n%s", err, response)
			return diag.FromErr(fmt.Errorf("CreateNotificationsRegistrationWithContext failed %s\n%s", err, response))
		}
	}

	return resourceIbmSmEnRegistrationRead(context, d, meta)
}

func resourceIbmSmEnRegistrationDelete(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	secretsManagerClient, err := meta.(conns.ClientSession).SecretsManagerV2()
	if err != nil {
		return diag.FromErr(err)
	}

	id := strings.Split(d.Id(), "/")
	region := id[0]
	instanceId := id[1]
	secretsManagerClient = getClientWithInstanceEndpoint(secretsManagerClient, instanceId, region, getEndpointType(secretsManagerClient, d))

	deleteNotificationsRegistrationOptions := &secretsmanagerv2.DeleteNotificationsRegistrationOptions{}

	response, err := secretsManagerClient.DeleteNotificationsRegistrationWithContext(context, deleteNotificationsRegistrationOptions)
	if err != nil {
		log.Printf("[DEBUG] DeleteNotificationsRegistrationWithContext failed %s\n%s", err, response)
		return diag.FromErr(fmt.Errorf("DeleteNotificationsRegistrationWithContext failed %s\n%s", err, response))
	}

	d.SetId("")

	return nil
}
