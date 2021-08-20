# How to run build

1. Create an admin user in SAST
2. Run export
```
    ./cxsast_exporter --user username --pass password --url http://localhost
```

This will generate an export package in the same folder where the command is run.

# How to build

1. Make sure you have access to AWS KMS
2. Set KMS key id environment variable
3. Make public key file
4. Build

Example for dev environment:
```
$env:SAST_EXPORT_KMS_KEY_ID="cb3052be-1e3a-4a9c-b3f0-84d963c53a06"
make public_key
make build
```

# How to run repo

In order to run during development we'll need to `make public_key`, like for build.
Once the public.key file exists, we can run and test using the file reference:
```
go run -ldflags "-X sast-export/internal.buildTimeRSAPublicKey=$(cat .\public.key)" .
```

```
go test -ldflags "-X sast-export/internal.buildTimeRSAPublicKey=$(cat .\public.key)" .\...
```
