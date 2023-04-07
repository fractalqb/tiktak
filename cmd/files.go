package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
	"gopkg.in/yaml.v3"
)

const (
	EnvTiktakData = "TIKTAK_DATA"
	DataFileExt   = ".tiktak"
)

type Config struct{}

func (*Config) DataFile(t time.Time) string {
	y, m, _ := t.Date()
	n := fmt.Sprintf("%04d-%02d%s", y, m, DataFileExt)
	return TikTakFile(n)
}

func OutputBasename(tl tiktak.TimeLine, day bool) string {
	switch len(tl) {
	case 0:
		return "empty" + DataFileExt
	case 1:
		y, m, d := tl[0].When().Date()
		if day {
			return fmt.Sprintf("%04d-%02d-%02d%s", y, m, d, DataFileExt)
		} else {
			return fmt.Sprintf("%04d-%02d%s", y, m, DataFileExt)
		}
	}
	sd := tiktak.DateOf(tl[0].When())
	ed := tiktak.DateOf(tl[len(tl)-1].When())
	if sd.Year == ed.Year && sd.Month == ed.Month {
		if day {
			if sd.Day == ed.Day {
				return fmt.Sprintf("%04d-%02d-%02d%s",
					sd.Year,
					sd.Month,
					sd.Day,
					DataFileExt,
				)
			}
			return fmt.Sprintf("%04d-%02d-%02d_%02d%s",
				sd.Year,
				sd.Month,
				sd.Day,
				ed.Day,
				DataFileExt,
			)
		}
		return fmt.Sprintf("%04d-%02d%s",
			sd.Year,
			sd.Month,
			DataFileExt,
		)
	}
	if day {
		return fmt.Sprintf("%04d-%02d-%02d_%04d-%02d-%02d%s",
			sd.Year,
			sd.Month,
			sd.Day,
			ed.Year,
			ed.Month,
			ed.Day,
			DataFileExt,
		)
	}
	return fmt.Sprintf("%04d-%02d_%04d-%02d%s",
		sd.Year,
		sd.Month,
		ed.Year,
		ed.Month,
		DataFileExt,
	)
}

func TikTakFile(base string) string {
	return filepath.Join(TikTakDir(), base)
}

func TikTakDir() string {
	dir := os.Getenv(EnvTiktakData)
	if dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	switch runtime.GOOS {
	case "windows":
		dir = filepath.Join(home, "AppData/Roaming/fqb/tiktak")
		os.MkdirAll(dir, 0777)
	default:
		home = filepath.Join(home, ".local/share")
		if s, err := os.Stat(home); err == nil && s.IsDir() {
			home = filepath.Join(home, "fqb/tiktak")
			os.MkdirAll(home, 0777)
			dir = home
		}
	}
	if dir == "" {
		dir = "."
	}
	return dir
}

func ReadConfig(cfg any) error {
	cfgFile := TikTakFile("tiktak.yaml")
	if s, err := os.Stat(cfgFile); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	} else if s.IsDir() {
		return fmt.Errorf("directory shadows config file %s", cfgFile)
	}
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, cfg)
	return err
}
