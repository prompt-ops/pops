package common

type CloudConnection interface {
	CreateConnection(name string) error
	CheckAuthentication() error
	PrintInitialContext() error
	MainLoop() error
	Type() string
	SubType() string
	CommandType() string
}

var AvailableCloudConnectionTypes = []string{
	"azure",
}

func GetAvailableCloudConnectionTypes() []string {
	return AvailableCloudConnectionTypes
}
