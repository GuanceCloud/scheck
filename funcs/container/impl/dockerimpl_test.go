package impl

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/golang/mock/gomock"
)

// nolint
func setImages() (images []types.ImageSummary) {
	for i := 0; i < 5; i++ {
		image := types.ImageSummary{
			ID:   "case01",
			Size: 1024,
		}
		images = append(images, image)
	}
	return
}

// nolint
func setContainers() (cons []types.Container) {
	for i := 0; i < 5; i++ {
		image := types.Container{
			ID:     "case01",
			Status: "ok",
		}
		cons = append(cons, image)
	}
	return
}

func TestGetImageList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIndex := NewMockcliMock(ctrl)
	images := setImages()
	mockIndex.EXPECT().ImageList(context.Background(), types.ImageListOptions{}).Return(images, nil) // literals work
	if _, err := GetImageList(mockIndex); err != nil {
		t.Logf("TestGetImageList err=%v", err)
	}
}

func TestGetContainerList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIndex := NewMockcliMock(ctrl)
	images := setContainers()
	mockIndex.EXPECT().ContainerList(context.Background(), types.ContainerListOptions{}).Return(images, nil) // literals work
	if _, err := GetContainerList(mockIndex); err != nil {
		t.Logf("TestGetContainerList err=%v", err)
	}
}
