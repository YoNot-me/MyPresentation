package adminService

import (
	"fmt"
	"os"
	"presentator/internal/core/entity"
)

func (a *serviceHelper) validateNotNil(toValidate []string) error {
	for _, v := range toValidate {
		if v == "" || v == " " {
			return entity.BadRequest
		}
	}

	return nil
}

func (a *serviceHelper) validateStatus(status string) error {

	switch status {
	case "в работе", "сдан", "на согласовании", "правка":
		return nil
	default:
		return entity.BadRequest
	}
}

func (s *serviceHelper) renameFolders(brandName, workName, newWorkName string) error {

	if newWorkName != "" {

		err := os.Rename(
			fmt.Sprintf("%s/%s/%s", dst, brandName, workName),
			fmt.Sprintf("%s/%s/%s", dst, brandName, newWorkName),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
