import { Button, Popconfirm, message } from "antd";
import { observer } from "mobx-react-lite";
import { useEffect } from "react";
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
    <Popconfirm title="Are you sure?" onConfirm={handleClick} okText="Yes">
      <Button
			block
        type="primary"
        style={{ marginBottom: "1rem" }}
        loading={managerStore.backupAllStatus?.state === "pending"}
        danger
      >
        Backup ALL (without filters)
      </Button>
    </Popconfirm>
  );
});
