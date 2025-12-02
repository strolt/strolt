import { observer } from "mobx-react-lite";

import { Button, Popconfirm } from "antd";

import { useStores } from "stores";

export interface BackupButtonProps {
  proxyName?: string;
  instanceName?: string;
  serviceName?: string;
  taskName?: string;
  isDisabled: boolean;
}
export const BackupButton: React.FC<BackupButtonProps> = observer(
  ({ isDisabled, proxyName, instanceName, serviceName, taskName }) => {
    const { managerStore } = useStores();

    const status = managerStore.backupStatusMap.get(
      managerStore.backupStatusMapKey(
        instanceName || "",
        serviceName || "",
        taskName || "",
        proxyName,
      ),
    );

    const isLoading =
      status?.state === "pending" ||
      managerStore.taskStatusMap.get(
        managerStore.getTaskStatusMapKey(
          instanceName || "",
          serviceName || "",
          taskName || "",
          proxyName,
        ),
      )?.isRunning;

    const disabled = isDisabled || !instanceName || !serviceName || !taskName;

    return (
      <Popconfirm
        title="Are you sure?"
        onConfirm={() =>
          managerStore.backup(instanceName || "", serviceName || "", taskName || "", proxyName)
        }
        disabled={isLoading || disabled}
        okText="Yes"
      >
        <Button block loading={isLoading} disabled={disabled} size="small" danger>
          Backup
        </Button>
      </Popconfirm>
    );
  },
);
