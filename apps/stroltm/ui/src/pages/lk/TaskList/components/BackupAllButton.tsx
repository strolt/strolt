import { useEffect } from "react";

import { Button, message, Popconfirm } from "antd";

import { observer } from "mobx-react-lite";
import { useStores } from "stores";

export const BackupAllButton = observer(() => {
  const { managerStore } = useStores();

  useEffect(() => {
    return () => managerStore.resetBackupAll();
  }, []);

  const handleClick = async () => {
    const data = await managerStore.backupAll();
    message.info(`Success started: ${data.successStarted?.length}`);
    message.error(
      `Error started: ${data.errorStarted?.length} [${data.errorStarted
        ?.map((el) => [el.instanceName, el.serviceName, el.taskName].filter(Boolean).join(" "))
        .filter(Boolean)
        .join(", ")}]`,
    );
  };

  return (
    <Popconfirm okText="Yes" onConfirm={handleClick} title="Are you sure?">
      <Button
        block
        danger
        loading={managerStore.backupAllStatus?.state === "pending"}
        style={{ marginBottom: "1rem" }}
        type="primary"
      >
        Backup ALL (without filters)
      </Button>
    </Popconfirm>
  );
});
