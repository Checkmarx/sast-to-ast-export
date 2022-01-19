# SAST to AST export

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
* Please refer to (TODO) [Wiki](insert-wiki-url) for more details

### Execution

Run export with:
```
.\cxsast_exporter --user username --pass password --url http://localhost
```

 * Replace `username` and `password` with user credentials.
 * Replace `http://localhost` with the url to SAST, if necessary.

## Additional Documentation

Refer to the project (TODO) [Wiki](insert-wiki-url) for additional information

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
