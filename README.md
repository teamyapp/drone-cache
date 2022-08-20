# Drone Cache
The most flexible cache plugin for Drone CI

## Getting Started

### Frontend project

```yaml
- name: restore cache
    image: ghcr.io/teamyapp/drone-cache:0.0.9
    settings:
      s3_endpoint: sfo3.digitaloceanspaces.com
      s3_access_key_id:
        from_secret: SPACE_ACCESS_KEY
      s3_secret:
        from_secret: SPACE_SECRET
      s3_bucket: teamyapp
      remote_root_dir: cache/node/teamy-web
      restore: true
      cacheable_relative_paths:
        - node_modules
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
  - name: refresh cache
    image: ghcr.io/teamyapp/drone-cache:0.0.9
    settings:
      s3_endpoint: sfo3.digitaloceanspaces.com
      s3_access_key_id:
        from_secret: SPACE_ACCESS_KEY
      s3_secret:
        from_secret: SPACE_SECRET
      s3_bucket: teamyapp
      remote_root_dir: cache/node/teamy-web
      refresh: true
      cacheable_relative_paths:
        - node_modules
```

### Go project
```yaml
steps:
  - name: restore cache
    image: ghcr.io/teamyapp/drone-cache:0.0.9
    volumes:
      - name: cache
        path: /go/pkg/mod
    settings:
      s3_endpoint: sfo3.digitaloceanspaces.com
      s3_access_key_id:
        from_secret: SPACE_ACCESS_KEY
      s3_secret:
        from_secret: SPACE_SECRET
      s3_bucket: teamyapp
      remote_root_dir: cache/go
      restore: true
      cacheable_absolute_paths:
        - /go/pkg/mod
  - name: run unit tests
    image: golang:1.18
    volumes:
      - name: cache
        path: /go/pkg/mod
    commands:
      - go test ./...
  - name: refresh cache
    image: ghcr.io/teamyapp/drone-cache:0.0.9
    volumes:
      - name: cache
        path: /go/pkg/mod
    settings:
      s3_endpoint: sfo3.digitaloceanspaces.com
      s3_access_key_id:
        from_secret: SPACE_ACCESS_KEY
      s3_secret:
        from_secret: SPACE_SECRET
      s3_bucket: teamyapp
      remote_root_dir: cache/go
      refresh: true
      cacheable_absolute_paths:
        - /go/pkg/mod
volumes:
  - name: cache
    temp: {}
```

## Available settings

| Setting                  | Data Type | Description                                       |
|--------------------------|-----------|---------------------------------------------------|
| debug                    | bool      | print the value for all settings                  |
| s3_endpoint              | string    | endpoint for s3 compatible storage                |
| s3_access_key_id         | string    | access key id for s3 compatible storage           |
| s3_secret                | string    | secret of access key for s3 compatible storage    |
| s3_bucket                | string    | bucket name for s3 compatible storage             |
| remote_root_dir          | string    | cache file root directory in the s3 bucket        |
| restore                  | bool      | restore cached directories during this build step |
| refresh                  | bool      | refresh cached during this build step             |
| cacheable_relative_paths | string    | cached paths, path relative to repo root          |
| cacheable_absolute_paths | string    | cached paths, absolute paths, volume not needed   |

## Publish new image
```
./scripts/publish.sh [version]
```

## License
MIT