import { FC } from "react";

import { observer, useStores } from "stores";

export interface LatestVersionLinkProps {
  version?: string;
}
export const LatestVersionLink: FC<LatestVersionLinkProps> = observer(({ version }) => {
  const { infoStore } = useStores();

  if (infoStore.latestVersion == version) {
    return null;
  }

  return (
    <a href="#" target="_blank" rel="noopener noreferrer">
      new version
    </a>
  );
});
