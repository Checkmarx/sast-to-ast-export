# How to run

1. Create an admin user in SAST
2. Run export
```
    ./cxsast_exporter --user username --pass password --url http://localhost
```

This will generate an export package in the same folder where the command is run.

add --debug parameter to bypass the zip and encryption process. 

# How to build

Make command:
```
make build [ENV=<environment>]
```

Available environments:
 * prod (default)
 * ppe
 * dev

Example prod build: `make build`

Example dev build: `make build ENV=dev`

## Package for distribution

Note generating zip packages requires zip command in your path.

Example command on Linux
```
make package
```

On Windows you can use WSL:
```
wsl make package
```

## Build similarity calculator

In order to calculate AST similarity IDs, the export tool relies on a .NET CLI app, available in external folder.  

1. Checkout https://github.com/CheckmarxDev/ast-sast-similarity-calculator
2. Open solution with visual studio 2019
3. Right-click on the solution to open context menu
4. Click on "Publish..."
5. Make sure you have a Folder publish profile:
   1. Target location: {{ast-sast-export folder}}/external/windows/amd64
   2. Configuration: Release
   3. Target framework: netcoreapp3.1
   4. Target runtime: win-x64
6. Click on "Publish" button

## Troubleshooting

### No such file or directory

Examples:

```
> make build
process_begin: CreateProcess(NULL, cat VERSION, ...) failed.
Makefile:28: pipe: No such file or directory
```

This happens because the shell being spawn doesn't support some commands needed.
On Windows, please make sure you have Git bash installed, and add Git bash's usr/bin folder to your PATH.
Git bash is `C:\Program Files\Git\usr\bin` and should be added in your user's `Path` variable.
