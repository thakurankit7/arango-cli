# Arango-CLI

An intuitive CLI tool for managing ArangoDB.

## About The Project

This project provides a command-line interface for managing ArangoDB databases, collections, and documents. It's built with Go and the Cobra library to provide a user-friendly experience. It also includes a configuration management system that allows you to easily switch between different ArangoDB environments.

## Getting Started

To get a local copy up and running follow these simple steps.

### Prerequisites

* Go (version 1.15 or later)

### Installation

1. Clone the repo
   ```sh
   git clone https://github.com/your_username/arango-cli.git
   ```
2. Build the project
   ```sh
   go build -o arango-cli
   ```

## Configuration

The Arango CLI is configured using a `env.yaml` file located in the `config` directory. This file allows you to define multiple ArangoDB environments and easily switch between them.

### Example `env.yaml`

```yaml
databases:
  development:
    host: localhost
    port: 8529
    username: root
    password: ""
    database: "_system"
    ssl: false
  production:
    host: your-production-host
    port: 8529
    username: your-production-user
    password: your-production-password
    database: "_system"
    ssl: true
default: development
```

### Switching Between Environments

You can easily switch between your configured environments using the `/switch` command:

```sh
/switch <environment-name>
```

## Usage

The Arango CLI provides several commands to interact with your ArangoDB instance.

### Available Commands

Once you're in the interactive shell, you can use the following commands:

* `/show databases` or `/db`: List all available databases.
* `/show collections` or `/col`: List all collections in the current database.
* `/use <database-name>`: Switch to a different database.
* `/list configs` or `/configs`: List all available configurations from your `env.yaml` file.
* `/switch <config-name>`: Switch to a different ArangoDB connection configuration.
* `/current`: Show the current connection details.
* `exit` or `quit`: Exit the interactive shell.
* `help`: Show help.

## Contributing

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Distributed under the MIT License.