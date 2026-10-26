function getConfig(config) {
  const goToolchain = "go1.26.4";

  return merge(config, {
    apps: {
      go: { binary: config.runtimes.go.managed },
      "golangci-lint": {
        dependsOn: [...(config.apps["golangci-lint"].dependsOn ?? []), "go"],
      },
    },
    ejectConfigs: [...(config.ejectConfigs ?? []), "gitleaks"],
    tools: {
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
globalThis.getMinVersion = () => "0.5.3";
