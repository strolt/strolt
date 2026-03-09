import type { FC } from "react";

import { observer, useStores } from "stores";

export interface LatestVersionLinkProps {
  label?: string;
  version?: string;
}
export const LatestVersionLink: FC<LatestVersionLinkProps> = observer(({ label, version }) => {
  const { infoStore } = useStores();

  if (infoStore.latestVersion == version) {
    return null;
  }

  return (
    <a href="#" rel="noopener noreferrer" target="_blank">
      {label || "new version"} ({infoStore.latestVersion})
    </a>
  );
});
