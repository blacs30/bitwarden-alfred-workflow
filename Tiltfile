# https://docs.tilt.dev/api.html
local_resource('go mod download', trigger_mode=TRIGGER_MODE_AUTO, cmd='GOPROXY=https://proxy.golang.org,https://goproxy.io go mod download',
ignore=[
  '.gitignore',
  'LICENSE',
  'README.md',
  ])
local_resource('go lint', trigger_mode=TRIGGER_MODE_AUTO, cmd='GOPROXY=https://proxy.golang.org,https://goproxy.io go get ./...; golangci-lint run --timeout 3m',
ignore=[
  '.gitignore',
  'LICENSE',
  'README.md',
  ])
local_resource('go vet', trigger_mode=TRIGGER_MODE_AUTO, cmd='GOPROXY=https://proxy.golang.org,https://goproxy.io go vet',
ignore=[
  '.gitignore',
  'LICENSE',
  'README.md',
  ])
local_resource('go test', trigger_mode=TRIGGER_MODE_AUTO, cmd='GOPROXY=https://proxy.golang.org,https://goproxy.io go test -short',
ignore=[
  '.gitignore',
  'LICENSE',
  'README.md',
  ])
local_resource('Alfred build', trigger_mode=TRIGGER_MODE_MANUAL, cmd='alfred build',
ignore=[
  '.gitignore',
  'LICENSE',
  'README.md',
  ])
