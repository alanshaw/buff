package app

type StorageConfig struct {
	DataDir string
	// Service-specific storage configurations
	Delegation DelegationStorageConfig
}

type DelegationStorageConfig struct {
	Dir string
}
