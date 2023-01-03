# Drone Cache

The most flexible cache plugin for Drone CI

## Getting Started

### Node.js project

```yaml
steps:
  - name: retrieve cache
    image: ghcr.io/teamyapp/drone-cache:0.1.9
    volumes:
      - name: cache
        path: /var/lib/cache
    settings:
      mode: retrieve
      version_file_path: yarn.lock
      cacheable_relative_paths:
        - node_modules
      storage_type: volume
      volume_cache_root_dir: /var/lib/cache
  - name: build frontend
    image: node:16.13.0-alpine3.13
    commands:
      - apk add --no-cache git g++ make python3
      - yarn install --frozen-lockfile
      - yarn build:staging
  - name: run unit tests
    image: node:16.13.0-alpine3.13
    commands:
      - apk add --no-cache git
      - CI=true yarn test
  - name: persist cache
    image: ghcr.io/teamyapp/drone-cache:0.1.9
    volumes:
      - name: cache
        path: /var/lib/cache
    settings:
      mode: persist
      version_file_path: yarn.lock
      cacheable_relative_paths:
        - node_modules
      storage_type: volume
      volume_cache_root_dir: /var/lib/cache
volumes:
  - name: cache
    host:
      path: /var/lib/cache
```

### Go project

```yaml
steps:
  - name: retrieve cache
    image: ghcr.io/teamyapp/drone-cache:0.1.9
    volumes:
      - name: go-mod-cache
        path: /go/pkg/mod
      - name: cache
        path: /var/lib/cache
    settings:
      mode: retrieve
      version_file_path: go.sum
      cacheable_absolute_paths:
        - /go/pkg/mod
      storage_type: volume
      volume_cache_root_dir: /var/lib/cache
  - name: run unit tests
    image: golang:1.18
    volumes:
      - name: go-mod-cache
        path: /go/pkg/mod
    commands:
      - go test ./...
  - name: persist cache
    image: ghcr.io/teamyapp/drone-cache:0.1.9
    volumes:
      - name: go-mod-cache
        path: /go/pkg/mod
      - name: cache
        path: /var/lib/cache
    settings:
      mode: persist
      version_file_path: go.sum
      cacheable_absolute_paths:
        - /go/pkg/mod
      storage_type: volume
      volume_cache_root_dir: /var/lib/cache
volumes:
  - name: go-mod-cache
    temp: { }
  - name: cache
    host:
      path: /var/lib/cache
```

## Available settings

| Setting                  | Data Type | Description                                    |
|--------------------------|-----------|------------------------------------------------|
| debug                    | bool      | print the value for all settings               |
| mode                     | string    | retrieve/persist                               | 
| version_file_path        | string    | use sha256 hash of this file as cache key      | 
| cacheable_relative_paths | string    | cached paths, path relative to repo root       |
| cacheable_absolute_paths | string    | cached paths, absolute paths                   |
| storage_type             | string    | volume/s3                                      |
| s3_endpoint              | string    | endpoint for s3 compatible storage             |
| s3_access_key_id         | string    | access key id for s3 compatible storage        |
| s3_secret                | string    | secret of access key for s3 compatible storage |
| s3_bucket                | string    | bucket name for s3 compatible storage          |
| s3_cache_root_dir        | string    | cache root directory in the s3 bucket          |
| volume_cache_root_dir    | string    | cache root directory for the attached volume   |

## Publish new image
```
./scripts/publish.sh [version]
```

## License

MIT