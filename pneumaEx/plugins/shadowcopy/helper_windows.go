package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

//507C37B4-CF5B-4e95-B0AF-14EB9767467E
var IID_IVSS_ASYNC = &ole.GUID{
	Data1: 0x507C37B4,
	Data2: 0xCF5B,
	Data3: 0x4e95,
	Data4: [8]byte{0xb0, 0xaf, 0x14, 0xeb, 0x97, 0x67, 0x46, 0x7e}}

type IVSSAsync struct {
	ole.IUnknown
}

type Returner struct {
	SnapshotID   string `json:"SnapshotID"`
	DeviceObject string `json:"DeviceObject"`
	SymLink      string `json:"SymLink"`
	Success      bool   `json:"Success"`
}

type SymLink struct {
	Path string `json:"Path"`
}

type ReturnerDelete struct {
	Message string `json:"Message"`
	Success bool   `json:"Success"`
}

type IVSSAsyncVtbl struct {
	ole.IUnknownVtbl
	cancel      uintptr
	wait        uintptr
	queryStatus uintptr
}

func (async *IVSSAsync) VTable() *IVSSAsyncVtbl {
	return (*IVSSAsyncVtbl)(unsafe.Pointer(async.RawVTable))
}

var VSS_S_ASYNC_PENDING int32 = 0x00042309
var VSS_S_ASYNC_FINISHED int32 = 0x0004230A
var VSS_S_ASYNC_CANCELLED int32 = 0x0004230B

func (async *IVSSAsync) Wait(seconds int) bool {

	startTime := time.Now().Unix()
	for {
		ret, _, _ := syscall.Syscall(async.VTable().wait, 2, uintptr(unsafe.Pointer(async)), uintptr(1000), 0)
		if ret != 0 {
			fmt.Println("IVSSASYNC_WAIT", "IVssAsync::Wait returned %d\n", ret)
		}

		var status int32
		ret, _, _ = syscall.Syscall(async.VTable().queryStatus, 3, uintptr(unsafe.Pointer(async)),
			uintptr(unsafe.Pointer(&status)), 0)
		if ret != 0 {
			fmt.Println("IVSSASYNC_QUERY", "IVssAsync::QueryStatus returned %d\n", ret)
		}

		if status == VSS_S_ASYNC_FINISHED {
			return true
		}
		if time.Now().Unix()-startTime > int64(seconds) {
			fmt.Println("IVSSASYNC_TIMEOUT", "IVssAsync is pending for more than %d seconds\n", seconds)
			return false
		}
	}
}

func getIVSSAsync(unknown *ole.IUnknown, iid *ole.GUID) (async *IVSSAsync) {
	r, _, _ := syscall.Syscall(
		unknown.VTable().QueryInterface,
		3,
		uintptr(unsafe.Pointer(unknown)),
		uintptr(unsafe.Pointer(iid)),
		uintptr(unsafe.Pointer(&async)))

	if r != 0 {
		fmt.Println("IVSSASYNC_QUERY", "IVSSAsync::QueryInterface returned %d\n", r)
		return nil
	}
	return
}

var IID_IVSS = &ole.GUID{
	Data1: 0x665c1d5f,
	Data2: 0xc218,
	Data3: 0x414d,
	Data4: [8]byte{0xa0, 0x5d, 0x7f, 0xef, 0x5f, 0x9d, 0x5c, 0x86}}

type IVSS struct {
	ole.IUnknown
}

type IVSSVtbl struct {
	ole.IUnknownVtbl
	getWriterComponentsCount      uintptr
	getWriterComponents           uintptr
	initializeForBackup           uintptr
	setBackupState                uintptr
	initializeForRestore          uintptr
	setRestoreState               uintptr
	gatherWriterMetadata          uintptr
	getWriterMetadataCount        uintptr
	getWriterMetadata             uintptr
	freeWriterMetadata            uintptr
	addComponent                  uintptr
	prepareForBackup              uintptr
	abortBackup                   uintptr
	gatherWriterStatus            uintptr
	getWriterStatusCount          uintptr
	freeWriterStatus              uintptr
	getWriterStatus               uintptr
	setBackupSucceeded            uintptr
	setBackupOptions              uintptr
	setSelectedForRestore         uintptr
	setRestoreOptions             uintptr
	setAdditionalRestores         uintptr
	setPreviousBackupStamp        uintptr
	saveAsXML                     uintptr
	backupComplete                uintptr
	addAlternativeLocationMapping uintptr
	addRestoreSubcomponent        uintptr
	setFileRestoreStatus          uintptr
	addNewTarget                  uintptr
	setRangesFilePath             uintptr
	preRestore                    uintptr
	postRestore                   uintptr
	setContext                    uintptr
	startSnapshotSet              uintptr
	addToSnapshotSet              uintptr
	doSnapshotSet                 uintptr
	deleteSnapshots               uintptr
	importSnapshots               uintptr
	breakSnapshotSet              uintptr
	getSnapshotProperties         uintptr
	query                         uintptr
	isVolumeSupported             uintptr
	disableWriterClasses          uintptr
	enableWriterClasses           uintptr
	disableWriterInstances        uintptr
	exposeSnapshot                uintptr
	revertToSnapshot              uintptr
	queryRevertStatus             uintptr
}

func (vss *IVSS) VTable() *IVSSVtbl {
	return (*IVSSVtbl)(unsafe.Pointer(vss.RawVTable))
}

func (vss *IVSS) InitializeForBackup() int {
	ret, _, _ := syscall.Syscall(vss.VTable().initializeForBackup, 2, uintptr(unsafe.Pointer(vss)), 0, 0)
	return int(ret)
}

func (vss *IVSS) GatherWriterMetadata() (int, *IVSSAsync) {
	var unknown *ole.IUnknown
	ret, _, _ := syscall.Syscall(vss.VTable().gatherWriterMetadata, 2,
		uintptr(unsafe.Pointer(vss)),
		uintptr(unsafe.Pointer(&unknown)), 0)

	if ret != 0 {
		return int(ret), nil
	} else {
		return int(ret), getIVSSAsync(unknown, IID_IVSS_ASYNC)
	}
}

func (vss *IVSS) StartSnapshotSet(snapshotID *ole.GUID) int {
	ret, _, _ := syscall.Syscall(vss.VTable().startSnapshotSet, 2,
		uintptr(unsafe.Pointer(vss)),
		uintptr(unsafe.Pointer(snapshotID)), 0)
	return int(ret)
}

func (vss *IVSS) AddToSnapshotSet(drive string, snapshotID *ole.GUID) int {

	volumeName := syscall.StringToUTF16Ptr(drive)

	var ret uintptr
	if runtime.GOARCH == "386" {
		ret, _, _ = syscall.Syscall9(vss.VTable().addToSnapshotSet, 7,
			uintptr(unsafe.Pointer(vss)),
			uintptr(unsafe.Pointer(volumeName)),
			0, 0, 0, 0,
			uintptr(unsafe.Pointer(snapshotID)), 0, 0)
	} else {
		ret, _, _ = syscall.Syscall6(vss.VTable().addToSnapshotSet, 4,
			uintptr(unsafe.Pointer(vss)),
			uintptr(unsafe.Pointer(volumeName)),
			uintptr(unsafe.Pointer(ole.IID_NULL)),
			uintptr(unsafe.Pointer(snapshotID)), 0, 0)
	}
	return int(ret)
}

func (vss *IVSS) SetBackupState() int {
	VSS_BT_COPY := 5
	ret, _, _ := syscall.Syscall6(vss.VTable().setBackupState, 4,
		uintptr(unsafe.Pointer(vss)),
		0, 0, uintptr(VSS_BT_COPY), 0, 0)
	return int(ret)
}

func (vss *IVSS) PrepareForBackup() (int, *IVSSAsync) {
	var unknown *ole.IUnknown
	ret, _, _ := syscall.Syscall(vss.VTable().prepareForBackup, 2,
		uintptr(unsafe.Pointer(vss)),
		uintptr(unsafe.Pointer(&unknown)), 0)

	if ret != 0 {
		return int(ret), nil
	} else {
		return int(ret), getIVSSAsync(unknown, IID_IVSS_ASYNC)
	}
}

func (vss *IVSS) DoSnapshotSet() (int, *IVSSAsync) {
	var unknown *ole.IUnknown
	ret, _, _ := syscall.Syscall(vss.VTable().doSnapshotSet, 2,
		uintptr(unsafe.Pointer(vss)),
		uintptr(unsafe.Pointer(&unknown)), 0)

	if ret != 0 {
		return int(ret), nil
	} else {
		return int(ret), getIVSSAsync(unknown, IID_IVSS_ASYNC)
	}
}

type SnapshotProperties struct {
	SnapshotID           ole.GUID
	SnapshotSetID        ole.GUID
	SnapshotsCount       uint32
	SnapshotDeviceObject *uint16
	OriginalVolumeName   *uint16
	OriginatingMachine   *uint16
	ServiceMachine       *uint16
	ExposedName          *uint16
	ExposedPath          *uint16
	ProviderId           ole.GUID
	SnapshotAttributes   uint32
	CreationTimestamp    int64
	Status               int
}

func (vss *IVSS) GetSnapshotProperties(snapshotSetID ole.GUID, properties *SnapshotProperties) int {
	var ret uintptr
	if runtime.GOARCH == "386" {
		address := uint(uintptr(unsafe.Pointer(&snapshotSetID)))
		ret, _, _ = syscall.Syscall6(vss.VTable().getSnapshotProperties, 6,
			uintptr(unsafe.Pointer(vss)),
			uintptr(*(*uint32)(unsafe.Pointer(uintptr(address)))),
			uintptr(*(*uint32)(unsafe.Pointer(uintptr(address + 4)))),
			uintptr(*(*uint32)(unsafe.Pointer(uintptr(address + 8)))),
			uintptr(*(*uint32)(unsafe.Pointer(uintptr(address + 12)))),
			uintptr(unsafe.Pointer(properties)))
	} else {
		ret, _, _ = syscall.Syscall(vss.VTable().getSnapshotProperties, 3,
			uintptr(unsafe.Pointer(vss)),
			uintptr(unsafe.Pointer(&snapshotSetID)),
			uintptr(unsafe.Pointer(properties)))
	}
	return int(ret)
}

func uint16ArrayToString(p *uint16) string {
	if p == nil {
		return ""
	}
	s := make([]uint16, 0)
	address := uintptr(unsafe.Pointer(p))
	for {
		c := *(*uint16)(unsafe.Pointer(address))
		if c == 0 {
			break
		}

		s = append(s, c)
		address = uintptr(int(address) + 2)
	}

	return syscall.UTF16ToString(s)
}

func getIVSS(unknown *ole.IUnknown, iid *ole.GUID) (ivss *IVSS) {
	r, _, _ := syscall.Syscall(
		unknown.VTable().QueryInterface,
		3,
		uintptr(unsafe.Pointer(unknown)),
		uintptr(unsafe.Pointer(iid)),
		uintptr(unsafe.Pointer(&ivss)))

	if r != 0 {
		fmt.Println("IVSS_QUERY", "IVSS::QueryInterface returned %d\n", r)
		return nil
	}

	return ivss
}

var vssBackupComponent *IVSS
var snapshotID ole.GUID
var shadowLink string

func CreateShadowCopy(volumeletter string) string {

	volume := volumeletter + ":\\"
	timeoutInSeconds := 300

	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	dllVssApi := syscall.NewLazyDLL("VssApi.dll")
	procCreateVssBackupComponents :=
		dllVssApi.NewProc("?CreateVssBackupComponents@@YAJPEAPEAVIVssBackupComponents@@@Z")
	if runtime.GOARCH == "386" {
		procCreateVssBackupComponents =
			dllVssApi.NewProc("?CreateVssBackupComponents@@YGJPAPAVIVssBackupComponents@@@Z")
	}

	var unknown *ole.IUnknown
	r, _, err := procCreateVssBackupComponents.Call(uintptr(unsafe.Pointer(&unknown)))

	if r == 0x80070005 {
		return string("Error: Only administrators can create shadow copies")
	}

	if r != 0 {
		return string("Error: Failed to create the VSS backup component")
	}

	vssBackupComponent = getIVSS(unknown, IID_IVSS)
	if vssBackupComponent == nil {
		return string("Error: Failed to create the VSS backup component")
	}
	ret := vssBackupComponent.InitializeForBackup()
	if ret != 0 {
		return string("Error: Shadow copy creation failed: InitializeForBackup")
	}

	var async *IVSSAsync
	ret, async = vssBackupComponent.GatherWriterMetadata()
	if ret != 0 {
		return string("Error: Shadow copy creation failed: GatherWriterMetadata")
	}
	if async == nil {
		return string("Shadow copy creation failed: GatherWriterMetadata failed to return a valid IVssAsync object")
	}
	if !async.Wait(timeoutInSeconds) {
		return string("Shadow copy creation failed: GatherWriterMetadata didn't finish properly")
	}
	async.Release()

	var snapshotSetID ole.GUID

	ret = vssBackupComponent.StartSnapshotSet(&snapshotSetID)
	if ret != 0 {
		return string("Error: Shadow copy creation failed: StartSnapshotSet")
	}

	ret = vssBackupComponent.AddToSnapshotSet(volume, &snapshotID)
	if ret != 0 {
		return string("Error: Shadow copy creation failed: AddToSnapshotSet")
	}

	ret = vssBackupComponent.SetBackupState()
	if ret != 0 {
		return string("Error: Shadow copy creation failed: SetBackupState")
	}

	ret, async = vssBackupComponent.PrepareForBackup()
	if ret != 0 {
		return string("Error: Shadow copy creation failed: PrepareForBackup")
	}
	if async == nil {
		return string("Shadow copy creation failed: PrepareForBackup failed to return a valid IVssAsync object")
	}

	if !async.Wait(timeoutInSeconds) {
		return string("Shadow copy creation failed: PrepareForBackup didn't finish properly")
	}
	async.Release()

	ret, async = vssBackupComponent.DoSnapshotSet()
	if ret != 0 {
		return string("Error: Shadow copy creation failed: DoSnapshotSet")
	}
	if async == nil {
		return string("Shadow copy creation failed: DoSnapshotSet failed to return a valid IVssAsync object")
	}

	if !async.Wait(timeoutInSeconds) {
		return string("Shadow copy creation failed: DoSnapshotSet didn't finish properly")
	}
	async.Release()

	properties := SnapshotProperties{}

	ret = vssBackupComponent.GetSnapshotProperties(snapshotID, &properties)
	if ret != 0 {
		return string("Error: Shadow copy creation failed: GetSnapshotProperties")
	}

	SnapshotIDString, _ := ole.StringFromIID(&properties.SnapshotID)
	snapshotPath := uint16ArrayToString(properties.SnapshotDeviceObject)

	preferencePath := volumeletter + ":"
	shadowLink = preferencePath + "\\shadow"
	os.Remove(shadowLink)
	err = os.Symlink(snapshotPath+"\\", shadowLink)
	if err != nil {
		return string("Failed to create a symbolic link to the shadow copy")
	}
	shadowLink := strings.Replace(shadowLink, "\\", "\\\\", 1)
	toReturn := Returner{
		SnapshotID:   SnapshotIDString,
		DeviceObject: snapshotPath,
		SymLink:      shadowLink,
		Success:      true,
	}

	toResult, err := json.Marshal(toReturn)
	return string(toResult)
}
