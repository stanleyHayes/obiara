package config

import "strings"

// ObjectStorage addresses the bucket that holds member media.
//
// Every field is read from the environment with no default. A default bucket
// would be somebody else's bucket, and a default credential is a credential
// in source control; the media context is simply left uncomposed when this is
// absent, which is what happens in local development and in tests.
type ObjectStorage struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	PathStyle bool
}

// Configured reports whether enough is present to compose the media context.
// The credential pair is checked together: half a credential is a
// misconfiguration that should read as "off", not as "on and broken".
func (storage ObjectStorage) Configured() bool {
	return storage.Region != "" && storage.Bucket != "" &&
		storage.AccessKey != "" && storage.SecretKey != ""
}

func loadObjectStorage(getenv func(string) string) ObjectStorage {
	return ObjectStorage{
		Endpoint:  strings.TrimSpace(getenv("OBJECT_STORAGE_ENDPOINT")),
		Region:    strings.TrimSpace(getenv("OBJECT_STORAGE_REGION")),
		Bucket:    strings.TrimSpace(getenv("OBJECT_STORAGE_BUCKET")),
		AccessKey: strings.TrimSpace(getenv("OBJECT_STORAGE_ACCESS_KEY")),
		SecretKey: strings.TrimSpace(getenv("OBJECT_STORAGE_SECRET_KEY")),
		// R2 and MinIO require path-style addressing; AWS S3 accepts either.
		PathStyle: strings.EqualFold(strings.TrimSpace(getenv("OBJECT_STORAGE_PATH_STYLE")), "true"),
	}
}
