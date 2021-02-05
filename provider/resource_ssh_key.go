package provider

import (
	"fmt"
	"log"
	"github.com/IBM/vpc-go-sdk/vpcv1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	isKeyName          = "name"
	isKeyPublicKey     = "public_key"
	isKeyType          = "type"
	isKeyFingerprint   = "fingerprint"
	isKeyLength        = "length"
	isKeyTags          = "tags"
	isKeyResourceGroup = "resource_group"
)

const (
	prodBaseController  = "https://cloud.ibm.com"
	//ResourceName ...
	ResourceName = "resource_name"
	//ResourceCRN ...
	ResourceCRN = "resource_crn"
	//ResourceStatus ...
	ResourceStatus = "resource_status"
	//ResourceGroupName ...
	ResourceGroupName = "resource_group_name"
)

func getBaseController(meta interface{}) (string, error) {
	return prodBaseController, nil
}

func resourceIBMISSSHKey() *schema.Resource {
	return &schema.Resource{
		Create:   resourceIBMISSSHKeyCreate,
		Read:     resourceIBMISSSHKeyRead,
		Update:   resourceIBMISSSHKeyUpdate,
		Delete:   resourceIBMISSSHKeyDelete,
		Exists:   resourceIBMISSSHKeyExists,
		Importer: &schema.ResourceImporter{},



		Schema: map[string]*schema.Schema{
			isKeyName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     false,
				Description:  "SSH Key name",
			},

			isKeyPublicKey: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "SSH Public key data",
			},

			isKeyType: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key type",
			},

			isKeyFingerprint: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SSH key Fingerprint info",
			},

			isKeyLength: {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "SSH key Length",
			},

			isKeyResourceGroup: {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
				Description: "Resource group ID",
			},

			ResourceName: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the resource",
			},

			ResourceCRN: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The crn of the resource",
			},

			ResourceGroupName: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name in which resource is provisioned",
			},
		},
	}
}

func keyGetByName(d *schema.ResourceData, meta interface{}, name string) error {
	sess, err := vpcClient(meta)
	if err != nil {
		return err
	}
	listKeysOptions := &vpcv1.ListKeysOptions{}
	keys, response, err := sess.ListKeys(listKeysOptions)
	if err != nil {
		return fmt.Errorf("Error Fetching Keys %s\n%s", err, response)
	}
	for _, key := range keys.Keys {
		if *key.Name == name {
			d.SetId(*key.ID)
			d.Set(isKeyName, *key.Name)
			d.Set(isKeyType, *key.Type)
			d.Set(isKeyFingerprint, *key.Fingerprint)
			d.Set(isKeyLength, *key.Length)
			d.Set(ResourceName, *key.Name)
			d.Set(ResourceCRN, *key.CRN)
			if key.ResourceGroup != nil {
				d.Set(ResourceGroupName, *key.ResourceGroup.ID)
			}
			return nil
		}
	}
	return fmt.Errorf("No SSH Key found with name %s", name)
}

func resourceIBMISSSHKeyCreate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get(isKeyName).(string)
	publickey := d.Get(isKeyPublicKey).(string)

	err := keyGetByName(d, meta, name)
	if err != nil {
		err := keyCreate(d, meta, name, publickey)
		if err != nil {
			return err
		}
	}
	return resourceIBMISSSHKeyRead(d, meta)
}



func keyCreate(d *schema.ResourceData, meta interface{}, name, publickey string) error {
	sess, err := vpcClient(meta)
	if err != nil {
		return err
	}
	options := &vpcv1.CreateKeyOptions{
		PublicKey: &publickey,
		Name:      &name,
	}

	if rgrp, ok := d.GetOk(isKeyResourceGroup); ok {
		rg := rgrp.(string)
		options.ResourceGroup = &vpcv1.ResourceGroupIdentity{
			ID: &rg,
		}
	}

	key, response, err := sess.CreateKey(options)
	if err != nil {
		return fmt.Errorf("[DEBUG] Create SSH Key %s\n%s", err, response)
	}
	d.SetId(*key.ID)
	log.Printf("[INFO] Key : %s", *key.ID)

	return nil
}

func resourceIBMISSSHKeyRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	err := keyGet(d, meta, id)
	if err != nil {
		return err
	}
	return nil
}


func keyGet(d *schema.ResourceData, meta interface{}, id string) error {
	sess, err := vpcClient(meta)
	if err != nil {
		return err
	}
	options := &vpcv1.GetKeyOptions{
		ID: &id,
	}
	key, response, err := sess.GetKey(options)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error Getting SSH Key (%s): %s\n%s", id, err, response)
	}
	d.Set(isKeyName, *key.Name)
	d.Set(isKeyPublicKey, *key.PublicKey)
	d.Set(isKeyType, *key.Type)
	d.Set(isKeyFingerprint, *key.Fingerprint)
	d.Set(isKeyLength, *key.Length)
	d.Set(ResourceName, *key.Name)
	d.Set(ResourceCRN, *key.CRN)
	if key.ResourceGroup != nil {
		d.Set(ResourceGroupName, *key.ResourceGroup.Name)
		d.Set(isKeyResourceGroup, *key.ResourceGroup.ID)
	}
	return nil
}

func resourceIBMISSSHKeyUpdate(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	name := ""
	hasChanged := false

	if d.HasChange(isKeyName) {
		name = d.Get(isKeyName).(string)
		hasChanged = true
	}


	err := keyUpdate(d, meta, id, name, hasChanged)
	if err != nil {
		return err
	}
	return resourceIBMISSSHKeyRead(d, meta)
}


func keyUpdate(d *schema.ResourceData, meta interface{}, id, name string, hasChanged bool) error {
	sess, err := vpcClient(meta)
	if err != nil {
		return err
	}
	if hasChanged {
		options := &vpcv1.UpdateKeyOptions{
			ID: &id,
		}
		keyPatchModel := &vpcv1.KeyPatch{
			Name: &name,
		}
		keyPatch, err := keyPatchModel.AsPatch()
		if err != nil {
			return fmt.Errorf("Error calling asPatch for KeyPatch: %s", err)
		}
		options.KeyPatch = keyPatch
		_, response, err := sess.UpdateKey(options)
		if err != nil {
			return fmt.Errorf("Error updating vpc SSH Key: %s\n%s", err, response)
		}
	}
	return nil
}

func resourceIBMISSSHKeyDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	err := keyDelete(d, meta, id)
	if err != nil {
		return err
	}
	
	return nil
}

func keyDelete(d *schema.ResourceData, meta interface{}, id string) error {
	sess, err := vpcClient(meta)
	if err != nil {
		d.SetId("")
		return nil
	}

	getKeyOptions := &vpcv1.GetKeyOptions{
		ID: &id,
	}
	_, response, err := sess.GetKey(getKeyOptions)
	if err != nil {
		d.SetId("")
		return nil
	}
	if response != nil {}

	options := &vpcv1.DeleteKeyOptions{
		ID: &id,
	}
	response, err = sess.DeleteKey(options)
	if err != nil {
		//swallow
	}
	if response != nil {}
	d.SetId("")
	return nil
}

func resourceIBMISSSHKeyExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	id := d.Id()

	exists, err := keyExists(d, meta, id)
	return exists, err
}


func keyExists(d *schema.ResourceData, meta interface{}, id string) (bool, error) {
	sess, err := vpcClient(meta)
	if err != nil {
		return false, err
	}
	options := &vpcv1.GetKeyOptions{
		ID: &id,
	}
	_, response, err := sess.GetKey(options)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("Error getting SSH Key: %s\n%s", err, response)
	}
	return true, nil
}
