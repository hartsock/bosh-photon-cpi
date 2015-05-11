package main

import (
	"errors"
	"github.com/esxcloud/bosh-esxcloud-cpi/cpi"
	ec "github.com/esxcloud/esxcloud-go-sdk/esxcloud"
	"math"
	"net/http"
)

func CreateDisk(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	size, ok := args[0].(int)
	if !ok {
		return nil, errors.New("Unexpected argument where size should be")
	}
	size = toGB(size)
	if size < 1 {
		return nil, errors.New("Must provide a size in MiB that rounds up to at least 1 GiB for esxcloud")
	}
	vmCID, ok := args[1].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where vm_cid should be")
	}

	diskSpec := &ec.DiskCreateSpec{
		Flavor:     ctx.Config.ESXCloud.DiskFlavor,
		Kind:       "persistent-disk",
		CapacityGB: size,
		Name:       "disk-for-vm-" + vmCID,
	}

	task, err := ctx.Client.Projects.CreateDisk(ctx.Config.ESXCloud.ProjectID, diskSpec)
	if err != nil {
		return
	}
	task, err = ctx.Client.Tasks.Wait(task.ID)
	if err != nil {
		return
	}
	return task.Entity.ID, nil
}

func DeleteDisk(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	diskCID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where disk_cid should be")
	}
	task, err := ctx.Client.Disks.Delete(diskCID, true)
	if err != nil {
		return
	}
	task, err = ctx.Client.Tasks.Wait(task.ID)
	if err != nil {
		return
	}
	return nil, nil
}

func HasDisk(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	diskCID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where disk_cid should be")
	}
	_, err = ctx.Client.Disks.Get(diskCID)
	if err != nil {
		apiErr, ok := err.(ec.ApiError)
		if ok && apiErr.HttpStatusCode == http.StatusNotFound {
			return false, nil
		}
		return nil, err
	}
	return true, nil
}

func GetDisks(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	vmCID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where vim_cid should be")
	}
	disks, err := ctx.Client.Projects.FindDisks(ctx.Config.ESXCloud.ProjectID, nil)
	if err != nil {
		return
	}
	res := []string{}
	for _, disk := range disks.Items {
		for _, vm := range disk.VMs {
			if vm == vmCID {
				res = append(res, disk.ID)
			}
		}
	}
	return res, nil
}

func AttachDisk(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	vmCID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where vm_cid should be")
	}
	diskCID, ok := args[1].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where disk_cid should be")
	}
	op := &ec.VmDiskOperation{DiskID: diskCID}
	task, err := ctx.Client.VMs.AttachDisk(vmCID, op)
	if err != nil {
		return
	}
	task, err = ctx.Client.Tasks.Wait(task.ID)
	if err != nil {
		return
	}
	return nil, nil
}

func DetachDisk(ctx *cpi.Context, args []interface{}) (result interface{}, err error) {
	vmCID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where vm_cid should be")
	}
	diskCID, ok := args[1].(string)
	if !ok {
		return nil, errors.New("Unexpected argument where disk_cid should be")
	}
	op := &ec.VmDiskOperation{DiskID: diskCID}
	task, err := ctx.Client.VMs.DetachDisk(vmCID, op)
	if err != nil {
		return
	}
	task, err = ctx.Client.Tasks.Wait(task.ID)
	if err != nil {
		return
	}
	return nil, nil
}

func toGB(mb int) int {
	return int(math.Ceil(float64(mb) / 1000.0))
}
