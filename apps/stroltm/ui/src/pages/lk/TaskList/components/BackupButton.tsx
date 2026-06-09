import { Button, Popconfirm } from "antd";

import { observer } from "mobx-react-lite";
import { useStores } from "stores";

export interface BackupButtonProps {
  instanceName?: string;
  isDisabled: boolean;
  proxyName?: string;
  serviceName?: string;
  taskName?: string;
}
export const BackupButton: React.FC<BackupButtonProps> = observer(
  ({ instanceName, isDisabled, proxyName, serviceName, taskName }) => {
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
        disabled={isLoading || disabled}
        okText="Yes"
        onConfirm={() =>
          managerStore.backup(instanceName || "", serviceName || "", taskName || "", proxyName)
        }
        title="Are you sure?"
      >
        <Button block danger disabled={disabled} loading={isLoading} size="small">
          Backup
        </Button>
      </Popconfirm>
    );
  },
);
