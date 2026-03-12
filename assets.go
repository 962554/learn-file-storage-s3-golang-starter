package main

import (
	"os"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func createRandomBase() string {
	buf := make([]byte, 32)
	rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}

func createFileName(mediaType string) string {
	base := createRandomBase()
	ext := strings.Split(mediaType, "/")[1]

	return base + "." + ext
}

func (cfg apiConfig) createAsset(mediaType string) string {
	return filepath.Join(cfg.assetsRoot, createFileName(mediaType))
}

