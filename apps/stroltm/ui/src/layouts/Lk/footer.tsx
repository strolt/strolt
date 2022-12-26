import { LatestVersionLink } from "components";

import { useStores, observer } from "stores";

export const Footer = observer(() => {
  const { infoStore } = useStores();

  return (
    <footer style={{ display: "flex", justifyContent: "center", padding: "1rem" }}>
      <div>
        version: {infoStore.version} <LatestVersionLink version={infoStore.version} />
      </div>
    </footer>
  );
});
