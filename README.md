# git-file-sync

A tool to sync specific files from a Git repository to local filesystem.

## Features

- Fetch specific files from a Git repository using commit hash
- Support for SSH authentication with deploy keys
- YAML configuration file

## Requirements

- Git client installed
- SSH client installed
- Read access to target Git repository

## Configuration

Create a `config.yaml` file:

```yaml
git:
  repoUrl: "git@github.com:your-org/your-repo.git"
  commitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0"

files:
  # Single file
  - source: "configs/production/app-config.json"
    destination: "/app/configs/app-config.json"
  
  # Another single file
  - source: "assets/logo.svg"
    destination: "/app/static/logo.svg"
  
  # Entire directory
  - source: "scripts/"
    destination: "/app/scripts/"

sshPrivateKeyPath: "/secrets/deploy_key"
```

`destination` is the final path where the source is placed. When `source` is a directory, its contents are written directly under `destination` and the source directory name is not appended. A trailing slash does not change the behavior.

## Usage

```bash
# Use default config.yaml in current directory
git-file-sync

# Specify custom config file
git-file-sync -config=/path/to/config.yaml

# Enable verbose logging
git-file-sync -verbose
```

## Use Cases

### Container Init Container

Use as an init container to provide configuration files to your main application:

- Deploy configuration files before application startup
- Update configuration without rebuilding application images
- Share files between containers via volumes

### CI/CD Pipeline

Fetch specific files during build or deployment:

- Pull build scripts from a central repository
- Retrieve environment-specific configurations
- Sync deployment manifests

### Configuration Management

Manage application configurations across environments:

- Use commit hashes for versioned configuration deployment
- Rollback to previous configurations by changing commit hash
- Audit configuration changes through Git history

### Static Asset Distribution

Distribute static files to multiple locations:

- Sync documentation files
- Deploy static website assets
- Distribute shared libraries or scripts

## Testing

1. Create a test repository with sample files
2. Generate an SSH deploy key:
   ```bash
   $ ssh-keygen -t ed25519 -f deploy_key -N ""
   ```
3. Add the public key to your repository as a read-only deploy key
4. Create a test config file:
   ```yaml
   git:
     repoUrl: "git@github.com:your-org/test-repo.git"
     commitHash: "your-commit-hash"
   files:
     - source: "test.txt"
       destination: "/tmp/test.txt"
   sshPrivateKeyPath: "./deploy_key"
   ```
5. Run the syncer:
   ```bash
   $ git-file-sync -config=test-config.yaml -verbose
   ```

## License

This project is licensed under the [MIT License](./LICENSE).
