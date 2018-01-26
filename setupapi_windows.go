// +build windows

package setupapi

import (
	"unsafe"

	"github.com/oakmound/w32"
	"golang.org/x/sys/windows"
)

const InvalidHandle = ^w32.HANDLE(0)

//go:generate go run mksyscall_windows.go -output zsetupapi_windows.go setupapi_windows.go

//sys	setupDiClassGuidsFromNameEx(ClassName string, guid *Guid, size uint32, required_size *uint32, machineName string, reserved uint32) (err error) = setupapi.SetupDiClassGuidsFromNameExW
//sys	setupDiGetClassDevsEx(ClassGuid *Guid, Enumerator *string, hwndParent uintptr, Flags uint32, DeviceInfoSet uintptr, MachineName string, reserved uint32) (handle Handle, err error) = setupapi.SetupDiGetClassDevsExW
//sys	setupDiEnumDeviceInfo(DeviceInfoSet Handle, MemberIndex uint32, DeviceInfoData *spDeviceInformationData) (err error) = setupapi.SetupDiEnumDeviceInfo
//sys	setupDiGetDeviceInstanceId(DeviceInfoSet Handle, DeviceInfoData *spDeviceInformationData, DeviceInstanceId unsafe.Pointer, DeviceInstanceIdSize uint32, RequiredSize *uint32) (err error) = setupapi.SetupDiGetDeviceInstanceIdW

type SPDeviceInformationData struct {
	spDeviceInformationData
	devInfo w32.HDEVINFO
}

type spDeviceInformationData struct {
	cbSize    uint32
	ClassGuid w32.GUID
	DevInst   uint32
	reserved  uintptr
}

// SetupDiClassGuidsFromNameEx retrieves the GUIDs associated with the specified class name. This resulting list contains the classes currently installed on a local or remote computer.
func SetupDiClassGuidsFromNameEx(className string, machineName string) ([]w32.GUID, error) {
	requiredSize := uint32(0)
	err := setupDiClassGuidsFromNameEx(className, nil, 0, &requiredSize, machineName, 0)

	rets := make([]w32.GUID, requiredSize, requiredSize)
	err = setupDiClassGuidsFromNameEx(className, &rets[0], 1, &requiredSize, machineName, 0)
	return rets, err
}

// SetupDiEnumDeviceInfo returns a SP_DEVINFO_DATA structure that specifies a device information element in a device information set.
func EnumDeviceInfo(di w32.HDEVINFO, memberIndex uint32) (*SPDeviceInformationData, error) {
	did := spDeviceInformationData{}

	did.cbSize = uint32(unsafe.Sizeof(did))

	err := setupDiEnumDeviceInfo(w32.HANDLE(di), memberIndex, &did)
	return &SPDeviceInformationData{
		spDeviceInformationData: did,
		devInfo:                 di,
	}, err
}

// InstanceID retrieves the device instance ID that is associated with a device information element
func (did *SPDeviceInformationData) InstanceID() (string, error) {
	requiredSize := uint32(0)
	err := setupDiGetDeviceInstanceId(w32.HANDLE(did.devInfo), &did.spDeviceInformationData, nil, 0, &requiredSize)

	buff := make([]uint16, requiredSize)
	err = setupDiGetDeviceInstanceId(w32.HANDLE(did.devInfo), &did.spDeviceInformationData, unsafe.Pointer(&buff[0]), uint32(len(buff)), &requiredSize)
	if err != nil {
		return "", err
	}

	return windows.UTF16ToString(buff[:]), err
}

// SetupDiGetClassDevsEx returns a handle to a device information set that contains requested device information elements for a local or a remote computer.
func SetupDiGetClassDevsEx(ClassGuid w32.GUID, Enumerator string, hwndParent uintptr, Flags uint32, DeviceInfoSet uintptr, MachineName string, reserved uint32) (w32.HDEVINFO, error) {
	enumerator := &Enumerator

	if Enumerator == "" {
		enumerator = nil
	}

	hDevInfo, err := setupDiGetClassDevsEx(&ClassGuid, enumerator, hwndParent, uint32(Flags), DeviceInfoSet, MachineName, 0)
	return w32.HDEVINFO(hDevInfo), err
}
