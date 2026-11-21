#!/bin/sh
set -eu

ROOT=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
SWAG_BIN=${SWAG_BIN:-swag}
SWAGGER_BIN=${SWAGGER_BIN:-swagger}
GO_PATH="${GOROOT:?Run codegen through the Taskfile managed Go environment}/bin:$PATH"

if [ "$#" -eq 0 ]; then
  set -- strolt stroltp stroltm
fi

for app in "$@"; do
  case "$app" in
    strolt | stroltp | stroltm) ;;
    *)
      printf 'unsupported schema: %s\n' "$app" >&2
      exit 2
      ;;
  esac
done

for app in "$@"; do
  rm -rf "$ROOT/.swagger/$app"
  case "$app" in
    strolt)
      rm -rf "$ROOT/shared/sdk/strolt/generated/strolt_client" "$ROOT/shared/sdk/strolt/generated/strolt_models"
      ;;
    stroltp)
      rm -rf "$ROOT/shared/sdk/stroltp/generated/stroltp_client" "$ROOT/shared/sdk/stroltp/generated/stroltp_models"
      ;;
  esac
done

for app in "$@"; do
  mkdir -p "$ROOT/.swagger/$app"
  (cd "$ROOT/apps/$app" && PATH="$GO_PATH" "$SWAG_BIN" init -g ./internal/api/api.go -q --parseDependency --output "$ROOT/.swagger/$app")
  case "$app" in
    strolt | stroltp)
      if [ "$app" = strolt ]; then
        client=strolt
      else
        client=stroltp
      fi
      (cd "$ROOT/shared/sdk/$client/generated" && PATH="$GO_PATH" "$SWAGGER_BIN" generate client -q -c "${client}Client" -m "${client}Models" -f "$ROOT/.swagger/$app/swagger.yaml" --template-dir="$ROOT/.go-swagger/templates" --allow-template-override)
      ;;
  esac
done
