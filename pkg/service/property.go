package service

import (
	"os"

	"next-terminal/pkg/guacd"
	"next-terminal/server/model"
	"next-terminal/server/repository"
	"next-terminal/server/utils"
)

type ConfigsService struct {
	configsRepository *repository.ConfigsRepository
}

func NewConfigsService(configsRepository *repository.ConfigsRepository) *ConfigsService {
	return &ConfigsService{configsRepository: configsRepository}
}

func (r ConfigsService) InitConfigs() error {
	propertyMap := r.configsRepository.FindAllMap()

	if len(propertyMap[guacd.Host]) == 0 {
		property := model.Configs{
			Ckey: guacd.Host,
			Cval: utils.GetKeyFromYaml("terminal.guacd.host", "127.0.0.1"),
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.Port]) == 0 {
		property := model.Configs{
			Ckey: guacd.Port,
			Cval: utils.GetKeyFromYaml("terminal.guacd.port", "4822"),
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.EnableRecording]) == 0 {
		property := model.Configs{
			Ckey: guacd.EnableRecording,
			Cval: "true",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.RecordingPath]) == 0 {
		path, _ := os.Getwd()
		property := model.Configs{
			Ckey: guacd.RecordingPath,
			Cval: path + "/recording/",
		}
		if !utils.FileExists(property.Cval) {
			if err := os.Mkdir(property.Cval, os.ModePerm); err != nil {
				return err
			}
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.CreateRecordingPath]) == 0 {
		property := model.Configs{
			Ckey: guacd.CreateRecordingPath,
			Cval: "true",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.DriveName]) == 0 {
		property := model.Configs{
			Ckey: guacd.DriveName,
			Cval: "File-System",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.DrivePath]) == 0 {

		path, _ := os.Getwd()

		property := model.Configs{
			Ckey: guacd.DrivePath,
			Cval: path + "/drive/",
		}
		if !utils.FileExists(property.Cval) {
			if err := os.Mkdir(property.Cval, os.ModePerm); err != nil {
				return err
			}
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.FontName]) == 0 {
		property := model.Configs{
			Ckey: guacd.FontName,
			Cval: "menlo",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.FontSize]) == 0 {
		property := model.Configs{
			Ckey: guacd.FontSize,
			Cval: "12",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.ColorScheme]) == 0 {
		property := model.Configs{
			Ckey: guacd.ColorScheme,
			Cval: "gray-black",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.EnableDrive]) == 0 {
		property := model.Configs{
			Ckey: guacd.EnableDrive,
			Cval: "true",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.EnableWallpaper]) == 0 {
		property := model.Configs{
			Ckey: guacd.EnableWallpaper,
			Cval: "false",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.EnableTheming]) == 0 {
		property := model.Configs{
			Ckey: guacd.EnableTheming,
			Cval: "false",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.EnableFontSmoothing]) == 0 {
		property := model.Configs{
			Ckey: guacd.EnableFontSmoothing,
			Cval: "false",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.EnableFullWindowDrag]) == 0 {
		property := model.Configs{
			Ckey: guacd.EnableFullWindowDrag,
			Cval: "false",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.EnableDesktopComposition]) == 0 {
		property := model.Configs{
			Ckey: guacd.EnableDesktopComposition,
			Cval: "false",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.EnableMenuAnimations]) == 0 {
		property := model.Configs{
			Ckey: guacd.EnableMenuAnimations,
			Cval: "false",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.DisableBitmapCaching]) == 0 {
		property := model.Configs{
			Ckey: guacd.DisableBitmapCaching,
			Cval: "false",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.DisableOffscreenCaching]) == 0 {
		property := model.Configs{
			Ckey: guacd.DisableOffscreenCaching,
			Cval: "false",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}

	if len(propertyMap[guacd.DisableGlyphCaching]) == 0 {
		property := model.Configs{
			Ckey: guacd.DisableGlyphCaching,
			Cval: "true",
		}
		if err := r.configsRepository.Create(&property); err != nil {
			return err
		}
	}
	return nil
}
