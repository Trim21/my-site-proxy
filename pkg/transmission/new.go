package transmission

import (
	"fmt"
	"net/url"
	"os"

	"github.com/hekmon/transmissionrpc/v2"

	"app/pkg/errgo"
	"app/pkg/utils"
)

func New() (*transmissionrpc.Client, error) {
	entryPoint, found := os.LookupEnv("TRANSMISSION_API_ENTRYPOINT")
	if !found {
		return nil, nil
	}

	u, err := url.Parse(entryPoint)
	if err != nil {
		return nil, errgo.Wrap(err, fmt.Sprintf("TRANSMISSION_API_ENTRYPOINT '%s' is not valid url", entryPoint))
	}

	username, password := utils.GetUserPass(u.User)
	port := utils.GetPort(u)

	client, err := transmissionrpc.New(u.Hostname(), username, password, &transmissionrpc.AdvancedConfig{
		HTTPS: u.Scheme == "https",
		Port:  port,
	})
	if err != nil {
		return nil, errgo.Wrap(err, "failed to create transmission client")
	}

	return client, nil
}
