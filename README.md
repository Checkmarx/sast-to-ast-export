# How to run build

1. Create an admin user in SAST
2. Run export
```
    ./cxsast_exporter --user username --pass password --url http://localhost
```

This will generate an export package in the same folder where the command is run.

# How to run repo

In order to run during development we'll need to:
1. Get public key from KMS and save in public.key file
2. Run with build time variable setting

Example for dev environment:
```
$env:SAST_EXPORT_KMS_KEY_ID="cb3052be-1e3a-4a9c-b3f0-84d963c53a06"
aws kms get-public-key --key-id $env:SAST_EXPORT_KMS_KEY_ID | jq -r .PublicKey > public.key
go run -ldflags "-X sast-export/internal.buildTimeRSAPublicKey=$(cat .\public.key)" . --user <admin_username> --pass <admin_password> --url http://localhost
```

Once the public.key file is created we can run and test using the file reference:
```
go run -ldflags "-X sast-export/internal.buildTimeRSAPublicKey=$(cat .\public.key)" .
```

```
go test -ldflags "-X sast-export/internal.buildTimeRSAPublicKey=$(cat .\public.key)" .\...
```

# How to build

1. Set KMS key id environment variable
2. Make public key file
3. Build

Example for dev environment:
```
$env:SAST_EXPORT_KMS_KEY_ID="cb3052be-1e3a-4a9c-b3f0-84d963c53a06"
make public_key
make build
```
