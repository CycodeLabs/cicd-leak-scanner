# CICD Leak Scanner

CICD Leak Scanner is a security tool designed to scan build logs from CI/CD pipelines to identify leaks of sensitive data, tokens, or credentials.

Due to the recent malicious actions involving [tj-actions/changed-files](https://cycode.com/blog/github-action-tj-actions-changed-files-supply-chain-attack-the-complete-guide/), we've created this tool to empower security teams to proactively scan and determine if their CI/CD pipelines have been compromised.

## Features

* Sensitive Data Detection: Quickly identifies leaked secrets, tokens, and credentials in build logs.
* Configurable Rules: Customize the scanner to detect specific patterns or keywords.
* Reporting: Generates a detailed report with the list of leaks found in the build logs.


## Installation

``` bash
git clone https://github.com/CycodeLabs/cicd-leak-scanner.git
cd cicd-leak-scanner
go build -o cicd-leak-scanner .
```


## Usage

Scan public build logs:

``` bash
./cicd-leak-scanner -t $GITHUB_TOKEN
```

Scan Specific Organization:

``` bash
./cicd-leak-scanner -t $GITHUB_TOKEN -o organization
```

Scan Specific Repository:

``` bash
./cicd-leak-scanner -t $GITHUB_TOKEN  -r orgName/repoName
```


## Configuration

The scanner uses a configuration file to define the rules for detecting leaks. The configuration file is located at `config.yaml` and can be customized to detect specific patterns or keywords.

``` yaml
output:
  method: file
  filename: output.json

rules:
  - name: Detect base64 secrets leaked by tj-actions/changed-files
    query: tj-actions/changed-files language:yaml path:.github/workflows
    regex: >
      ##\[group\]changed-files\s*\r?\n\d{4}-\d{2}-\d{2}T[\d:.]+Z\s+([A-Za-z0-9+/=]+)
    decoders:
      - id: base64_decode
        repeat: 2
```

## Contributing

Contributions are welcome! Please open a pull request or an issue if you'd like to suggest improvements or report bugs.

## License

[Apache License 2.0](https://github.com/CycodeLabs/cicd-leak-scanner/blob/main/LICENSE.md)