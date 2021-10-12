package bluetooth

import (
	"errors"
	"fmt"
	"time"

	"github.com/godbus/dbus"
	"pkg.deepin.io/lib/dbusutil"
)

func (b *Bluetooth) ConnectDevice(device dbus.ObjectPath, apath dbus.ObjectPath) *dbus.Error {
	logger.Debugf("ConnectDevice %q %q", device, apath)
	b.setInitiativeConnect(device, true)
	err := b.sysBt.ConnectDevice(0, device, apath)
	return dbusutil.ToError(err)
}

func (b *Bluetooth) DisconnectDevice(device dbus.ObjectPath) *dbus.Error {
	logger.Debugf("DisconnectDevice %q", device)
	err := b.sysBt.DisconnectDevice(0, device)
	return dbusutil.ToError(err)
}

func (b *Bluetooth) RemoveDevice(adapter, device dbus.ObjectPath) *dbus.Error {
	logger.Debugf("RemoveDevice %q %q", adapter, device)
	err := b.sysBt.RemoveDevice(0, adapter, device)
	return dbusutil.ToError(err)
}

func (b *Bluetooth) SetDeviceAlias(device dbus.ObjectPath, alias string) *dbus.Error {
	logger.Debugf("SetDeviceAlias %q %q", device, alias)
	err := b.sysBt.SetDeviceAlias(0, device, alias)
	return dbusutil.ToError(err)
}

func (b *Bluetooth) SetDeviceTrusted(device dbus.ObjectPath, trusted bool) *dbus.Error {
	logger.Debugf("SetDeviceTrusted %q %v", device, trusted)
	err := b.sysBt.SetDeviceTrusted(0, device, trusted)
	return dbusutil.ToError(err)
}

// GetDevices return all device objects that marshaled by json.
func (b *Bluetooth) GetDevices(adapter dbus.ObjectPath) (devicesJSON string, busErr *dbus.Error) {
	logger.Debugf("GetDevices %q", adapter)
	devices := b.devices.getDevices(adapter)
	devicesJson := marshalJSON(devices)
	return devicesJson, nil
}

// GetAdapters return all adapter objects that marshaled by json.
func (b *Bluetooth) GetAdapters() (adaptersJSON string, busErr *dbus.Error) {
	logger.Debug("GetAdapters")
	return b.adapters.toJSON(), nil
}

func (b *Bluetooth) RequestDiscovery(adapter dbus.ObjectPath) *dbus.Error {
	logger.Debugf("RequestDiscovery %q", adapter)
	err := b.sysBt.RequestDiscovery(0, adapter)
	return dbusutil.ToError(err)
}

// SendFiles 用来发送文件给蓝牙设备，仅支持发送给已连接设备
func (b *Bluetooth) SendFiles(devAddress string, files []string) (sessionPath dbus.ObjectPath, busErr *dbus.Error) {
	if len(files) == 0 {
		return "", dbusutil.ToError(errors.New("files is empty"))
	}
	can, err := b.sysBt.CanSendFile(0)
	if err != nil {
		return "", dbusutil.ToError(err)
	}
	if !can {
		return "", dbusutil.ToError(errors.New("no permission"))
	}

	// 检查设备是否已经连接
	dev := b.getConnectedDeviceByAddress(devAddress)
	if dev == nil {
		logger.Debug("device is nil", dev)
		return "", dbusutil.ToError(errors.New("device not connected"))
	}

	sessionPath, err = b.sendFiles(dev, files)
	return sessionPath, dbusutil.ToError(err)
}

// CancelTransferSession 用来取消发送的会话，将会终止会话中所有的传送任务
func (b *Bluetooth) CancelTransferSession(sessionPath dbus.ObjectPath) *dbus.Error {
	//添加延时，确保sessionPath被remove，防止死锁
	time.Sleep(500 * time.Millisecond)
	b.sessionCancelChMapMu.Lock()
	defer b.sessionCancelChMapMu.Unlock()

	cancelCh, ok := b.sessionCancelChMap[sessionPath]
	if !ok {
		return dbusutil.ToError(errors.New("session not exists"))
	}

	cancelCh <- struct{}{}

	return nil
}

func (b *Bluetooth) SetAdapterPowered(adapter dbus.ObjectPath,
	powered bool) *dbus.Error {

	logger.Debugf("SetAdapterPowered %q %v", adapter, powered)
	err := b.sysBt.SetAdapterPowered(0, adapter, powered)
	return dbusutil.ToError(err)
}

func (b *Bluetooth) SetAdapterAlias(adapter dbus.ObjectPath, alias string) *dbus.Error {
	logger.Debugf("SetAdapterAlias %q %q", adapter, alias)
	err := b.sysBt.SetAdapterAlias(0, adapter, alias)
	return dbusutil.ToError(err)
}

func (b *Bluetooth) SetAdapterDiscoverable(adapter dbus.ObjectPath,
	discoverable bool) *dbus.Error {

	logger.Debugf("SetAdapterDiscoverable %q %v", adapter, discoverable)
	err := b.sysBt.SetAdapterDiscoverable(0, adapter, discoverable)
	return dbusutil.ToError(err)
}

func (b *Bluetooth) SetAdapterDiscovering(adapter dbus.ObjectPath,
	discovering bool) *dbus.Error {

	logger.Debugf("SetAdapterDiscovering %q %v", adapter, discovering)
	err := b.sysBt.SetAdapterDiscovering(0, adapter, discovering)
	return dbusutil.ToError(err)
}

func (b *Bluetooth) SetAdapterDiscoverableTimeout(adapter dbus.ObjectPath,
	discoverableTimeout uint32) *dbus.Error {
	logger.Debugf("SetAdapterDiscoverableTimeout %q %v", adapter, discoverableTimeout)
	err := b.sysBt.SetAdapterDiscoverableTimeout(0, adapter, discoverableTimeout)
	return dbusutil.ToError(err)
}

//Confirm should call when you receive RequestConfirmation signal
func (b *Bluetooth) Confirm(device dbus.ObjectPath, accept bool) *dbus.Error {
	logger.Infof("Confirm %q %v", device, accept)
	err := b.feed(device, accept, "")
	return dbusutil.ToError(err)
}

//FeedPinCode should call when you receive RequestPinCode signal, notice that accept must true
//if you accept connect request. If accept is false, pinCode will be ignored.
func (b *Bluetooth) FeedPinCode(device dbus.ObjectPath, accept bool, pinCode string) *dbus.Error {
	logger.Infof("FeedPinCode %q %v %q", device, accept, pinCode)
	err := b.feed(device, accept, pinCode)
	return dbusutil.ToError(err)
}

//FeedPasskey should call when you receive RequestPasskey signal, notice that accept must true
//if you accept connect request. If accept is false, passkey will be ignored.
//passkey must be range in 0~999999.
func (b *Bluetooth) FeedPasskey(device dbus.ObjectPath, accept bool, passkey uint32) *dbus.Error {
	logger.Infof("FeedPasskey %q %v %d", device, accept, passkey)
	err := b.feed(device, accept, fmt.Sprintf("%06d", passkey))
	return dbusutil.ToError(err)
}

func (b *Bluetooth) DebugInfo() (info string, busErr *dbus.Error) {
	logger.Debug("DebugInfo")
	info, err := b.sysBt.DebugInfo(0)
	return info, dbusutil.ToError(err)
}

//ClearUnpairedDevice will remove all device in unpaired list
func (b *Bluetooth) ClearUnpairedDevice() *dbus.Error {
	logger.Debug("ClearUnpairedDevice")
	err := b.sysBt.ClearUnpairedDevice(0)
	return dbusutil.ToError(err)
}

func (b *Bluetooth) CanSendFile() (can bool, busErr *dbus.Error) {
	logger.Debug("CanSendFile")
	can, err := b.sysBt.CanSendFile(0)
	return can, dbusutil.ToError(err)
}
