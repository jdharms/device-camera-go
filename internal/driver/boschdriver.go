package driver

import (
	"fmt"
	"github.com/edgexfoundry/device-sdk-go"
	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"strconv"
	"strings"
	"sync"
)

var once sync.Once
var lock sync.Mutex

var onvifClients map[string]*OnvifClient

var boschClients map[string]*RcpClient
var stopChans map[string]chan bool
var stoppedChans map[string]chan bool

var driver *Driver

type Driver struct {
	lc               logger.LoggingClient
	asynchCh         chan<- *sdkModel.AsyncValues
	config           *configuration
}

func NewProtocolDriver() sdkModel.ProtocolDriver {
	once.Do(func() {
		driver = new(Driver)
		onvifClients = make(map[string]*OnvifClient)
		boschClients = make(map[string]*RcpClient)
		stopChans = make(map[string]chan bool)
		stoppedChans = make(map[string]chan bool)
	})

	return driver
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	var responses = make([]*sdkModel.CommandValue, len(reqs))
	if _, ok := protocols["HTTP"]; !ok {
		d.lc.Error("No HTTP address found for device. Check configuration file.")
		return responses, fmt.Errorf("No HTTP address in protocols map")
	}

	if _, ok := protocols["HTTP"]["Address"]; !ok {
		d.lc.Error("No HTTP address found for device. Check configuration file.")
		return responses, fmt.Errorf("No HTTP address in protocols map")
	}

	addr := protocols["HTTP"]["Address"]

	// check for existence of both clients

	onvifClient, ok := getOnvifClient(addr)

	if !ok {
		dev, err := device.RunningService().GetDeviceByName(deviceName)
		if err != nil {
			err = fmt.Errorf("Device not found: %s", deviceName)
			d.lc.Error(err.Error())

			return responses, err
		}

		onvifClient = initializeOnvifClient(dev, d.config.Camera.User, d.config.Camera.Password)
	}

	boschClient, ok := getBoschClient(addr)

	if !ok {
		dev, err := device.RunningService().GetDeviceByName(deviceName)
		if err != nil {
			err = fmt.Errorf("Device not found: %s", deviceName)
			d.lc.Error(err.Error())

			return responses, err
		}
		boschClient = initializeBoschClient(dev, d.config.Camera.User, d.config.Camera.Password)
	}

	for i, req := range reqs {
		var result string
		switch req.DeviceResourceName {
		case "onvif_device_information":
			data, err := onvifClient.GetDeviceInformation()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			result = mapToString(data)

			cv := sdkModel.NewStringValue(reqs[i].DeviceResourceName, 0, string(result))
			responses[i] = cv
		case "onvif_profile_information":
			data, err := onvifClient.GetProfileInformation()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			profiles := make([]string, 0)
			for _, e := range data {
				profiles = append(profiles, mapToString(e))
			}

			result = strings.Join(profiles, ",,")

			cv := sdkModel.NewStringValue(reqs[i].DeviceResourceName, 0, string(result))
			responses[i] = cv
		case "motion_detected": fallthrough
		case "tamper_detected":
			alarmType, err := strconv.Atoi(req.Attributes["alarm_type"])
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}
			data := boschClient.GetAlarmState(alarmType)

			cv, err := sdkModel.NewBoolValue(reqs[i].DeviceResourceName, 0, data)
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}
			responses[i] = cv
		case "occupancy": fallthrough
		case "counter":
			counterType := req.Attributes["counter_name"]
			data := boschClient.GetCounterState(counterType)

			cv, err := sdkModel.NewUint32Value(reqs[i].DeviceResourceName, 0, uint32(data))
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}
			responses[i] = cv
		}
	}

	return responses, nil
}



// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource (aka DeviceObject).
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	return nil
}

// DisconnectDevice handles protocol-specific cleanup when a device
// is removed.
func (d *Driver) DisconnectDevice(deviceName string, protocols map[string]contract.ProtocolProperties) error {
	errString := "No HTTP address found for device. Check configuration file."
	if _, ok := protocols["HTTP"]; !ok {
		d.lc.Error(errString)
		return fmt.Errorf(errString)
	}

	if _, ok := protocols["HTTP"]["Address"]; !ok {
		d.lc.Error(errString)
		return fmt.Errorf(errString)
	}

	addr := protocols["HTTP"]["Address"]

	shutdownBoschClient(addr)
	shutdownOnvifClient(addr)
	return nil
}

// Initialize performs protocol-specific initialization for the device
// service.
func (d *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues) error {
	d.lc = lc
	d.asynchCh = asyncCh

	config, err := LoadConfigFromFile()
	if err != nil {
		panic(fmt.Errorf("read bosch driver configuration from file failed: %d", err))
	}
	d.config = config

	for _, dev := range device.RunningService().Devices() {
		initializeBoschClient(dev, config.Camera.User, config.Camera.Password)
		initializeOnvifClient(dev, config.Camera.User, config.Camera.Password)
	}

	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (d *Driver) Stop(force bool) error {
	for addr, ch := range stopChans {
		close(ch)
		if !force {
			<-stoppedChans[addr]
		}
	}

	close(d.asynchCh)

	return nil
}

func getOnvifClient(addr string) (*OnvifClient, bool) {
	lock.Lock()
	client, ok := onvifClients[addr]
	lock.Unlock()
	return client, ok
}

func getBoschClient(addr string) (*RcpClient, bool) {
	lock.Lock()
	client, ok := boschClients[addr]
	lock.Unlock()
	return client, ok
}

func initializeOnvifClient(device contract.Device, user string, password string) *OnvifClient {
	addr := device.Protocols["HTTP"]["Address"]
	client := NewOnvifClient(addr, user, password, driver.lc)
	lock.Lock()
	onvifClients[addr] = client
	lock.Unlock()
	return client
}

func initializeBoschClient(device contract.Device, user string, password string) *RcpClient {
	addr := device.Protocols["HTTP"]["Address"]

	client := NewRcpClient(driver.asynchCh, driver.lc)
	stopChan, stoppedChan := client.RcpCameraInit(device, addr, user, password)

	lock.Lock()
	boschClients[addr] = client
	stopChans[addr] = stopChan
	stoppedChans[addr] = stoppedChan
	lock.Unlock()

	return client
}

func shutdownOnvifClient(addr string) {
	// nothing much to do here at the moment
	lock.Lock()
	delete(onvifClients, addr)
	lock.Unlock()
}

func shutdownBoschClient(addr string) {
	lock.Lock()

	close(stopChans[addr])
	<-stoppedChans[addr]

	delete(stopChans, addr)
	delete(stoppedChans, addr)
	delete(boschClients, addr)

	lock.Unlock()
}

func mapToString(m map[string]string) string {
	var pairs []string
	for k, v := range m {
		pairs = append(pairs, fmt.Sprintf("%s:%s", k, v))
	}

	result := strings.Join(pairs, ",")
	return result
}