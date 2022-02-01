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
</p>

# SAST to AST Export

Exports SAST triaged results for importing in AST.

## Description

Fetches SAST triaged results and exports as an encrypted package, which can then be imported in AST.

## Getting Started

### Dependencies

Requires Microsoft Windows x64.

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

Refer to the project [Wiki](https://checkmarx.atlassian.net/wiki/spaces/AST/pages/6247580171/SAST+Migration+to+AST) for additional information

## Encryption keys

The export package produced is encrypted using a public RSA keys which is embedded in the binary produced with the build.
Keys for each environment are provided in the `keys` folder and are rotated every year.

## Similarity Calculator

The exporter relies on a Windows binary for similarity calculation.
This is internally built by Checkmarx and provided in the `external` folder for inclusion with the build. 

## Version History

 * 1.0
   * Initial Release

## Contributing

We appreciate feedback and contribution to this repo! Before you get started, please see the following:

- [Checkmarx general contribution guidelines](CONTRIBUTING.md)
- [Checkmarx code of conduct guidelines](CODE-OF-CONDUCT.md)

## Support + Feedback

Include information on how to get support. Consider adding:

- Use [Issues](https://github.com/CheckmarxDev/ast-sast-export/issues) for code-level support

## License

Project License can be found [here](LICENSE)
