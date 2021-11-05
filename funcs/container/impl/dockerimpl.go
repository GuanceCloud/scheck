// Package impl  funcs export to docker
package impl

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

/*
cliMock :mock interface
	## how to mock func:
	> mockgen -destination dockerimpl_mock.go -package impl -source dockerimpl.go
*/
type cliMock interface {
	ImageList(ctx context.Context, opt types.ImageListOptions) (list []types.ImageSummary, err error)
	ContainerList(ctx context.Context, opt types.ContainerListOptions) (list []types.Container, err error)
}

func GetCli() (cli *client.Client, err error) {
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	return
}

// GetImageList : cli is mock interface, and replace struct is (client.Client).
func GetImageList(cli cliMock) (list []types.ImageSummary, err error) {
	ctx := context.Background()
	list, err = cli.ImageList(ctx, types.ImageListOptions{})
	return list, err
}

func GetContainerList(cli cliMock) (list []types.Container, err error) {
	ctx := context.Background()
	list, err = cli.ContainerList(ctx, types.ContainerListOptions{})
	return list, err
}
