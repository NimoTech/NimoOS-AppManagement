# NimoOS-AppManagement

> ### About
>
> NimoOS is a fork of [CasaOS](https://github.com/IceWhaleTech/CasaOS)
> (Apache-2.0), originally developed by IceWhale Technology Co., Ltd.
> Building on that foundation, NimoOS adds an AI agent, RAG-based
> retrieval, a knowledge layer, and a built-in web terminal.
>
> See [`NOTICE`](./NOTICE) for attribution details. CasaOS and IceWhale
> are trademarks of IceWhale Technology Co., Ltd. NimoOS is an independent
> project and is not affiliated with, endorsed by, or sponsored by
> IceWhale Technology Co., Ltd.

> ⚠️ Multi-user isolation is incomplete — Photos and Search are not yet
> per-user scoped. Read
> [SECURITY.md](https://github.com/NimoTech/NimoOS/blob/main/SECURITY.md#known-limitations)
> before deploying NimoOS for more than one person.

[![Go Reference](https://pkg.go.dev/badge/github.com/NimoTech/NimoOS-AppManagement.svg)](https://pkg.go.dev/github.com/NimoTech/NimoOS-AppManagement)
[![Go Report Card](https://goreportcard.com/badge/github.com/NimoTech/NimoOS-AppManagement)](https://goreportcard.com/report/github.com/NimoTech/NimoOS-AppManagement)
[![goreleaser](https://github.com/NimoTech/NimoOS-AppManagement/actions/workflows/release.yml/badge.svg)](https://github.com/NimoTech/NimoOS-AppManagement/actions/workflows/release.yml)
[![codecov](https://codecov.io/gh/NimoTech/NimoOS-AppManagement/branch/main/graph/badge.svg?token=ZCWZOFKXJT)](https://codecov.io/gh/NimoTech/NimoOS-AppManagement)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=NimoTech_NimoOS-AppManagement&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=NimoTech_NimoOS-AppManagement)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=NimoTech_NimoOS-AppManagement&metric=bugs)](https://sonarcloud.io/summary/new_code?id=NimoTech_NimoOS-AppManagement)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=NimoTech_NimoOS-AppManagement&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=NimoTech_NimoOS-AppManagement)
[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=NimoTech_NimoOS-AppManagement&metric=ncloc)](https://sonarcloud.io/summary/new_code?id=NimoTech_NimoOS-AppManagement)
[![Duplicated Lines (%)](https://sonarcloud.io/api/project_badges/measure?project=NimoTech_NimoOS-AppManagement&metric=duplicated_lines_density)](https://sonarcloud.io/summary/new_code?id=NimoTech_NimoOS-AppManagement)

App management service manages NimoOS apps lifecycle, such as installation, running, etc.
