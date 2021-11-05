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
2. Make sure that the KMS key exists with the alias sast-migration-key in the eu-west-1 region
3. Make public key file
4. Build

Example for dev environment:
```
make public_key
make build
```

## Troubleshooting

### No such file or directory

Examples:

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
go run -ldflags "-X sast-export/internal/encryption.BuildTimeRSAPublicKey=$(cat .\public.key)" .
```

```
go test -ldflags "-X sast-export/internal/encryption.BuildTimeRSAPublicKey=$(cat .\public.key)" .\...
```

