import { observer } from "mobx-react-lite";

import { Space } from "antd";

import { LatestVersionLink } from "components";

import { infoStore } from "stores/info.store";

export interface PrintVersionProps {
  proxyName?: string;
  instanceName?: string;
}
export const PrintVersion: React.FC<PrintVersionProps> = observer(({ proxyName, instanceName }) => {
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
      <LatestVersionLink version={info.version} label="Upgrade" />
    </Space>
  );
});
