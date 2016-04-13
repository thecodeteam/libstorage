package snapshot

import (
	"net/http"
	"strings"

	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/server/services"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	apihttp "github.com/emccode/libstorage/api/types/http"
	apisvcs "github.com/emccode/libstorage/api/types/services"
	"github.com/emccode/libstorage/api/utils"
	"github.com/emccode/libstorage/api/utils/schema"
)

//the filtering mechanism applies a simple match, you could do something like
//this future https://github.com/golang/appengine/blob/master/datastore/query.go
func applyFilter(obj *types.Volume, filters map[string][]string) bool {
	include := true
	for key, values := range filters {
		//fmt.Print("Filter Key: ", key, "\n")
		if len(obj.Fields[key]) == 0 {
			//fmt.Print("Key ", key, " not found\n")
			include = false
			break
		}
		if !include {
			//fmt.Print("Exiting early with no key found\n")
			break
		}

		found := false
		for _, value := range values {
			//fmt.Print("Filter Val: ", value, "\n")
			//omit adding to the slice if the key and value doesnt exist
			if strings.Compare(value, obj.Fields[key]) == 0 {
				//fmt.Print(value, " = ", obj.Fields[key], "\n")
				found = true //key exists and value exists in the map
				break
			}
		}
		if !found {
			//fmt.Print("Exiting early with no value found\n")
			include = false
			break
		}

		//fmt.Print("Found: ", found, "\n")
		include = include && found
		if !include {
			//fmt.Print("Exiting early with no key found\n")
			break
		}
	}

	return include
}

func (r *router) snapshots(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	var (
		tasks   = map[string]*types.Task{}
		taskIDs []int
		reply   apihttp.ServiceSnapshotMap = map[string]apihttp.SnapshotMap{}
	)

	//filtering is done by query parameters on the URI
	var filters map[string][]string
	filters = req.URL.Query()

	for service := range services.StorageServices() {

		run := func(
			ctx context.Context,
			svc apisvcs.StorageService) (interface{}, error) {

			objs, err := svc.Driver().Snapshots(ctx, store)
			if err != nil {
				return nil, err
			}

			objMap := map[string]*types.Snapshot{}
			for _, obj := range objs {
				if !applyFilter(obj, filters) {
					continue //object didnt not meet filter requirements
				}
				objMap[obj.ID] = obj
			}
			return objMap, nil
		}

		task := service.TaskExecute(ctx, run, schema.SnapshotMapSchema)
		taskIDs = append(taskIDs, task.ID)
		tasks[service.Name()] = task
	}

	run := func(ctx context.Context) (interface{}, error) {

		services.TaskWaitAll(taskIDs...)

		for k, v := range tasks {
			if v.Error != nil {
				return nil, utils.NewBatchProcessErr(reply, v.Error)
			}

			objMap, ok := v.Result.(map[string]*types.Snapshot)
			if !ok {
				return nil, utils.NewBatchProcessErr(
					reply, goof.New("error casting to []*types.Snapshot"))
			}
			reply[k] = objMap
		}

		return reply, nil
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		services.TaskExecute(ctx, run, schema.ServiceSnapshotMapSchema),
		http.StatusOK)
}

func (r *router) snapshotsForService(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	//filtering is done by query parameters on the URI
	var filters map[string][]string
	filters = req.URL.Query()

	run := func(
		ctx context.Context,
		svc apisvcs.StorageService) (interface{}, error) {

		var reply apihttp.SnapshotMap = map[string]*types.Snapshot{}

		objs, err := svc.Driver().Snapshots(ctx, store)
		if err != nil {
			return nil, err
		}

		for _, obj := range objs {
			if !applyFilter(obj, filters) {
				continue //object didnt not meet filter requirements
			}
			reply[obj.ID] = obj
		}
		return reply, nil
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		service.TaskExecute(ctx, run, schema.SnapshotSchema),
		http.StatusOK)
}

func (r *router) snapshotInspect(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	run := func(
		ctx context.Context,
		svc apisvcs.StorageService) (interface{}, error) {

		return svc.Driver().SnapshotInspect(
			ctx,
			store.GetString("snapshotID"),
			store)
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		service.TaskExecute(ctx, run, schema.SnapshotSchema),
		http.StatusOK)
}

func (r *router) snapshotRemove(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	run := func(
		ctx context.Context,
		svc apisvcs.StorageService) (interface{}, error) {

		return nil, svc.Driver().SnapshotRemove(
			ctx,
			store.GetString("snapshotID"),
			store)
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		service.TaskExecute(ctx, run, nil),
		http.StatusResetContent)
}

func (r *router) volumeCreate(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	run := func(
		ctx context.Context,
		svc apisvcs.StorageService) (interface{}, error) {

		return svc.Driver().VolumeCreateFromSnapshot(
			ctx,
			store.GetString("snapshotID"),
			store.GetString("volumeName"),
			store)
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		service.TaskExecute(ctx, run, schema.VolumeSchema),
		http.StatusCreated)
}

func (r *router) snapshotCopy(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	run := func(
		ctx context.Context,
		svc apisvcs.StorageService) (interface{}, error) {

		return svc.Driver().SnapshotCopy(
			ctx,
			store.GetString("snapshotID"),
			store.GetString("snapshotName"),
			store.GetString("destinationID"),
			store)
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		service.TaskExecute(ctx, run, schema.SnapshotSchema),
		http.StatusCreated)
}
