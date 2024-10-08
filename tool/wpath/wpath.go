package wpath

import (
	"os"
	"path"
)

// GetHomePath get or nil
func GetHomePath() string {
	dir, _ := os.UserHomeDir()
	return dir
}

func JoinHomePath(paths ...string) string {
	return path.Join(GetHomePath(), path.Join(paths...))
}

// GetUserConfigPath get or nil
func GetUserConfigPath() string {
	dir, _ := os.UserConfigDir()
	return dir
}

func JoinUserConfigPath(paths ...string) string {
	return path.Join(GetUserConfigPath(), path.Join(paths...))
}

// GetUserCachePath get or nil
func GetUserCachePath() string {
	dir, _ := os.UserCacheDir()
	return dir
}

func JoinUserCachePath(paths ...string) string {
	return path.Join(GetUserCachePath(), path.Join(paths...))
}

func GetTempPath() string {
	return os.TempDir()
}

func JoinTempPath(paths ...string) string {
	return path.Join(GetTempPath(), path.Join(paths...))
}
