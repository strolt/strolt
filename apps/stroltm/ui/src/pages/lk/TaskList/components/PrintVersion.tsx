import { Space } from "antd";

import { LatestVersionLink } from "components";
import { observer } from "mobx-react-lite";
import { infoStore } from "stores/info.store";

export interface PrintVersionProps {
  instanceName?: string;
  proxyName?: string;
}
export const PrintVersion: React.FC<PrintVersionProps> = observer(({ instanceName, proxyName }) => {
  if (!instanceName) {
    return <>-</>;
  }

  const info = infoStore.map.get(infoStore.getKey(instanceName, proxyName));

  if (!info || !info.version) {
    return <>-</>;
  }

  return (
    <Space>
      {info.version}
      <LatestVersionLink label="Upgrade" version={info.version} />
    </Space>
  );
});
