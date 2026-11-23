function getConfig(config) {
  const goToolchain = "go1.27.3";
  const taskApp = config.apps.task ?? {};

  return merge(config, {
    apps: {
      go: { binary: config.runtimes.go.managed },
      "golangci-lint": {
        dependsOn: [...(config.apps["golangci-lint"].dependsOn ?? []), "go"],
      },
      "openapi-generator": {
        jvm: {
          jarHash: "sha256:cb087e40001e31eb08ef6140dd5de10938dbeb89016a1fe0481eaa25cd569026",
          jarUrl:
            "https://github.com/OpenAPITools/openapi-generator/releases/download/v7.23.0/openapi-generator-cli-7.23.0.jar",
          version: "7.23.0",
        },
      },
      "swag-v1": {
        dependsOn: ["go"],
        go: {
          lockFile:
            "br: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",
          packageName: "github.com/swaggo/swag/cmd/swag",
          version: "v1.16.6",
        },
      },
      "swagger-v033": {
        binary: {
          binaries: {
            darwin: {
              amd64: {
                unknown: {
                  contentType: "binary",
                  hash: "sha256:96acc02b6fb96576fa25bbfec9ad2d7bc8c3c489e53cd7059d27f42b6ceeba2e",
                  url: "https://github.com/go-swagger/go-swagger/releases/download/v0.33.1/swagger_darwin_amd64",
                },
              },
              arm64: {
                unknown: {
                  contentType: "binary",
                  hash: "sha256:79dbe810c215e2c9595f82dfe933ee0c69318254aee0f4c13e4b4ad2df32e3b5",
                  url: "https://github.com/go-swagger/go-swagger/releases/download/v0.33.1/swagger_darwin_arm64",
                },
              },
            },
            linux: {
              amd64: {
                unknown: {
                  contentType: "binary",
                  hash: "sha256:c668a4a1f96f65fdc166fef3f89adebe15b82f34e9cdeb71bc771436f6b94b07",
                  url: "https://github.com/go-swagger/go-swagger/releases/download/v0.33.1/swagger_linux_amd64",
                },
              },
              arm64: {
                unknown: {
                  contentType: "binary",
                  hash: "sha256:d21f57d56f3f6dcae0e801b0c47067549343b35a5234e76bfc608cae411598aa",
                  url: "https://github.com/go-swagger/go-swagger/releases/download/v0.33.1/swagger_linux_arm64",
                },
              },
            },
          },
          version: "v0.33.1",
        },
        dependsOn: ["go"],
      },
      task: {
        ...taskApp,
        dependsOn: [
          ...(taskApp.dependsOn ?? []),
          { app: "go", as: "go" },
          { app: "swag-v1", as: "swag" },
          { app: "swagger-v033", as: "swagger" },
        ],
        runtimeEnv: {
          ...(taskApp.runtimeEnv ?? {}),
          SWAG_BIN: "${APP_BIN:swag}",
          SWAGGER_BIN: "${APP_BIN:swagger}",
        },
      },
    },
    ejectConfigs: [...(config.ejectConfigs ?? []), "gitleaks"],
    tools: {
      actionlint: {
        operations: {
          lint: {
            args: [
              "-config-file",
              "../actionlint.yaml",
              "-no-color",
              "-format",
              "{{json .}}",
              "{file}",
            ],
          },
        },
      },
      "golangci-lint": {
        operations: {
          fix: { env: { GOTOOLCHAIN: goToolchain } },
          lint: { env: { GOTOOLCHAIN: goToolchain } },
        },
      },
    },
  });
}

globalThis.getConfig = getConfig;
globalThis.getMinVersion = () => "0.6.2";
