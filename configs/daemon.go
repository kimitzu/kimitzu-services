package configs

type Daemon struct {
	ArgLogLevel string
	DataPath    string
	LogLevel    int
	Version     string

	DialTo                string
	ApiListen             string
	KeyPath               string
	GenerateNewKeys       bool
	ShowHelp              bool
	DatabasePath          string
	BootstrapNodeIdentity string
	Testnet				  bool
}
