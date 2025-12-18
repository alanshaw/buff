package app

// AppConfig is the root configuration for the entire application
type AppConfig struct {
	Identity IdentityConfig
	Storage  StorageConfig
}
