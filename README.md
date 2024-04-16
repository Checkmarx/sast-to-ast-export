<img src="https://raw.githubusercontent.com/Checkmarx/ci-cd-integrations/main/.images/banner.png">
<br />
<div  align="center" >

[![Documentation][documentation-shield]][documentation-url]
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]

</div>

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
    <a href="https://checkmarx.atlassian.net/wiki/spaces/AST/pages/6247580171/SAST+Migration+to+AST"><strong>Explore the docs »</strong></a>
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

Microsoft Windows x64.

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
* Please refer to [Wiki](https://checkmarx.atlassian.net/wiki/spaces/AST/pages/6247580171/SAST+Migration+to+AST) for more details

### Execution

Run export with:
```
.\cxsast_exporter --user username --pass password --url http://localhost
```

 * Replace `username` and `password` with user credentials.
 * Replace `http://localhost` with the url to SAST, if necessary.

## Additional Documentation

Refer to the project [Wiki](https://checkmarx.com/resource/documents/en/34965-68669-sast-cli-export-tool.html) for additional information

## Similarity Calculator

The exporter relies on a Windows binary for similarity calculation.
This is internally built by Checkmarx and provided in the `external` folder for inclusion with the build. 

## Contributing

We appreciate feedback and contribution to this repo! Before you get started, please see the following:

- [Checkmarx general contribution guidelines](CONTRIBUTING.md)
- [Checkmarx code of conduct guidelines](CODE-OF-CONDUCT.md)

## License
Distributed under the [Apache 2.0](LICENSE). See `LICENSE` for more information.

[documentation-shield]: https://img.shields.io/badge/docs-viewdocs-blue.svg
[documentation-url]:https://checkmarx.com/resource/documents/en/34965-68669-sast-cli-export-tool.html
[contributors-shield]: https://img.shields.io/github/contributors/Checkmarx/sast-to-ast-export.svg
[contributors-url]: https://github.com/Checkmarx/sast-to-ast-export/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/Checkmarx/sast-to-ast-export.svg
[forks-url]: https://github.com/Checkmarx/sast-to-ast-export/network/members
[stars-shield]: https://img.shields.io/github/stars/Checkmarx/sast-to-ast-export.svg
[stars-url]: https://github.com/Checkmarx/sast-to-ast-export/stargazers
[issues-shield]: https://img.shields.io/github/issues/Checkmarx/sast-to-ast-export.svg
[issues-url]: https://github.com/Checkmarx/sast-to-ast-export/issues
[license-shield]: https://img.shields.io/github/license/Checkmarx/sast-to-ast-export.svg
[license-url]: https://github.com/Checkmarx/sast-to-ast-export/blob/master/LICENSE