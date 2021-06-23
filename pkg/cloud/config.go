package cloud

type Config struct {
	BindAddress string
	Version string `json:"-"`
}
