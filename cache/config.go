package cache

type storageType string

const (
	s3StorageType     storageType = "s3"
	volumeStorageType storageType = "volume"
)

type Config struct {
	RepoName               string   `envconfig:"DRONE_REPO" default:""`
	BranchName             string   `envconfig:"DRONE_BRANCH" default:""`
	Debug                  bool     `envconfig:"PLUGIN_DEBUG" default:"false"`
	Mode                   string   `envconfig:"PLUGIN_MODE" default:"retrieve"`
	VersionFilePath        string   `envconfig:"PLUGIN_VERSION_FILE_PATH" default:""`
	CacheableRelativePaths []string `envconfig:"PLUGIN_CACHEABLE_RELATIVE_PATHS" default:""`
	CacheableAbsolutePaths []string `envconfig:"PLUGIN_CACHEABLE_ABSOLUTE_PATHS" default:""`
	StorageType            string   `envconfig:"PLUGIN_STORAGE_TYPE" default:"volume"`
	S3Endpoint             string   `envconfig:"PLUGIN_S3_ENDPOINT"`
	S3AccessKeyID          string   `envconfig:"PLUGIN_S3_ACCESS_KEY_ID"`
	S3Secret               string   `envconfig:"PLUGIN_S3_SECRET"`
	S3Bucket               string   `envconfig:"PLUGIN_S3_BUCKET"`
	S3CacheRootDir         string   `envconfig:"PLUGIN_S3_CACHE_ROOT_DIR"`
	VolumeCacheRootDir     string   `envconfig:"PLUGIN_VOLUME_CACHE_ROOT_DIR"`
}
