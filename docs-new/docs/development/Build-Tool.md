---
layout: default
title: Build Tool
parent: Development
---

# Build Tool
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

There is a useful bash script you can use to build and deploy the operator easier.

Invoke it from the root of the repository directory
 
```
./build.sh
```

to get a basic help message.

## Commands

```
Help:
Apicurio Registry Operator build tool
Note: Run this script from the root dir of the project.

./build.sh [command] [parameters]...

Commands: 
  build
  help
  mkdeploy
  mkundeploy
  push

Parameters:
  -r|--repository [repository] Operator image repository
  -n|--namespace [namespace] Namespace where the operator is deployed
  --cr [file] Path to a file with 'ApicurioRegistry' custom resource to be deployed
  --nocr Do not deploy default 'ApicurioRegistry' custom resource
  --crname [name] Name of the 'ApicurioRegistry' custom resource (e.g. for mkundeploy), default is 'example-apicurioregistry'
  --latest Also push the image with the 'latest' tag
```
