# Flag

**Flag** is a configuration-based feature flag manager. Works on top of Git, Flag automatically adapts files in your repository based on which features are turned on or off.

> [!WARNING]
> Disclaimer: Use with caution in production environments.

## Table of Contents

-   [Overview](#overview)
-   [Key Features](#key-features)
-   [Blocks](#blocks)
-   [Delimiters](#delimiters)
-   [Versions](#versions)
-   [Commands](#commands)
-   [Getting Started](#getting-started)

---

## Overview

Feature flagging has become a critical tool for controlled feature deployment, but **Flag** takes this to the next level by allowing feature toggling at the **file level**. Instead of toggling functionality purely in code, Flag modifies actual file content based on active features.

By leveraging Git, Flag ensures that changes are versioned, tracked, and consistent across your codebase. Features can be managed dynamically through **blocks**, **delimiters**, and **versions**, with seamless Git integration to track every change.

---

## Key Features

-   **File-based Feature Toggling**: Modify file content dynamically based on active features.
-   **Git Integration**: Works only within Git repositories for version control and tracking.
-   **Blocks, Delimiters, and Versions**: Three core concepts for managing features in code and configuration files.
-   **Branch-based**: Feature flags are isolated at the branch level, allowing for granular control.

---

## Blocks

**Blocks** are the fundamental units in Flag, allowing you to define sections of code that can be toggled on or off based on active features.

### Structure of a Block:

```plaintext
// @feature(myFeature) //
<feature_content>
// @default(myFeature) //
<default_content>
// !feature //
```

-   <feature_content>: This content is visible when the feature is enabled.
-   <default_content>: This content is used when the feature is disabled.
-   //: The delimiters used to enclose block definitions.

The system reads and modifies the content based on the feature's status. When you toggle a feature on, the `<feature_content>` is inserted into the file. When it's off, the `<default_content>` is used.

Use always the sync ommand to keep your features updated

---

## Delimeters

Delimiters are used to define the boundaries for blocks. They let the system recognize where feature-specific content starts and ends in a file.

You can manage delimiters with the following commands:

-   Set a delimiter: `flag delimeters set <file_extension> <delimeter_start> <delimeter_end>`
-   List delimiters: `flag delimeters list`
-   Delete a delimiter: `flag delimeters delete <file_extension>`

These operations let you fully control how Flag identifies and processes blocks in your files.

---

## Versions

In cases where block delimiters are not allowed (such as JSON files, which do not support comments), Flag uses Versions to manage features.

### Version Workflow:

1. Create a base version: `flag versions base`
   This creates a base reference of the files to start tracking features.
2. Sync features

## States vs. Features

-   Features: Regular feature toggles.
-   States: A combination of multiple features, e.g., feature1+feature2.

Flag also supports operations like updating specific features, creating new states, and deleting features.

# Commands

```
NAME:
   flag - flag is a branch-level feature flag manager

USAGE:
   flag [global options] command [command options]

VERSION:
   v0.0.1

AUTHOR:
   costaluu

COMMANDS:
   init        creates a new workspace
   sync        updates all features on created, modified, deleted files
   report      shows a workspace report of features
   delimeters  operations for delimiters
   blocks      operations for block features
   versions     operations for version-based features
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

---

# Getting started

1. Install Flag: Download the last release for your OS.
2. Initialize Workspace: Run flag init to create a new workspace in your Git repository.
3. Define Feature Blocks: Use the @feature and @default block structure in your files.
4. Manage Delimiters: Set up delimiters using flag delimeters set.
5. Handle Files Without Comments: Use versions-based tracking with flag versions base and flag sync.
