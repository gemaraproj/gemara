# Gemara: GRC Engineering Model for Automated Risk Assessment  [![Go Reference](https://pkg.go.dev/badge/github.com/gemaraproj/gemara.svg)](https://pkg.go.dev/github.com/gemaraproj/gemara)

> Pronounced: Juh-MAH-ruh (think :gem:)

**Gemara** is a standardized, machine-readable data model designed to bridge the gap between high-level compliance requirements and low-level technical evidence. By providing a structured schema (powered by [CUE](https://cuelang.org/)), Gemara enables automated risk assessment, consistent reporting, and interoperability across the security toolchain.

## Resources

1. View the model and supporting resources at [gemara.openssf.org](https://gemara.openssf.org)
2. Find schemas in this repository, or in the CUE central registry.
  - Use the schemas directly with [cue](https://cuelang.org/) for validating Gemara data payloads against the schemas and more.
3. Use the Go SDK to integrate Gemara schemas into your automated tools
  - `github.com/gemaraproj/go-gemara` and consult our [go docs](https://pkg.go.dev/github.com/gemaraproj/go-gemara)


## Projects and tooling using Gemara

Some Gemara use cases include:

- [FINOS Common Cloud Controls](https://www.finos.org/common-cloud-controls-project) (Layer 2)
- [Open Source Project Security Baseline](https://baseline.openssf.org/) (Layer 2)
- [Privateer](https://github.com/privateerproj/privateer) (Layer 5)
  - ex. [OSPS Baseline Privateer Plugin](https://github.com/revanite-io/pvtr-github-repo)

## Contributing

We're so glad you asked - see [CONTRIBUTING.md](/CONTRIBUTING.md) and if you have any questions or feedback head over to the OpenSSF Slack in [#gemara](https://openssf.slack.com/archives/C09A9PP765Q)

You can also join the biweekly meeting on alternate Thursdays.  
See Gemara Bi-Weekly Meeting on the [OpenSSF calendar](https://calendar.google.com/calendar/u/0?cid=czYzdm9lZmhwNWk5cGZsdGI1cTY3bmdwZXNAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ) for details.