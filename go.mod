module device-camera-go

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/atagirov/onvif4go v0.3.0
	github.com/beevik/etree v1.1.0
	github.com/brianolson/cbor_go v1.0.0 // indirect
	github.com/edgexfoundry/device-sdk-go v0.0.0-20190522161733-b70462a1276d
	github.com/edgexfoundry/go-mod-core-contracts v0.1.0
	github.com/elgs/gostrgen v0.0.0-20161222160715-9d61ae07eeae // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/ugorji/go v1.1.4
)

replace (
	github.com/atagirov/onvif4go => ../onvif4go
	github.com/edgexfoundry/device-sdk-go v0.0.0-20190522161733-b70462a1276d => ../device-sdk-go
	github.com/satori/go.uuid v1.2.0 => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
)
