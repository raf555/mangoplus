# Contributing

This API client is a minimal working version to fulfill author's current requirements, therefore there are missing endpoints and some data in the API response may not complete.

Contributions are welcome, including (but not limited to):

- New endpoint implementations.
- Completing partial response types.
- Updating the protobuf schema to match the latest API.
- Bug fixes and documentation improvements.
- Tests.
- Performance improvements (please provide the benchmark result to demonstrate the gain).

The protobuf schema stubs are available in the `internal/proto` package.

Reverse engineering the MangaPlus Android app as well as its protos is out of scope for this document; contributors adding new endpoints are expected to be familiar with the relevant tooling.

## Issues

- Bug reports: please include the request/response (with secrets redacted), the client version, and steps to reproduce.
- Missing endpoint requests are accepted, but a PR is more likely to land than a request.
