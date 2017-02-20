package lsx

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/registry"
	apitypes "github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/api/utils"
	apiconfig "github.com/codedellemc/libstorage/api/utils/config"

	// load these packages
	_ "github.com/codedellemc/libstorage/imports/config"
	_ "github.com/codedellemc/libstorage/imports/executors"
)

var cmds = map[string]func(
	ctx apitypes.Context,
	d apitypes.StorageExecutor,
	store apitypes.Store,
	exitCode *int,
	args ...string) (interface{}, error){

	apitypes.LSXCmdInstanceID:    opInstanceID,
	apitypes.LSXCmdLocalDevices:  opLocalDevices,
	apitypes.LSXCmdMount:         opMount,
	apitypes.LSXCmdMounts:        opMounts,
	apitypes.LSXCmdNextDevice:    opNextDevice,
	apitypes.LSXCmdSupported:     opSupported,
	apitypes.LSXCmdUmount:        opUmount,
	"unmount":                    opUmount,
	apitypes.LSXCmdVolumeAttach:  opVolumeAttach,
	apitypes.LSXCmdVolumeCreate:  opVolumeCreate,
	apitypes.LSXCmdVolumeDetach:  opVolumeDetach,
	apitypes.LSXCmdVolumeRemove:  opVolumeRemove,
	apitypes.LSXCmdWaitForDevice: opWait,
}

// Run runs the executor CLI.
func Run() {

	args := os.Args
	if len(args) < 3 {
		printUsageAndExit()
	}

	var (
		driverName = args[1]
		ctx        = context.Background()
	)

	if parts := strings.Split(driverName, ":"); len(parts) > 1 {
		driverName = parts[0]
		ctx = context.WithValue(ctx, context.ServiceKey, parts[1])
	}

	d, err := registry.NewStorageExecutor(driverName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	config, err := apiconfig.NewConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	apiconfig.UpdateLogLevel(config)

	if err := d.Init(ctx, config); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	opName := strings.ToLower(args[2])
	opFunc, ok := cmds[opName]
	if !ok {
		printUsageAndExit()
	}

	var (
		result   interface{}
		exitCode int
		store    = utils.NewStore()
	)

	if result, err = opFunc(ctx, d, store, &exitCode, args...); err != nil {
		// if the function is not implemented then exit with
		// apitypes.LSXExitCodeNotImplemented to let callers
		// know that the function is unsupported on this system
		exitCode = 1
		if strings.EqualFold(err.Error(), apitypes.ErrNotImplemented.Error()) {
			exitCode = apitypes.LSXExitCodeNotImplemented
		}
		fmt.Fprintf(os.Stderr,
			"error: error getting %s: %v\n", opName, err)
		os.Exit(exitCode)
	}

	switch tr := result.(type) {
	case bool:
		fmt.Fprintf(os.Stdout, "%v", result)
	case string:
		fmt.Fprintln(os.Stdout, result)
	case encoding.TextMarshaler:
		buf, err := tr.MarshalText()
		if err != nil {
			fmt.Fprintf(
				os.Stderr, "error: error encoding %s: %v\n", opName, err)
			os.Exit(1)
		}
		os.Stdout.Write(buf)
	default:
		buf, err := json.Marshal(result)
		if err != nil {
			fmt.Fprintf(
				os.Stderr, "error: error encoding %s: %v\n", opName, err)
			os.Exit(1)
		}
		if isNullBuf(buf) {
			os.Stdout.Write(emptyJSONBuff)
		} else {
			os.Stdout.Write(buf)
		}
	}

	os.Exit(exitCode)
}

const (
	newline = 10
)

var (
	nullBuff      = []byte{110, 117, 108, 108}
	emptyJSONBuff = []byte{123, 125}
)

func isNullBuf(buf []byte) bool {
	return len(buf) == len(nullBuff) &&
		buf[0] == nullBuff[0] && buf[1] == nullBuff[1] &&
		buf[2] == nullBuff[2] && buf[3] == nullBuff[3]
}

func executorNames() <-chan string {
	c := make(chan string)
	go func() {
		for se := range registry.StorageExecutors() {
			c <- strings.ToLower(se.Name())
		}
		close(c)
	}()
	return c
}

func opSupported(
	ctx apitypes.Context,
	d apitypes.StorageExecutor,
	store apitypes.Store,
	exitCode *int,
	args ...string) (interface{}, error) {

	dws, ok := d.(apitypes.StorageExecutorWithSupported)
	if !ok {
		return nil, apitypes.ErrNotImplemented
	}

	ok, err := dws.Supported(ctx, store)
	if err != nil {
		return nil, err
	}

	if !ok {
		return apitypes.LSXSOpNone, nil
	}

	rflags := apitypes.LSXSOpInstanceID |
		apitypes.LSXSOpLocalDevices |
		apitypes.LSXSOpNextDevice |
		apitypes.LSXSOpWaitForDevice
	if _, ok := dws.(apitypes.StorageExecutorWithMount); ok {
		rflags = rflags | apitypes.LSXSOpMount
	}
	if _, ok := dws.(apitypes.StorageExecutorWithUnmount); ok {
		rflags = rflags | apitypes.LSXSOpUmount
	}
	if _, ok := dws.(apitypes.StorageExecutorWithMounts); ok {
		rflags = rflags | apitypes.LSXSOpMounts
	}
	if _, ok := dws.(apitypes.StorageExecutorWithVolumeCreate); ok {
		rflags = rflags | apitypes.LSXSOpVolumeCreate
	}
	if _, ok := dws.(apitypes.StorageExecutorWithVolumeRemove); ok {
		rflags = rflags | apitypes.LSXSOpVolumeRemove
	}
	if _, ok := dws.(apitypes.StorageExecutorWithVolumeAttach); ok {
		rflags = rflags | apitypes.LSXSOpVolumeAttach
	}
	if _, ok := dws.(apitypes.StorageExecutorWithVolumeDetach); ok {
		rflags = rflags | apitypes.LSXSOpVolumeDetach
	}

	return rflags, nil
}

func opMounts(
	ctx apitypes.Context,
	d apitypes.StorageExecutor,
	store apitypes.Store,
	exitCode *int,
	args ...string) (interface{}, error) {

	dd, ok := d.(apitypes.StorageExecutorWithMounts)
	if !ok {
		return nil, apitypes.ErrNotImplemented
	}

	mounts, err := dd.Mounts(ctx, store)
	if err != nil {
		return nil, err
	}

	if mounts == nil {
		return []*apitypes.MountInfo{}, nil
	}

	return mounts, nil
}

func opMount(
	ctx apitypes.Context,
	d apitypes.StorageExecutor,
	store apitypes.Store,
	exitCode *int,
	args ...string) (interface{}, error) {

	dd, ok := d.(apitypes.StorageExecutorWithMount)
	if !ok {
		return nil, apitypes.ErrNotImplemented
	}

	var (
		deviceName string
		mountPath  string
		mountOpts  = &apitypes.DeviceMountOpts{Opts: store}
	)
	mountArgs := args[3:]
	if len(mountArgs) == 0 {
		printUsageAndExit()
	}

	remArgs := []string{}
	for x := 0; x < len(mountArgs); {
		a := mountArgs[x]
		if x < len(mountArgs)-1 {
			switch a {
			case "-l":
				mountOpts.MountLabel = mountArgs[x+1]
				x = x + 2
				continue
			case "-o":
				mountOpts.MountOptions = mountArgs[x+1]
				x = x + 2
				continue
			}
		}
		remArgs = append(remArgs, a)
		x++
	}

	if len(remArgs) != 2 {
		printUsageAndExit()
	}

	deviceName = remArgs[0]
	mountPath = remArgs[1]

	if err := dd.Mount(ctx, deviceName, mountPath, mountOpts); err != nil {
		return nil, err
	}

	return mountPath, nil
}

func opUmount(
	ctx apitypes.Context,
	d apitypes.StorageExecutor,
	store apitypes.Store,
	exitCode *int,
	args ...string) (interface{}, error) {

	dd, ok := d.(apitypes.StorageExecutorWithUnmount)
	if !ok {
		return nil, apitypes.ErrNotImplemented
	}

	if len(args) < 4 {
		printUsageAndExit()
	}
	mountPath := args[3]
	if err := dd.Unmount(ctx, mountPath, store); err != nil {
		return nil, err
	}

	return mountPath, nil
}

func opVolumeCreate(
	ctx apitypes.Context,
	d apitypes.StorageExecutor,
	store apitypes.Store,
	exitCode *int,
	args ...string) (interface{}, error) {

	dd, ok := d.(apitypes.StorageExecutorWithVolumeCreate)
	if !ok {
		return nil, apitypes.ErrNotImplemented
	}

	args = args[3:]
	if len(args) == 0 {
		printUsageAndExit()
	}

	var (
		opts    = &apitypes.VolumeCreateOpts{Opts: store}
		remArgs = []string{}
	)

	for x := 0; x < len(args); {
		a := args[x]
		if a == "-e" {
			encrypted := true
			opts.Encrypted = &encrypted
			x = x + 1
			continue
		}
		if x < len(args)-1 {
			switch a {
			case "-k":
				opts.EncryptionKey = &(args[x+1])
				x = x + 2
				continue
			case "-i":
				i, err := strconv.Atoi(args[x+1])
				if err != nil {
					printUsageAndExit()
				}
				i64 := int64(i)
				opts.IOPS = &i64
				x = x + 2
				continue
			case "-s":
				i, err := strconv.Atoi(args[x+1])
				if err != nil {
					printUsageAndExit()
				}
				i64 := int64(i)
				opts.Size = &i64
				x = x + 2
				continue
			case "-t":
				opts.Type = &(args[x+1])
				x = x + 2
				continue
			case "-z":
				opts.AvailabilityZone = &(args[x+1])
				x = x + 2
				continue
			}
		}
		remArgs = append(remArgs, a)
		x++
	}

	if len(remArgs) != 1 {
		printUsageAndExit()
	}

	vol, err := dd.VolumeCreate(ctx, remArgs[0], opts)
	if err != nil {
		return nil, err
	}

	return vol, nil
}

func opVolumeRemove(
	ctx apitypes.Context,
	d apitypes.StorageExecutor,
	store apitypes.Store,
	exitCode *int,
	args ...string) (interface{}, error) {

	dd, ok := d.(apitypes.StorageExecutorWithVolumeRemove)
	if !ok {
		return nil, apitypes.ErrNotImplemented
	}

	args = args[3:]
	if len(args) == 0 {
		printUsageAndExit()
	}

	var (
		opts    = &apitypes.VolumeRemoveOpts{Opts: store}
		remArgs = []string{}
	)

	for x := 0; x < len(args); {
		a := args[x]
		if a == "-f" {
			opts.Force = true
			x = x + 1
			continue
		}
		remArgs = append(remArgs, a)
		x++
	}

	if len(remArgs) != 1 {
		printUsageAndExit()
	}

	if err := dd.VolumeRemove(ctx, remArgs[0], opts); err != nil {
		return nil, err
	}

	return nil, nil
}

func opVolumeAttach(
	ctx apitypes.Context,
	d apitypes.StorageExecutor,
	store apitypes.Store,
	exitCode *int,
	args ...string) (interface{}, error) {

	dd, ok := d.(apitypes.StorageExecutorWithVolumeAttach)
	if !ok {
		return nil, apitypes.ErrNotImplemented
	}

	args = args[3:]
	if len(args) == 0 {
		printUsageAndExit()
	}

	var (
		opts    = &apitypes.VolumeAttachOpts{Opts: store}
		remArgs = []string{}
	)

	for x := 0; x < len(args); {
		a := args[x]
		if a == "-f" {
			opts.Force = true
			x = x + 1
			continue
		}
		if x < len(args)-1 && a == "-n" {
			opts.NextDevice = &(args[x+1])
			x = x + 2
			continue
		}
		remArgs = append(remArgs, a)
		x++
	}

	if len(remArgs) != 1 {
		printUsageAndExit()
	}

	if opts.NextDevice == nil {
		nd, err := d.NextDevice(ctx, store)
		if err != nil && err != apitypes.ErrNotImplemented {
			return nil, err
		}
		opts.NextDevice = &nd
	}

	vol, tok, err := dd.VolumeAttach(ctx, remArgs[0], opts)
	if err != nil {
		return nil, err
	}

	return &apitypes.LSXVolumeAttachResult{Volume: vol, Token: tok}, nil
}

func opVolumeDetach(
	ctx apitypes.Context,
	d apitypes.StorageExecutor,
	store apitypes.Store,
	exitCode *int,
	args ...string) (interface{}, error) {

	dd, ok := d.(apitypes.StorageExecutorWithVolumeDetach)
	if !ok {
		return nil, apitypes.ErrNotImplemented
	}

	args = args[3:]
	if len(args) == 0 {
		printUsageAndExit()
	}

	var (
		opts    = &apitypes.VolumeDetachOpts{Opts: store}
		remArgs = []string{}
	)

	for x := 0; x < len(args); {
		a := args[x]
		if a == "-f" {
			opts.Force = true
			x = x + 1
			continue
		}
		remArgs = append(remArgs, a)
		x++
	}

	if len(remArgs) != 1 {
		printUsageAndExit()
	}

	vol, err := dd.VolumeDetach(ctx, remArgs[0], opts)
	if err != nil {
		return nil, err
	}

	return vol, nil
}

func opInstanceID(
	ctx apitypes.Context,
	d apitypes.StorageExecutor,
	store apitypes.Store,
	exitCode *int,
	args ...string) (interface{}, error) {

	result, err := d.InstanceID(ctx, store)
	if err != nil {
		return nil, err
	}
	result.Driver = d.Name()
	return result, nil
}

func opNextDevice(
	ctx apitypes.Context,
	d apitypes.StorageExecutor,
	store apitypes.Store,
	exitCode *int,
	args ...string) (interface{}, error) {

	result, err := d.NextDevice(ctx, store)
	if err != nil {
		if err != apitypes.ErrNotImplemented {
			return nil, err
		}
		return nil, nil
	}
	return result, nil
}

func opLocalDevices(
	ctx apitypes.Context,
	d apitypes.StorageExecutor,
	store apitypes.Store,
	exitCode *int,
	args ...string) (interface{}, error) {

	if len(args) < 4 {
		printUsageAndExit()
	}
	result, err := d.LocalDevices(ctx, &apitypes.LocalDevicesOpts{
		ScanType: apitypes.ParseDeviceScanType(args[3]),
		Opts:     store,
	})
	if err != nil {
		return nil, err
	}
	result.Driver = d.Name()
	return result, nil
}

func opWait(
	ctx apitypes.Context,
	d apitypes.StorageExecutor,
	store apitypes.Store,
	exitCode *int,
	args ...string) (interface{}, error) {

	if len(args) < 6 {
		printUsageAndExit()
	}

	opts := &apitypes.WaitForDeviceOpts{
		LocalDevicesOpts: apitypes.LocalDevicesOpts{
			ScanType: apitypes.ParseDeviceScanType(args[3]),
			Opts:     store,
		},
		Token:   strings.ToLower(args[4]),
		Timeout: utils.DeviceAttachTimeout(args[5]),
	}

	ldl := func() (bool, *apitypes.LocalDevices, error) {
		ldm, err := d.LocalDevices(ctx, &opts.LocalDevicesOpts)
		if err != nil {
			return false, nil, err
		}
		for k := range ldm.DeviceMap {
			if strings.ToLower(k) == opts.Token {
				return true, ldm, nil
			}
		}
		return false, ldm, nil
	}

	var (
		timeoutC = time.After(opts.Timeout)
		tick     = time.Tick(500 * time.Millisecond)
	)

	for {
		select {
		case <-timeoutC:
			*exitCode = apitypes.LSXExitCodeTimedOut
			return nil, nil
		case <-tick:
			found, result, err := ldl()
			if err != nil {
				return nil, err
			}
			if found {
				result.Driver = d.Name()
				return result, nil
			}
		}
	}
}

func printUsage() {
	buf := &bytes.Buffer{}
	w := io.MultiWriter(buf, os.Stderr)

	fmt.Fprintf(w, "usage: ")
	lpad1 := buf.Len()
	fmt.Fprintf(w, "%s <executor>[:<service>] ", os.Args[0])
	lpad2 := buf.Len()
	fmt.Fprintf(w, "supported\n")
	printUsageLeftPadded(w, lpad2, "instanceID\n")
	printUsageLeftPadded(w, lpad2, "nextDevice\n")
	printUsageLeftPadded(w, lpad2, "localDevices <scanType>\n")
	printUsageLeftPadded(w, lpad2, "wait <scanType> <attachToken> <timeout>\n")
	printUsageLeftPadded(w, lpad2, "mounts\n")
	printUsageLeftPadded(
		w, lpad2, "mount [-l label] [-o options] <device> <path>\n")
	printUsageLeftPadded(w, lpad2, "umount <path>\n")
	printUsageLeftPadded(
		w, lpad2, "volumeCreate [-e] [-k encryptionKey] "+
			"[-i iops] [-s size] [-t type] [-z zone] <name>\n")
	printUsageLeftPadded(w, lpad2, "volumeRemove [-f] <id>\n")
	printUsageLeftPadded(w, lpad2, "volumeAttach [-n nextDevice] [-f] <id>\n")
	printUsageLeftPadded(w, lpad2, "volumeDetach [-f] <id>\n")
	fmt.Fprintln(w)
	executorVar := "executor:    "
	printUsageLeftPadded(w, lpad1, executorVar)
	lpad3 := lpad1 + len(executorVar)

	execNames := []string{}
	for en := range executorNames() {
		execNames = append(execNames, en)
	}

	if len(execNames) > 0 {
		execNames = utils.SortByString(execNames)
		fmt.Fprintf(w, "%s\n", execNames[0])
		if len(execNames) > 1 {
			for x, en := range execNames {
				if x == 0 {
					continue
				}
				printUsageLeftPadded(w, lpad3, "%s\n", en)
			}
		}
		fmt.Fprintln(w)
	}

	printUsageLeftPadded(w, lpad1, "scanType:    0,quick | 1,deep\n\n")
	printUsageLeftPadded(w, lpad1, "attachToken: <token>\n\n")
	printUsageLeftPadded(w, lpad1, "timeout:     30s | 1h | 5m\n\n")
}

func printUsageLeftPadded(
	w io.Writer, lpadLen int, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	lpadFmt := fmt.Sprintf("%%%ds", lpadLen+len(text))
	fmt.Fprintf(w, lpadFmt, text)
}

func printUsageAndExit() {
	printUsage()
	os.Exit(1)
}
