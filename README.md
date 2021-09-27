# How to run build

1. Create an admin user in SAST
2. Run export
```
    ./cxsast_exporter --user username --pass password --url http://localhost
```

This will generate an export package in the same folder where the command is run.

add --debug parameter to bypass the zip and encryption process. 

# How to build

1. Make sure you have access to AWS KMS
2. Set KMS key id environment variable
3. Make public key file
4. Build

Example for dev environment:
```
make public_key -e SAST_EXPORT_KMS_KEY_ID="cb3052be-1e3a-4a9c-b3f0-84d963c53a06"
make build
```

## Troubleshooting

### -z was unexpected at this time

Examples:
```
> make public_key
if [ -z cb3052be-1e3a-4a9c-b3f0-84d963c53a06 ]; then echo "Please specify env var SAST_EXPORT_KMS_KEY_ID"; exit 1; fi
-z was unexpected at this time.
make: *** [Makefile:43: public_key] Error 255
```

```
> make build
process_begin: CreateProcess(NULL, cat VERSION, ...) failed.
Makefile:28: pipe: No such file or directory
```

This happens because the shell being spawn doesn't support some commands needed.
Please make sure you have Git bash installed, and add Git bash's usr/bin folder to your PATH.
For example, in Windows this is `C:\Program Files\Git\usr\bin` and should be added in your user's `Path` variable.

### make public_key fails because "jq" is missing

Command `jq` is being used to parse JSON. You can find installation instructions in https://stedolan.github.io/jq/. 

# How to run repo

In order to run during development we'll need to `make public_key`, like for build.
Once the public.key file exists, we can run and test using the file reference:
```
go run -ldflags "-X sast-export/internal.buildTimeRSAPublicKey=$(cat .\public.key)" .
```

```
go test -ldflags "-X sast-export/internal.buildTimeRSAPublicKey=$(cat .\public.key)" .\...
```
