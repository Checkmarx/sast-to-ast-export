<!-- PROJECT LOGO -->
<br />
<p align="center">
  <a href="">
    <img src="./logo.png" alt="Logo" width="80" height="80">
  </a>

<h3 align="center">SAST to AST Export</h3>

<p align="center">
    SAST to AST Export is a standalone Checkmarx tool.
<br />
    <a href="https://docs.checkmarx.com/en/34965-68666-migrating-from-sast-to-checkmarx-one.html"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/Checkmarx/sast-to-ast-export/issues/new/choose">Report Bug</a>
    ·
    <a href="https://github.com/Checkmarx/sast-to-ast-export/issues/new/choose">Request Feature</a>
</p>

# SAST to AST Export

Exports SAST triaged results for importing in AST.

## Description

Fetches SAST triaged results and exports as an encrypted package, which can then be imported in AST.

## Getting Started

### Prerequisites

Microsoft Windows x64

SAST v9.3 or higher.

### Installation

* Download the latest version and extract the package contents
* Create export user in SAST
  * Assign the following permissions:
    1. Sast > API > Use Odata
    2. Sast > Reports > Generate Scan Report
    3. Sast > Scan Results > View Results
    4. Access Control > General > Manage Authentication Providers
    5. Access Control > General > Manage Roles
* Please refer to [Wiki](https://docs.checkmarx.com/en/34965-68669-sast-cli-export-tool.html) for more details

### Execution

Run export with:
```
.\cxsast_exporter --user username --pass password --url http://localhost
```

 * Replace `username` and `password` with user credentials.
 * Replace `http://localhost` with the url to SAST, if necessary.
* Please refer to [Wiki](https://docs.checkmarx.com/en/34965-68670-cxsast_exporter.html) for more details

## Additional Documentation

Importing SAST to Checkmarx One [Wiki](https://docs.checkmarx.com/en/34965-68672-importing-sast-to-checkmarx-one.html)

Internal information [Wiki](https://checkmarx.com/resource/documents/en/34965-68669-sast-cli-export-tool.html)

## Similarity Calculator

The exporter relies on a Windows binary for similarity calculation.
This is internally built by Checkmarx and provided in the `external` folder for inclusion with the build. 

## Contributing

We appreciate feedback and contribution to this repo! Before you get started, please see the following:

- [Checkmarx general contribution guidelines](CONTRIBUTING.md)
- [Checkmarx code of conduct guidelines](CODE-OF-CONDUCT.md)

## License
Distributed under the [Apache 2.0](LICENSE). See `LICENSE` for more information.
